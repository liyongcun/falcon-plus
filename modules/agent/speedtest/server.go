package speedtest

//  author liyc
// 本期client的uuid不做验证，意义不大
//本期只支持client->server 模式
import (
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"net"
	"strconv"
)

func Server() {
	port := ":" + strconv.Itoa(g.Config().Net_speed.Port)
	buffersize := unitToNumber(g.Config().Net_speed.BufLen)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Could not listen on %s: %s", port, err)
	}
	go func(l net.Listener) {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Printf("Error accepting new bandwidth connection: %v", err)
				continue
			}
			log.Printf("Accepted connection from %s", conn.RemoteAddr())
			go handleConnection(conn, buffersize)
		}
	}(ln)
}

func handleConnection(conn net.Conn, buffersize uint64) {
	err := ReceiveData(conn, buffersize)
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
}
