package ethr

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"os"
	"sort"
	"sync/atomic"
	"time"
)

var gCert []byte

func runServer(testParam EthrTestParam) {
	runTCPBandwidthServer()
	runTCPCpsServer()
	runTCPLatencyServer()
}
func runTCPBandwidthServer() {
	l, err := net.Listen(tcp(ipVer), hostAddr+":"+tcpBandwidthPort)
	if err != nil {
		log.Info("Fatal error listening on "+tcpBandwidthPort+" for TCP bandwidth tests: %v", err)
		os.Exit(1)
	}
	log.Info("Listening on " + tcpBandwidthPort + " for TCP bandwidth tests")
	go func(l net.Listener) {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				//ui.printErr("Error accepting new bandwidth connection: %v", err)
				continue
			}
			server, port, _ := net.SplitHostPort(conn.RemoteAddr().String())
			test := getTest(server, TCP, Bandwidth)
			if test == nil {
				log.Info("Received unsolicited TCP connection on port %s from %s port %s", tcpBandwidthPort, server, port)
				conn.Close()
				continue
			}
			go runTCPBandwidthHandler(conn, test)
		}
	}(l)
}

func closeConn(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		log.Info("Failed to close TCP connection, error: %v", err)
	}
}

func runTCPBandwidthHandler(conn net.Conn, test *ethrTest) {
	defer closeConn(conn)
	size := test.testParam.BufferSize
	buff := make([]byte, size)
	for i := uint32(0); i < test.testParam.BufferSize; i++ {
		buff[i] = byte(i)
	}
ExitForLoop:
	for {
		select {
		case <-test.done:
			break ExitForLoop
		default:
			var err error
			if test.testParam.Reverse {
				_, err = conn.Write(buff)
			} else {
				_, err = io.ReadFull(conn, buff)
			}
			if err != nil {
				log.Info("Error sending/receiving data on a connection for bandwidth test: %v", err)
				continue
			}
			atomic.AddUint64(&test.testResult.data, uint64(size))
		}
	}
}

func runTCPCpsServer() {
	l, err := net.Listen(tcp(ipVer), hostAddr+":"+tcpCpsPort)
	if err != nil {
		fmt.Printf("Fatal error listening on "+tcpCpsPort+" for TCP conn/s tests: %v", err)
		os.Exit(1)
	}
	log.Info("Listening on " + tcpCpsPort + " for TCP conn/s tests")
	go func(l net.Listener) {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Info("Error accepting new conn/s connection: %v", err)
				continue
			}
			go runTCPCpsHandler(conn)
		}
	}(l)
}

func runTCPCpsHandler(conn net.Conn) {
	defer conn.Close()
	server, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	test := getTest(server, TCP, Cps)
	if test != nil {
		atomic.AddUint64(&test.testResult.data, 1)
	} else {
		log.Info("Error: Unsolicited connection received.")
	}
}

func runTCPLatencyServer() {
	l, err := net.Listen(tcp(ipVer), hostAddr+":"+tcpLatencyPort)
	if err != nil {
		fmt.Printf("Fatal error listening on "+tcpLatencyPort+" for TCP latency tests: %v", err)
		os.Exit(1)
	}
	log.Info("Listening on " + tcpLatencyPort + " for TCP latency tests")
	go func(l net.Listener) {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Info("Error accepting new latency connection: %v", err)
				continue
			}
			server, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			test := getTest(server, TCP, Latency)
			if test == nil {
				conn.Close()
				continue
			}
			go runTCPLatencyHandler(conn, test)
		}
	}(l)
}

