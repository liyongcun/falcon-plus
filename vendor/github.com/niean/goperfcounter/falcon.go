package goperfcounter

import "os"

/*
{
    "debug": false, // 是否开启调制，默认为false
    "hostname": "", // 机器名(也即endpoint名称)，默认为本机名称
    "tags": "", // tags标签，默认为空。一个tag形如"key=val"，多个tag用逗号分隔；name为保留字段，因此不允许设置形如"name=xxx"的tag。eg. "cop=xiaomi,module=perfcounter"
    "step": 60, // 上报周期，单位s，默认为60s
    "bases":[], // gvm基础信息采集，可选值为"debug"、"runtime"，默认不采集
    "push": { // push数据到Open-Falcon
        "enabled":true, // 是否开启自动push，默认开启
        "api": "" // Open-Falcon接收器地址，默认为本地agent，即"http:// 127.0.0.1:1988/v1/push"
    },
    "http": { // http服务，为了安全考虑，当前只允许本地访问
        "enabled": false, // 是否开启http服务，默认不开启
        "listen": "" // http服务监听地址，默认为空。eg. "0.0.0.0:2015"表示在2015端口开启http监听
    }
}
*/
import (
	"encoding/json"
	"fmt"
	//"github.com/open-falcon/falcon-plus/modules/agent/g"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/niean/go-metrics-lite"
	bhttp "github.com/niean/gotools/http/httpclient/beego"
)

const (
	GAUGE = "GAUGE"
)

func Hostname() (string, error) {
	hostname := cfg.Hostname

	if os.Getenv("FALCON_ENDPOINT") != "" {
		hostname = os.Getenv("FALCON_ENDPOINT")
		return hostname, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("ERROR: os.Hostname() fail", err)
	}
	hostname_bak := IP()
	if strings.Contains(hostname, "localhost") && len(hostname_bak) > 7 && hostname_bak != "127.0.0.1" {
		hostname = hostname_bak
	}
	return hostname, err
}

func IP() string {
	host_ip := "127.0.0.1"
	out, err := exec.Command("hostname", "--all-ip-addresses").Output()
	if err == nil {
		ip_list := strings.Split(strings.Replace(string(out), "\n", "", -1), " ")
		for vv := range ip_list {
			if ip_list[vv] == "127.0.0.1" || ip_list[vv] == "::1" || strings.Contains(ip_list[vv], "localhost") || len(ip_list[vv]) < 5 {
				continue
			}
			host_ip = ip_list[vv]
			break
		}
	} else {
		log.Println("hostname --all-ip-addresses err" + err.Error())
	}
	return host_ip
}

func pushToFalcon() {
	cfg := config()
	step := cfg.Step
	api := cfg.Push.Api
	debug := cfg.Debug

	// align push start ts
	alignPushStartTs(step)

	for _ = range time.Tick(time.Duration(step) * time.Second) {
		selfMeter("pfc.push.cnt", 1) // statistics

		fms := falconMetrics()
		start := time.Now()
		err := push(fms, api, debug)
		selfGauge("pfc.push.ms", int64(time.Since(start)/time.Millisecond)) // statistics

		if err != nil {
			if debug {
				log.Printf("[perfcounter] send to %s error: %v", api, err)
			}
			selfGauge("pfc.push.size", int64(0)) // statistics
		} else {
			selfGauge("pfc.push.size", int64(len(fms))) // statistics
		}
	}
}

func falconMetric(types []string) []*MetricValue {
	fd := []*MetricValue{}
	for _, ty := range types {
		if r, ok := values[ty]; ok && r != nil {
			data := _falconMetric(r)
			fd = append(fd, data...)
		}
	}
	return fd
}

func falconMetrics() []*MetricValue {
	data := make([]*MetricValue, 0)
	for _, r := range values {
		nd := _falconMetric(r)
		data = append(data, nd...)
	}
	return data
}

// internal
func _falconMetric(r metrics.Registry) []*MetricValue {
	cfg := config()
	endpoint := cfg.Hostname
	hostname_bak, h_err := Hostname()
	if h_err != nil {
		println("get hostname err")
	}
	if strings.Contains(endpoint, "localhost") {
		endpoint = hostname_bak
	}
	step := cfg.Step
	tags := cfg.Tags
	ts := time.Now().Unix()

	data := make([]*MetricValue, 0)
	r.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Gauge:
			m := gaugeMetricValue(metric, name, endpoint, tags, step, ts)
			data = append(data, m...)
		case metrics.GaugeFloat64:
			m := gaugeFloat64MetricValue(metric, name, endpoint, tags, step, ts)
			data = append(data, m...)
		case metrics.Counter:
			m := counterMetricValue(metric, name, endpoint, tags, step, ts)
			data = append(data, m...)
		case metrics.Meter:
			m := metric.Snapshot()
			ms := meterMetricValue(m, name, endpoint, tags, step, ts)
			data = append(data, ms...)
		case metrics.Histogram:
			h := metric.Snapshot()
			ms := histogramMetricValue(h, name, endpoint, tags, step, ts)
			data = append(data, ms...)
		}
	})

	return data
}

