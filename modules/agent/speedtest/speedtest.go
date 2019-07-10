package speedtest

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"math"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
)

type BytesPerTime struct {
	Bytes    uint64
	Duration time.Duration
}
type UUID [16]byte

var timeBase = time.Date(1582, time.October, 15, 0, 0, 0, 0, time.UTC).Unix()
var hardwareAddr []byte
var clockSeq uint32

func TimeUUID() UUID {
	return FromTime(time.Now())
}

const (
	// UNO represents 1 unit.
	UNO = 1

	// KILO represents k.
	KILO = 1024

	// MEGA represents m.
	MEGA = 1024 * 1024

	// GIGA represents g.
	GIGA = 1024 * 1024 * 1024

	// TERA represents t.
	TERA = 1024 * 1024 * 1024 * 1024
)

func unitToNumber(s string) uint64 {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	i := strings.IndexFunc(s, unicode.IsLetter)

	if i == -1 {
		bytes, err := strconv.ParseFloat(s, 64)
		if err != nil || bytes <= 0 {
			return 0
		}
		return uint64(bytes)
	}

	bytesString, multiple := s[:i], s[i:]
	bytes, err := strconv.ParseFloat(bytesString, 64)
	if err != nil || bytes <= 0 {
		return 0
	}

	switch multiple {
	case "T", "TB", "TIB":
		return uint64(bytes * TERA)
	case "G", "GB", "GIB":
		return uint64(bytes * GIGA)
	case "M", "MB", "MIB":
		return uint64(bytes * MEGA)
	case "K", "KB", "KIB":
		return uint64(bytes * KILO)
	case "B":
		return uint64(bytes)
	default:
		return 0
	}
}
func FromTime(aTime time.Time) UUID {
	var u UUID
	utcTime := aTime.In(time.UTC)
	t := uint64(utcTime.Unix()-timeBase)*10000000 + uint64(utcTime.Nanosecond()/100)
	u[0], u[1], u[2], u[3] = byte(t>>24), byte(t>>16), byte(t>>8), byte(t)
	u[4], u[5] = byte(t>>40), byte(t>>32)
	u[6], u[7] = byte(t>>56)&0x0F, byte(t>>48)

	clock := atomic.AddUint32(&clockSeq, 1)
	u[8] = byte(clock >> 8)
	u[9] = byte(clock)

	copy(u[10:], hardwareAddr)

	u[6] |= 0x10 // set version to 1 (time based uuid)
	u[8] &= 0x3F // clear variant
	u[8] |= 0x80 // set to IETF variant

	return u
}

func (u UUID) String() string {
	var offsets = [...]int{0, 2, 4, 6, 9, 11, 14, 16, 19, 21, 24, 26, 28, 30, 32, 34}
	const hexString = "0123456789abcdef"
	r := make([]byte, 36)
	for i, b := range u {
		r[offsets[i]] = hexString[b>>4]
		r[offsets[i]+1] = hexString[b&0xF]
	}
	r[8] = '-'
	r[13] = '-'
	r[18] = '-'
	r[23] = '-'
	return string(r)
}

func IBytes(s uint64) string {
	sizes := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	return humanateBytes(s, 1024, sizes)
}
func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}
func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}
	return fmt.Sprintf(f, val, suffix)
}

func SendData(conn net.Conn, buffer []byte) (size int, err error) {
	w, err := conn.Write(buffer)
	if err != nil {
		return 0, fmt.Errorf("Error while writing: %s", err)
	}
	//log.Info("write data "+strconv.Itoa(w))
	if w != len(buffer) {
		return w, fmt.Errorf("Error while writing size: %s", w)
	}
	return w, nil
}

func ReceiveData(conn net.Conn, buffersize uint64) error {
	b := make([]byte, buffersize)
	defer conn.Close()
	//var t int =0
	for {
		w, err := conn.Read(b)
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			} else {
				return fmt.Errorf("Read: %d, Error: %s in conn read ", w, err)
			}
		}
		//t=t+w;
		//if t < buffersize{
		//	continue
		//}
		if w == 16 {
			log.Info("reback uuid=======")
			m := make([]byte, 0)
			for i := 0; i < 16; i++ {
				m = append(m, b[i])
			}
			w, err := conn.Write(m)
			if err != nil {
				return fmt.Errorf("Read: %d, Error: %s in seand uuid", w, err)
			}
		}
	}
}