func runTCPLatencyHandler(conn net.Conn, test *ethrTest) {
	defer conn.Close()
	bytes := make([]byte, test.testParam.BufferSize)
	bytes = make([]byte, 1)
	rttCount := test.testParam.RttCount
	latencyNumbers := make([]time.Duration, rttCount)
	for {
		_, err := io.ReadFull(conn, bytes)
		if err != nil {
			log.Info("Error receiving data for latency test: %v", err)
			return
		}
		for i := uint32(0); i < rttCount; i++ {
			s1 := time.Now()
			_, err = conn.Write(bytes)
			if err != nil {
				log.Info("Error sending data for latency test: %v", err)
				return
			}
			_, err = io.ReadFull(conn, bytes)
			if err != nil {
				log.Info("Error receiving data for latency test: %v", err)
				return
			}
			e2 := time.Since(s1)
			latencyNumbers[i] = e2
		}
		sum := int64(0)
		for _, d := range latencyNumbers {
			sum += d.Nanoseconds()
		}
		elapsed := time.Duration(sum / int64(rttCount))
		sort.SliceStable(latencyNumbers, func(i, j int) bool {
			return latencyNumbers[i] < latencyNumbers[j]
		})
		rttCountFixed := rttCount
		if rttCountFixed == 1 {
			rttCountFixed = 2
		}
		atomic.SwapUint64(&test.testResult.data, uint64(elapsed.Nanoseconds()))
		//avg := elapsed
		//min := latencyNumbers[0]
		//max := latencyNumbers[rttCount-1]
		//p50 := latencyNumbers[((rttCountFixed*50)/100)-1]
		//p90 := latencyNumbers[((rttCountFixed*90)/100)-1]
		//p95 := latencyNumbers[((rttCountFixed*95)/100)-1]
		//p99 := latencyNumbers[((rttCountFixed*99)/100)-1]
		//p999 := latencyNumbers[uint64(((float64(rttCountFixed)*99.9)/100)-1)]
		//p9999 := latencyNumbers[uint64(((float64(rttCountFixed)*99.99)/100)-1)]
		//log.Info(
		//	test.session.remoteAddr
		//	protoToString(test.testParam.TestID.Protocol),
		//	avg, min, max, p50, p90, p95, p99, p999, p9999)
	}
}

/*
func runUDPBandwidthServer(test *ethrTest) error {
	udpAddr, err := net.ResolveUDPAddr(udp(ipVer), hostAddr+":"+udpBandwidthPort)
	if err != nil {
		log.Info("Unable to resolve UDP address: %v", err)
		return err
	}
	l, err := net.ListenUDP(udp(ipVer), udpAddr)
	if err != nil {
		log.Info("Error listening on %s for UDP pkt/s tests: %v", udpPpsPort, err)
		return err
	}
	go func(l *net.UDPConn) {
		defer l.Close()
		//
		// We use NumCPU here instead of NumThreads passed from client. The
		// reason is that for UDP, there is no connection, so all packets come
		// on same CPU, so it isn't clear if there are any benefits to running
		// more threads than NumCPU(). TODO: Evaluate this in future.
		//
		for i := 0; i < runtime.NumCPU(); i++ {
			go runUDPBandwidthHandler(test, l)
		}
		<-test.done
	}(l)
	return nil
}

func runUDPBandwidthHandler(test *ethrTest, conn *net.UDPConn) {
	buffer := make([]byte, test.testParam.BufferSize)
	n, remoteAddr, err := 0, new(net.UDPAddr), error(nil)
	for err == nil {
		n, remoteAddr, err = conn.ReadFromUDP(buffer)
		if err != nil {
			log.Info("Error receiving data from UDP for bandwidth test: %v", err)
			continue
		}
		ethrUnused(n)
		server, port, _ := net.SplitHostPort(remoteAddr.String())
		test := getTest(server, UDP, Bandwidth)
		if test != nil {
			atomic.AddUint64(&test.testResult.data, uint64(n))
		} else {
			log.Info("Received unsolicited UDP traffic on port %s from %s port %s", udpPpsPort, server, port)
		}
	}
}
*/
/*
func runUDPPpsServer(test *ethrTest) error {
	udpAddr, err := net.ResolveUDPAddr(udp(ipVer), hostAddr+":"+udpPpsPort)
	if err != nil {
		log.Info("Unable to resolve UDP address: %v", err)
		return err
	}
	l, err := net.ListenUDP(udp(ipVer), udpAddr)
	if err != nil {
		log.Info("Error listening on %s for UDP pkt/s tests: %v", udpPpsPort, err)
		return err
	}
	go func(l *net.UDPConn) {
		defer l.Close()
		for i := 0; i < runtime.NumCPU(); i++ {
			go runUDPPpsHandler(test, l)
		}
		<-test.done
	}(l)
	return nil
}

func runUDPPpsHandler(test *ethrTest, conn *net.UDPConn) {
	buffer := make([]byte, test.testParam.BufferSize)
	n, remoteAddr, err := 0, new(net.UDPAddr), error(nil)
	for err == nil {
		n, remoteAddr, err = conn.ReadFromUDP(buffer)
		if err != nil {
			log.Info("Error receiving data from UDP for pkt/s test: %v", err)
			continue
		}
		ethrUnused(n)
		server, port, _ := net.SplitHostPort(remoteAddr.String())
		test := getTest(server, UDP, Pps)
		if test != nil {
			atomic.AddUint64(&test.testResult.data, 1)
		} else {
			log.Info("Received unsolicited UDP traffic on port %s from %s port %s", udpPpsPort, server, port)
		}
	}
}
*/
