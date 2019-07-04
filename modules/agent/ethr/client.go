package ethr

import (
	"encoding/gob"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"io"
	"net"
	"os"
	"os/signal"
	"sort"
	"sync/atomic"
	"time"
)

var gIgnoreCert bool

//这里是要进行时间间隔
func runClient(testParam EthrTestParam, clientParam ethrClientParam, server string) {
	for {
		server = "[" + server + "]"
		test, err := establishSession(testParam, server)
		if err != nil {
			log.Info("%v", err)
			return
		}
		runTest(test, clientParam.duration)
		//todo 发送消息
		mvs := model.MetricValue{
			g.IP(),
			"network.test",
			float64(test.testResult.data) / clientParam.duration.Seconds(),
			600,
			"GAUGE",
			g.IP() + "." + server,
			time.Now().Unix()}
		var metrics []*model.MetricValue
		metrics = append(metrics, &mvs)
		g.SendToTransfer(metrics)
		//todo 这里要释放内存了，因为一般测试时命令行是直接退出的。而这里不是
		deleteTest(test)
		time.Sleep(g.COLLECT_INTERVAL)
	}
}

func establishSession(testParam EthrTestParam, server string) (test *ethrTest, err error) {
	conn, err := net.Dial(tcp(ipVer), server+":"+ctrlPort)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	ethrMsg := createSynMsg(testParam)
	err = sendSessionMsg(enc, ethrMsg)
	if err != nil {
		return
	}
	rserver, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	server = "[" + rserver + "]"
	test, err = newTest(server, conn, testParam, enc, dec)
	if err != nil {
		ethrMsg = createFinMsg(err.Error())
		sendSessionMsg(enc, ethrMsg)
		return
	}
	ethrMsg = recvSessionMsg(test.dec)
	if ethrMsg.Type != EthrAck {
		if ethrMsg.Type == EthrFin {
			err = fmt.Errorf("%s", ethrMsg.Fin.Message)
		} else {
			err = fmt.Errorf("Unexpected control message received. %v", ethrMsg)
		}
		deleteTest(test)
		return nil, err
	}
	gCert = ethrMsg.Ack.Cert
	napDuration := ethrMsg.Ack.NapDuration
	time.Sleep(napDuration)
	return
}

const (
	timeout    = 0
	interrupt  = 1
	serverDone = 2
)

func handleCtrlC(toStop chan int) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigChan
		switch sig {
		case os.Interrupt:
			fallthrough
		case os.Kill:
			toStop <- interrupt
		}
	}()
}

func runDurationTimer(d time.Duration, toStop chan int) {
	go func() {
		dSeconds := uint64(d.Seconds())
		if dSeconds == 0 {
			return
		}
		time.Sleep(d)
		toStop <- timeout
	}()
}

func runTest(test *ethrTest, d time.Duration) {
	startStatsTimer()
	if test.testParam.TestID.Protocol == TCP {
		if test.testParam.TestID.Type == Bandwidth {
			go runTCPBandwidthTest(test)
		} else if test.testParam.TestID.Type == Cps {
			go runTCPCpsTest(test)
		} else if test.testParam.TestID.Type == Latency {
			go runTCPLatencyTest(test)
		}
	}
	test.isActive = true
	toStop := make(chan int, 1)
	runDurationTimer(d, toStop)
	handleCtrlC(toStop)
	reason := <-toStop
	close(test.done)
	sendSessionMsg(test.enc, &EthrMsg{})
	stopStatsTimer()
	switch reason {
	case timeout:
		log.Info("Ethr done, duration: " + d.String() + ".")
	case interrupt:
		log.Info("Ethr done, received interrupt signal.")
	case serverDone:
		log.Info("Ethr done, server terminated the session.")
	}
}

func runTCPBandwidthTest(test *ethrTest) {
	server := test.session.remoteAddr
	log.Info("Connecting to host %s, port %s", server, tcpBandwidthPort)
	for th := uint32(0); th < test.testParam.NumThreads; th++ {
		buff := make([]byte, test.testParam.BufferSize)
		for i := uint32(0); i < test.testParam.BufferSize; i++ {
			buff[i] = byte(i)
		}
		go func() {
			conn, err := net.Dial(tcp(ipVer), server+":"+tcpBandwidthPort)
			if err != nil {
				log.Error("%v", err)
				os.Exit(1)
				return
			}
			defer conn.Close()
			ec := test.newConn(conn)
			rserver, rport, _ := net.SplitHostPort(conn.RemoteAddr().String())
			lserver, lport, _ := net.SplitHostPort(conn.LocalAddr().String())
			log.Info("[%3d] local %s port %s connected to %s port %s",
				ec.fd, lserver, lport, rserver, rport)
			blen := len(buff)
		ExitForLoop:
			for {
				select {
				case <-test.done:
					break ExitForLoop
				default:
					n := 0
					if test.testParam.Reverse {
						n, err = io.ReadFull(conn, buff)
					} else {
						n, err = conn.Write(buff)
					}
					if err != nil || n < blen {
						log.Debug("Error sending/receiving data on a connection for bandwidth test: %v", err)
						continue
					}
					atomic.AddUint64(&ec.data, uint64(blen))
					atomic.AddUint64(&test.testResult.data, uint64(blen))
				}
			}
		}()
	}
}

func runTCPCpsTest(test *ethrTest) {
	server := test.session.remoteAddr
	for th := uint32(0); th < test.testParam.NumThreads; th++ {
		go func() {
		ExitForLoop:
			for {
				select {
				case <-test.done:
					break ExitForLoop
				default:
					conn, err := net.Dial(tcp(ipVer), server+":"+tcpCpsPort)
					if err == nil {
						atomic.AddUint64(&test.testResult.data, 1)
						tcpconn, ok := conn.(*net.TCPConn)
						if ok {
							tcpconn.SetLinger(0)
						}
						conn.Close()
					}
				}
			}
		}()
	}
}

