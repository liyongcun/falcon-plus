package ethr

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"net"
	"os"
	"strconv"
	"strings"
)

type ethrNetDevInfo struct {
	bytes      uint64
	packets    uint64
	drop       uint64
	errs       uint64
	fifo       uint64
	frame      uint64
	compressed uint64
	multicast  uint64
}

func getNetDevStats(stats *ethrNetStat) {
	ifs, err := net.Interfaces()
	if err != nil {
		log.Error("%v", err)
		return
	}

	netStatsFile, err := os.Open("/proc/net/dev")
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer netStatsFile.Close()

	reader := bufio.NewReader(netStatsFile)

	// Pass the header
	// Inter-|   Receive                                             |  Transmit
	//  face |bytes packets errs drop fifo frame compressed multicast|bytes packets errs drop fifo colls carrier compressed
	reader.ReadString('\n')
	reader.ReadString('\n')

	var line string
	for err == nil {
		line, err = reader.ReadString('\n')
		if line == "" {
			continue
		}
		netDevStat := buildNetDevStat(line)
		if isIfUp(netDevStat.interfaceName, ifs) {
			stats.netDevStats = append(stats.netDevStats, buildNetDevStat(line))
		}
	}
}

func buildNetDevStat(line string) ethrNetDevStat {
	fields := strings.Fields(line)
	interfaceName := strings.TrimSuffix(fields[0], ":")
	rxInfo := toNetDevInfo(fields[1:9])
	txInfo := toNetDevInfo(fields[9:17])
	return ethrNetDevStat{
		interfaceName: interfaceName,
		rxBytes:       rxInfo.bytes,
		txBytes:       txInfo.bytes,
		rxPkts:        rxInfo.packets,
		txPkts:        txInfo.packets,
	}
}

func toNetDevInfo(fields []string) ethrNetDevInfo {
	return ethrNetDevInfo{
		bytes:      toUInt64(fields[0]),
		packets:    toUInt64(fields[1]),
		errs:       toUInt64(fields[2]),
		drop:       toUInt64(fields[3]),
		fifo:       toUInt64(fields[4]),
		frame:      toUInt64(fields[5]),
		compressed: toUInt64(fields[6]),
		multicast:  toUInt64(fields[7]),
	}
}

func isIfUp(ifName string, ifs []net.Interface) bool {
	for _, ifi := range ifs {
		if ifi.Name == ifName {
			if (ifi.Flags & net.FlagUp) != 0 {
				return true
			}
			return false
		}
	}
	return false
}

func toUInt64(str string) uint64 {
	res, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		log.Debug("Error in string conversion: %v", err)
		return 0
	}
	return res
}

func getTCPStats(stats *ethrNetStat) {
	snmpStatsFile, err := os.Open("/proc/net/snmp")
	if err != nil {
		log.Debug("%v", err)
		return
	}
	defer snmpStatsFile.Close()

	reader := bufio.NewReader(snmpStatsFile)

	var line string
	for err == nil {
		// Tcp: RtoAlgorithm RtoMin RtoMax MaxConn ActiveOpens PassiveOpens AttemptFails EstabResets
		//      CurrEstab InSegs OutSegs RetransSegs InErrs OutRsts InCsumErrors
		line, err = reader.ReadString('\n')
		if line == "" || !strings.HasPrefix(line, "Tcp") {
			continue
		}
		// Skip the first line starting with Tcp
		line, err = reader.ReadString('\n')
		if !strings.HasPrefix(line, "Tcp") {
			break
		}
		fields := strings.Fields(line)
		stats.tcpStats.segRetrans = toUInt64(fields[12])
	}
}
