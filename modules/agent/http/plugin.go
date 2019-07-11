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

package http

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/sftp"
	"github.com/open-falcon/falcon-plus/modules/agent/g"
	"github.com/open-falcon/falcon-plus/modules/agent/plugins"
	"github.com/toolkits/file"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	//"os/exec"
	"time"
)

// TODO  add by liyc 这里修改为sftp的模式，因为ssh是远程访问的核心

type Sftp_client struct {
	clientConfig *ssh.ClientConfig
	sshClient    *ssh.Client
	sftpClient   *sftp.Client
	pubkey       ssh.AuthMethod
}

func (sshftp *Sftp_client) PublicKeyFile(keypath string) ssh.AuthMethod {
	if !file.IsExist(keypath) {
		return nil
	}
	buffer, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (sshftp *Sftp_client) Sftp_close() {
	sshftp.sftpClient.Close()
	sshftp.sshClient.Close()
}

func (sshftp *Sftp_client) Connect_init() error {
	var err error = nil

	if len(g.Config().Plugin.Ssh.User) <= 0 {
		return fmt.Errorf("ssh user not defined!")
	}
	if len(g.Config().Plugin.Ssh.Ip_addr) <= 0 {
		return fmt.Errorf("ssh  ip address  not defined!")
	}
	if g.Config().Plugin.Ssh.Ip_port <= 0 {
		return fmt.Errorf("ssh  ip  port  not defined!")
	}

	if len(g.Config().Plugin.Ssh.PrivateKey) <= 1 && len(g.Config().Plugin.Ssh.Password) <= 1 {
		return fmt.Errorf("ssh auth method not defined!,please use password or PrivateKey")
	}

	if len(g.Config().Plugin.Ssh.PrivateKey) >= 1 && len(g.Config().Plugin.Ssh.Password) >= 1 {
		log.Debug("ssh auth method password and PrivateKey both  defined , PrivateKey default")
	}

	if len(g.Config().Plugin.Ssh.PrivateKey) > 1 {
		sshftp.pubkey = sshftp.PublicKeyFile(g.Config().Plugin.Ssh.PrivateKey)
	}

	auth := make([]ssh.AuthMethod, 0)
	if sshftp.pubkey != nil {
		auth = append(auth, sshftp.pubkey)
	} else {
		auth = append(auth, ssh.Password(g.Config().Plugin.Ssh.Password))
	}

	sshftp.clientConfig = &ssh.ClientConfig{
		User:    g.Config().Plugin.Ssh.User,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	// connet to ssh
	addr := fmt.Sprintf("%s:%d", g.Config().Plugin.Ssh.Ip_addr, g.Config().Plugin.Ssh.Ip_port)
	if sshftp.sshClient, err = ssh.Dial("tcp", addr, sshftp.clientConfig); err != nil {
		return err
	}
	// create sftp client
	if sshftp.sftpClient, err = sftp.NewClient(sshftp.sshClient); err != nil {
		return err
	}
	return nil
}

/*
func configPluginRoutes() {
	http.HandleFunc("/plugin/update", func(w http.ResponseWriter, r *http.Request) {
		if !g.Config().Plugin.Enabled {
			w.Write([]byte("plugin not enabled"))
			return
		}

		dir := g.Config().Plugin.Dir
		parentDir := file.Dir(dir)
		file.InsureDir(parentDir)

		if file.IsExist(dir) {
			// git pull
			cmd := exec.Command("git", "pull")
			cmd.Dir = dir
			err := cmd.Run()
			if err != nil {
				w.Write([]byte(fmt.Sprintf("git pull in dir:%s fail. error: %s", dir, err)))
				return
			}
		} else {
			// git clone
			cmd := exec.Command("git", "clone", g.Config().Plugin.Git, file.Basename(dir))
			cmd.Dir = parentDir
			err := cmd.Run()
			if err != nil {
				w.Write([]byte(fmt.Sprintf("git clone in dir:%s fail. error: %s", parentDir, err)))
				return
			}
		}

		w.Write([]byte("success"))
	})

	http.HandleFunc("/plugin/reset", func(w http.ResponseWriter, r *http.Request) {
		if !g.Config().Plugin.Enabled {
			w.Write([]byte("plugin not enabled"))
			return
		}

		dir := g.Config().Plugin.Dir

		if file.IsExist(dir) {
			cmd := exec.Command("git", "reset", "--hard")
			cmd.Dir = dir
			err := cmd.Run()
			if err != nil {
				w.Write([]byte(fmt.Sprintf("git reset --hard in dir:%s fail. error: %s", dir, err)))
				return
			}
		}
		w.Write([]byte("success"))
	})

	http.HandleFunc("/plugins", func(w http.ResponseWriter, r *http.Request) {
		//TODO: not thread safe
		RenderDataJson(w, plugins.Plugins)
	})
}*/
func sftp_get(reset_flag bool) (msg string, erra error) {
	dir := g.Config().Plugin.Dir
	parentDir := file.Dir(dir)
	file.InsureDir(parentDir)
	sshCli := Sftp_client{nil, nil, nil, nil}
	err := sshCli.Connect_init()
	if err != nil {
		return "open ssh fail", err
	}
	defer sshCli.Sftp_close()
	if file.IsExist(dir) {
		// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
		if err != nil {
			log.Fatal(err)
		}
		wl := sshCli.sftpClient.Walk(g.Config().Plugin.Ssh.Path)
		if err != nil {
			return fmt.Sprintf("update using ssh err in dir:%s fail. error: %s", dir, err), err
		}
		for wl.Step() {
			aRel, err := filepath.Rel(g.Config().Plugin.Ssh.Path, wl.Path())
			if err != nil {
				return fmt.Sprintf("update using ssh get root path in dir:%s ,real path %s fail. error: %s", dir, wl.Path()), err
			}
			if aRel == "." || aRel == ".." {
				continue
			}
			//sftp文件信息
			//log.Debug("ssh file:"+ wl.Path())
			FileInfo, err := sshCli.sftpClient.Stat(wl.Path())
			if err != nil {
				return fmt.Sprintf("update using ssh get real path  stat in :%s fail. error: %s", wl.Path(), err), err
			}
			lRpath := filepath.Join(dir, aRel)
			log.Debug("ssh file:  " + aRel + " -----  local file : " + lRpath)
			if !FileInfo.IsDir() {
				//处理文件,比较时间
				if file.IsExist(lRpath) {
					//本地文件信息
					lfile, err := os.Stat(lRpath)
					if err != nil {
						return fmt.Sprintf("get local file info err :%s fail. error: %s", lRpath, err), err
					}
					if lfile.ModTime().Unix() >= FileInfo.ModTime().Unix() && !reset_flag {
						continue
					}
				}
				//这里关闭文件采用显示关闭，不然文件过多，出现标准的linux 错误：too many open file，因为defer的特性
				srcFile, err := sshCli.sftpClient.Open(wl.Path())
				//defer srcFile.Close()
				if err != nil {
					return fmt.Sprintf("open ssh  file  err :%s fail. error: %s", lRpath, err), err
				}
				dstFile, err := os.Create(lRpath)
				//defer dstFile.Close()
				if err != nil {
					return fmt.Sprintf("open local file  err :%s fail. error: %s", lRpath, err), err
				}
				if _, err = srcFile.WriteTo(dstFile); err != nil {
					return fmt.Sprintf("write  to  local file  err :%s fail. error: %s", lRpath, err), err
				}
				fe := srcFile.Close()
				if fe != nil {
					log.Error("close ssh file err ! ", fe.Error())
				}
				de := dstFile.Close()
				if de != nil {
					log.Error("close local plugin file err ! ", de.Error())
				}
			} else {
				file.InsureDir(lRpath)
			}
		}
	}
	return "Success", nil
}

func configPluginRoutes() {
	http.HandleFunc("/plugin/update", func(w http.ResponseWriter, r *http.Request) {
		if !g.Config().Plugin.Enabled {
			w.Write([]byte("plugin not enabled"))
			return
		}
		err_msg, err := sftp_get(false)
		if err != nil {
			w.Write([]byte("plugin not update : [" + err_msg + "]"))
			return
		}
		w.Write([]byte("success"))
	})
	http.HandleFunc("/plugin/reset", func(w http.ResponseWriter, r *http.Request) {
		if !g.Config().Plugin.Enabled {
			w.Write([]byte("plugin not enabled"))
			return
		}
		err_msg, err := sftp_get(true)
		if err != nil {
			w.Write([]byte("plugin not reset: [" + err_msg + "]"))
			return
		}
		w.Write([]byte("success"))
	})
	http.HandleFunc("/plugins", func(w http.ResponseWriter, r *http.Request) {
		//TODO: not thread safe
		RenderDataJson(w, plugins.Plugins)
	})
}