func runTCPLatencyTest(test *ethrTest) {
	server := test.session.remoteAddr
	conn, err := net.Dial(tcp(ipVer), server+":"+tcpLatencyPort)
	if err != nil {
		log.Error("Error dialing the latency connection: %v", err)
		os.Exit(1)
		return
	}
	defer conn.Close()
	buffSize := test.testParam.BufferSize
	// TODO Override buffer size to 1 for now. Evaluate if we need to allow
	// client to specify the buffer size in future.
	buffSize = 1
	buff := make([]byte, buffSize)
	for i := uint32(0); i < buffSize; i++ {
		buff[i] = byte(i)
	}
	blen := len(buff)
	rttCount := test.testParam.RttCount
	latencyNumbers := make([]time.Duration, rttCount)
ExitForLoop:
	for {
	ExitSelect:
		select {
		case <-test.done:
			break ExitForLoop
		default:
			for i := uint32(0); i < rttCount; i++ {
				s1 := time.Now()
				n, err := conn.Write(buff)
				if err != nil {
					// ui.printErr(err)
					// return
					break ExitSelect
				}
				if n < blen {
					// ui.printErr("Partial write: " + strconv.Itoa(n))
					// return
					break ExitSelect
				}
				_, err = io.ReadFull(conn, buff)
				if err != nil {
					// ui.printErr(err)
					// return
					break ExitSelect
				}
				e2 := time.Since(s1)
				latencyNumbers[i] = e2
			}
			// TODO temp code, fix it better, this is to allow server to do
			// server side latency measurements as well.
			_, _ = conn.Write(buff)

			calcLatency(test, rttCount, latencyNumbers)
		}
	}
}

/*
func runUDPBandwidthTest(test *ethrTest) {
	server := test.session.remoteAddr
	for th := uint32(0); th < test.testParam.NumThreads; th++ {
		go func() {
			buff := make([]byte, test.testParam.BufferSize)
			conn, err := net.Dial(udp(ipVer), server+":"+udpBandwidthPort)
			if err != nil {
				log.Debug("Unable to dial UDP, error: %v", err)
				return
			}
			defer conn.Close()
			ec := test.newConn(conn)
			rserver, rport, _ := net.SplitHostPort(conn.RemoteAddr().String())
			lserver, lport, _ := net.SplitHostPort(conn.LocalAddr().String())
			log.Info("[%3d] local %s port %s connected to %s port %s",
				ec.fd, lserver, lport, rserver, rport)
			blen := len(buff)
		ExitForLoop:
			for {
				select {
				case <-test.done:
					break ExitForLoop
				default:
					n, err := conn.Write(buff)
					if err != nil {
						log.Debug("%v", err)
						continue
					}
					if n < blen {
						log.Debug("Partial write: %d", n)
						continue
					}
					atomic.AddUint64(&ec.data, uint64(n))
					atomic.AddUint64(&test.testResult.data, uint64(n))
				}
			}
		}()
	}
}

func runUDPPpsTest(test *ethrTest) {
	server := test.session.remoteAddr
	for th := uint32(0); th < test.testParam.NumThreads; th++ {
		go func() {
			buff := make([]byte, test.testParam.BufferSize)
			conn, err := net.Dial(udp(ipVer), server+":"+udpPpsPort)
			if err != nil {
				log.Debug("Unable to dial UDP, error: %v", err)
				return
			}
			defer conn.Close()
			rserver, rport, _ := net.SplitHostPort(conn.RemoteAddr().String())
			lserver, lport, _ := net.SplitHostPort(conn.LocalAddr().String())
			log.Info("[udp] local %s port %s connected to %s port %s",
				lserver, lport, rserver, rport)
			blen := len(buff)
		ExitForLoop:
			for {
				select {
				case <-test.done:
					break ExitForLoop
				default:
					n, err := conn.Write(buff)
					if err != nil {
						log.Debug("%v", err)
						continue
					}
					if n < blen {
						log.Debug("Partial write: %d", n)
						continue
					}
					atomic.AddUint64(&test.testResult.data, 1)
				}
			}
		}()
	}
}
*/
func calcLatency(test *ethrTest, rttCount uint32, latencyNumbers []time.Duration) {
	sum := int64(0)
	for _, d := range latencyNumbers {
		sum += d.Nanoseconds()
	}
	elapsed := time.Duration(sum / int64(rttCount))
	sort.SliceStable(latencyNumbers, func(i, j int) bool {
		return latencyNumbers[i] < latencyNumbers[j]
	})
	//
	// Special handling for rttCount == 1. This prevents negative index
	// in the latencyNumber index. The other option is to use
	// roundUpToZero() but that is more expensive.
	//
	rttCountFixed := rttCount
	if rttCountFixed == 1 {
		rttCountFixed = 2
	}
	avg := elapsed
	min := latencyNumbers[0]
	max := latencyNumbers[rttCount-1]
	p50 := latencyNumbers[((rttCountFixed*50)/100)-1]
	p90 := latencyNumbers[((rttCountFixed*90)/100)-1]
	p95 := latencyNumbers[((rttCountFixed*95)/100)-1]
	p99 := latencyNumbers[((rttCountFixed*99)/100)-1]
	p999 := latencyNumbers[uint64(((float64(rttCountFixed)*99.9)/100)-1)]
	p9999 := latencyNumbers[uint64(((float64(rttCountFixed)*99.99)/100)-1)]
	ui.emitLatencyResults(
		test.session.remoteAddr,
		protoToString(test.testParam.TestID.Protocol),
		avg, min, max, p50, p90, p95, p99, p999, p9999)
}
