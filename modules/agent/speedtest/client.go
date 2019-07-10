package speedtest

//  author liyc
// 客户端只管发，然后发送结束标识，server返回结束标识，如果收到结束标识，标识收取完成，总的数据量=buff+flag_size*2
// 时间=(收到标识，time)-start
//本期只支持client->server 模式
import (
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	. "io/ioutil"
	"math"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var data uint64 = 0

func runDurationTimer(d time.Duration, toStop chan int) {
	go func() {
		dSeconds := uint64(d.Seconds())
		if dSeconds == 0 {
			return
		}
		time.Sleep(d)
		toStop <- 1
		close(toStop)
	}()
}

func Client() {
	var wg sync.WaitGroup
	port := ":" + strconv.Itoa(g.Config().Net_speed.Port)
	buffersize := unitToNumber(g.Config().Net_speed.BufLen)
	for {
		var client_list []string
		log.Info(filepath.Join(g.Config().Plugin.Dir, "/ethr_client.txt"))
		if contents, err := ReadFile(filepath.Join(g.Config().Plugin.Dir, "/ethr_client.txt")); err == nil {
			client_list = strings.Split(string(contents), "\n")
		} else {
			log.Error("打开测试客户端列表的文件打开失败" + g.Config().Plugin.Dir + "/ethr_client.txt" + err.Error())
			break
		}
		for vv := range client_list {

			if len(client_list[vv]) < 5 {
				log.Error("err client addr", client_list[vv])
				continue
			}
			log.Info("client begin  client ----------", client_list[vv])
			startTime := time.Now()
			data = 0
			log.Printf("client begin with %d threads", g.Config().Net_speed.Threads)

			toStop := make(chan int, 1)
			log.Debug("run time is ", g.Config().Net_speed.Duration)
			runDurationTimer(time.Duration(g.Config().Net_speed.Duration)*time.Second, toStop)
			for th := 0; th < g.Config().Net_speed.Threads; th++ {
				buff := make([]byte, buffersize)
				for i := uint64(0); i < buffersize; i++ {
					buff[i] = byte(i)
				}
				wg.Add(1)
				go func() {
					server := "[" + client_list[vv] + "]" + port
					log.Info("begin to test to " + server)
					conn, err := net.Dial("tcp", server)
					if err != nil {
						log.Fatalf("Could not connect: %s", err)
					}
					defer conn.Close()
					wg.Done()
				ExitForLoop:
					for {
						select {
						case <-toStop:
							log.Debug(" there is break")
							break ExitForLoop
						default:
							w, err := SendData(conn, buff)
							if err != nil {
								log.Printf("Error: %s", err)
							} else {
								atomic.AddUint64(&data, uint64(w))
							}
						}
					}
				}()
			}
			//todo 发送消息
			wg.Wait()
			usedtime := time.Since(startTime)
			ratio := float64(float64(data*8)/usedtime.Seconds()) / 1000.0 / 1000.0
			log.Printf("---Throughput--: %fMib/s --- %d", ratio, data)
			mvs := model.MetricValue{
				g.IP(),
				"network.test",
				math.Trunc(ratio),
				int64(g.Config().Net_speed.Interval),
				"GAUGE",
				g.IP() + "." + client_list[vv],
				time.Now().Unix()}
			var metrics []*model.MetricValue
			metrics = append(metrics, &mvs)
			g.SendToTransfer(metrics)
			time.Sleep(10 * time.Second)
		}
		time.Sleep(time.Duration(g.Config().Net_speed.Interval) * time.Second)
	}
}