func gaugeMetricValue(metric metrics.Gauge, metricName, endpoint, oldtags string, step, ts int64) []*MetricValue {
	tags := getTags(metricName, oldtags)
	c := newMetricValue(endpoint, "value", metric.Value(), step, GAUGE, tags, ts)
	return []*MetricValue{c}
}

func gaugeFloat64MetricValue(metric metrics.GaugeFloat64, metricName, endpoint, oldtags string, step, ts int64) []*MetricValue {
	tags := getTags(metricName, oldtags)
	c := newMetricValue(endpoint, "value", metric.Value(), step, GAUGE, tags, ts)
	return []*MetricValue{c}
}

func counterMetricValue(metric metrics.Counter, metricName, endpoint, oldtags string, step, ts int64) []*MetricValue {
	tags := getTags(metricName, oldtags)
	c1 := newMetricValue(endpoint, "count", metric.Count(), step, GAUGE, tags, ts)
	return []*MetricValue{c1}
}

func meterMetricValue(metric metrics.Meter, metricName, endpoint, oldtags string, step, ts int64) []*MetricValue {
	data := make([]*MetricValue, 0)
	tags := getTags(metricName, oldtags)

	c1 := newMetricValue(endpoint, "rate", metric.RateStep(), step, GAUGE, tags, ts)
	c2 := newMetricValue(endpoint, "sum", metric.Count(), step, GAUGE, tags, ts)
	data = append(data, c1, c2)

	return data
}

func histogramMetricValue(metric metrics.Histogram, metricName, endpoint, oldtags string, step, ts int64) []*MetricValue {
	data := make([]*MetricValue, 0)
	tags := getTags(metricName, oldtags)

	values := make(map[string]interface{})
	ps := metric.Percentiles([]float64{0.75, 0.95, 0.99})
	values["min"] = metric.Min()
	values["max"] = metric.Max()
	values["mean"] = metric.Mean()
	values["75th"] = ps[0]
	values["95th"] = ps[1]
	values["99th"] = ps[2]
	for key, val := range values {
		c := newMetricValue(endpoint, key, val, step, GAUGE, tags, ts)
		data = append(data, c)
	}

	return data
}

func newMetricValue(endpoint, metric string, value interface{}, step int64, t, tags string, ts int64) *MetricValue {
	return &MetricValue{
		Endpoint:  endpoint,
		Metric:    metric,
		Value:     value,
		Step:      step,
		Type:      t,
		Tags:      tags,
		Timestamp: ts,
	}
}

func getTags(name string, tags string) string {
	if tags == "" {
		return fmt.Sprintf("name=%s", name)
	}
	return fmt.Sprintf("%s,name=%s", tags, name)
}

//
func push(data []*MetricValue, url string, debug bool) error {
	dlen := len(data)
	pkg := 200 //send pkg items once
	sent := 0
	for {
		if sent >= dlen {
			break
		}

		end := sent + pkg
		if end > dlen {
			end = dlen
		}

		pkgData := data[sent:end]
		jr, err := json.Marshal(pkgData)
		if err != nil {
			return err
		}

		response, err := bhttp.Post(url).Body(jr).String()
		if err != nil {
			return err
		}
		sent = end

		if debug {
			log.Printf("[perfcounter] push result: %v, data: %v\n", response, pkgData)
		}
	}
	return nil
}

//
func alignPushStartTs(stepSec int64) {
	nw := time.Duration(time.Now().UnixNano())
	step := time.Duration(stepSec) * time.Second
	sleepNano := step - nw%step
	if sleepNano > 0 {
		time.Sleep(sleepNano)
	}
}

//
type MetricValue struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

func (this *MetricValue) String() string {
	return fmt.Sprintf(
		"<Endpoint:%s, Metric:%s, Tags:%s, Type:%s, Step:%d, Timestamp:%d, Value:%v>",
		this.Endpoint,
		this.Metric,
		this.Tags,
		this.Type,
		this.Step,
		this.Timestamp,
		this.Value,
	)
}
