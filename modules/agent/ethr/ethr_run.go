package ethr

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"runtime"
	"time"
)

func Run(ServerFlag bool) {
	//
	// Set GOMAXPROCS to 1024 as running large number of goroutines in a loop
	// to send network traffic results in timer starvation, as well as unfair
	// processing time across goroutines resulting in starvation of many TCP
	// connections. Using a higher number of threads via GOMAXPROCS solves this
	// problem.
	//
	runtime.GOMAXPROCS(1024)
	testTypePtr := g.Config().Net_speed.TestTypeStr // flag.String("t", "", "")
	thCount := g.Config().Net_speed.ThCount         // flag.Int("n", 1, "")
	bufLenStr := g.Config().Net_speed.BufLen        //flag.String("l", "16KB", "")
	duration := g.Config().Net_speed.Duration       //flag.Duration("d", 10*time.Second, "")
	rttCount := g.Config().Net_speed.RttCount       //flag.Int("i", 1000, "")
	Port := g.Config().Net_speed.Port               //flag.String("ports", "", "")
	use6 := g.Config().Net_speed.Ipv6               //flag.Bool("6", false, "")
	gap := g.Config().Net_speed.Gap                 //flag.Duration("g", 0, "")
	reverse := g.Config().Net_speed.Reverse         //flag.Bool("r", false, "")

	if reverse && ServerFlag {
		log.Error("Invalid arguments, reverse can only be used in client mode.")
	}
	if use6 {
		ipVer = ethrIPv6
	} else {
		ipVer = ethrIPv4
	}
	bufLen := unitToNumber(bufLenStr)
	if bufLen == 0 {
		log.Error(fmt.Sprintf("Invalid length specified: %s" + bufLenStr))
	}

	if rttCount <= 0 {
		log.Error(fmt.Sprintf("Invalid RTT count for latency test: %d", rttCount))
	}

	var testType EthrTestType
	switch testTypePtr {
	case "b":
		testType = Bandwidth
	case "c":
		testType = Cps
	case "p":
		testType = Pps
	case "l":
		testType = Latency
	case "cl":
		testType = ConnLatency
	default:
		log.Error(fmt.Sprintf("Invalid value %s specified for parameter testType.\n", testTypePtr))
	}
	proto := TCP
	if thCount <= 0 {
		thCount = runtime.NumCPU()
	}
	//
	// For Pkt/s, we always override the buffer size to be just 1 byte.
	// TODO: Evaluate in future, if we need to support > 1 byte packets for
	//       Pkt/s testing.
	//
	if testType == Pps {
		bufLen = 1
	}

	testParam := EthrTestParam{EthrTestID{EthrProtocol(proto), testType},
		uint32(thCount),
		uint32(bufLen),
		uint32(rttCount),
		reverse}
	validateTestParam(ServerFlag, testParam)

	generatePortNumbers(Port)

	clientParam := ethrClientParam{time.Duration(duration) * time.Second, time.Duration(gap) * time.Second}
	//todo 要判断是否运行定时检测，如果要定时检测，就需要提供ip列表，本机ip除外，每隔多少时间，扫描一次，并上报
	if ServerFlag {
		runServer(testParam)
	} else {
		go runClient(testParam, clientParam, g.Config().Plugin.Dir+"ethr_client.txt")
	}
}

func emitUnsupportedTest(testParam EthrTestParam) {
	log.Error(fmt.Sprintf("%s test for %s is not supported.\n",
		testToString(testParam.TestID.Type), protoToString(testParam.TestID.Protocol)))
}

func printReverseModeError() {
	log.Error("Reverse mode is only supported for TCP Bandwidth tests.")
}

func validateTestParam(sFlag bool, testParam EthrTestParam) {
	testType := testParam.TestID.Type
	protocol := testParam.TestID.Protocol
	if !sFlag {
		switch protocol {
		case TCP:
			if testType != Bandwidth && testType != Cps && testType != Latency {
				emitUnsupportedTest(testParam)
			}
			if testParam.Reverse && testType != Bandwidth {
				printReverseModeError()
			}
		default:
			emitUnsupportedTest(testParam)
		}
	}
}
