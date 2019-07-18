// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package g

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/toolkits/file"
)

type Ssh struct {
	Ip_addr    string `json:"ip_addr"`
	Ip_port    int    `json:"ip_port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Path       string `json:"path"`
	PrivateKey string `json:"privatekey"`
}

type PluginConfig struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
	Ssh     Ssh    `json:"ssh"`
	LogDir  string `json:"logs"`
}

type HeartbeatConfig struct {
	Enabled  bool   `json:"enabled"`
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

type TransferConfig struct {
	Enabled  bool     `json:"enabled"`
	Addrs    []string `json:"addrs"`
	Interval int      `json:"interval"`
	Timeout  int      `json:"timeout"`
}

type HttpConfig struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
	Backdoor bool   `json:"backdoor"`
}

type CollectorConfig struct {
	IfacePrefix []string `json:"ifacePrefix"`
	MountPoint  []string `json:"mountPoint"`
}
type Net_speed struct {
	IsServer bool   `json:"isServer"`
	IsTest   bool   `json:"isTest"`
	BufLen   string `json:"bufflength"`
	Duration int    `json:"duration"`
	Threads  int    `json:"threads"`
	Port     int    `json:"port"`
	Interval int    `json:"interval"`
}

type GlobalConfig struct {
	Debug         bool              `json:"debug"`
	Hostname      string            `json:"hostname"`
	IP            string            `json:"ip"`
	Hostname2ip   bool              `json:"hostname2ip"`
	Plugin        *PluginConfig     `json:"plugin"`
	Heartbeat     *HeartbeatConfig  `json:"heartbeat"`
	Transfer      *TransferConfig   `json:"transfer"`
	Http          *HttpConfig       `json:"http"`
	Collector     *CollectorConfig  `json:"collector"`
	DefaultTags   map[string]string `json:"default_tags"`
	IgnoreMetrics map[string]bool   `json:"ignore"`
	Net_speed     *Net_speed        `json:"net_speed"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func Hostname() (string, error) {

	if Config().Hostname2ip {
		return IP(), nil
	}

	hostname := Config().Hostname
	if hostname != "" {
		return hostname, nil
	}

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

func Real_Hostname() (string, error) {

	hostname := Config().Hostname
	if hostname != "" {
		return hostname, nil
	}

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
	ip := Config().IP
	if ip != "" {
		// use ip in configuration
		return ip
	}

	if len(LocalIp) > 0 {
		ip = LocalIp
	}
	var host_ip string = ""
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
	//log.Printf(" get system ip %s --- real ip %s",host_ip,ip)
	if (ip == "127.0.0.1" || ip == "::1" || ip == "localhost") && len(host_ip) > 1 {
		ip = host_ip
	}
	return ip
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c
	log.Println("read config file:", cfg, "successfully")
}
