/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"encoding/json"
	"github.com/zouyx/agollo/v4/component/log"
	"strings"
	"sync"
	"time"
)

var (
	//next try connect period - 60 second
	nextTryConnectPeriod int64 = 60
)

//File 读写配置文件
type File interface {
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)

	Write(content interface{}, configPath string) error
}

//AppConfig 配置文件
type AppConfig struct {
	AppID            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	IP               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	IsBackupConfig   bool   `default:"true" json:"isBackupConfig"`
	BackupConfigPath string `json:"backupConfigPath"`
	Secret           string `json:"secret"`
	//real servers ip
	servers sync.Map
}

//ServerInfo 服务器信息
type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceID  string `json:"instanceId"`
	HomepageURL string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

//GetIsBackupConfig whether backup config after fetch config from apollo
//false : no
//true : yes (default)
func (a *AppConfig) GetIsBackupConfig() bool {
	return a.IsBackupConfig
}

//GetBackupConfigPath GetBackupConfigPath
func (a *AppConfig) GetBackupConfigPath() string {
	return a.BackupConfigPath
}

//GetHost GetHost
func (a *AppConfig) GetHost() string {
	if strings.HasPrefix(a.IP, "http") {
		if !strings.HasSuffix(a.IP, "/") {
			return a.IP + "/"
		}
		return a.IP
	}
	return "http://" + a.IP + "/"
}

//SetNextTryConnTime if this connect is fail will set this time
func (a *AppConfig) SetNextTryConnTime(nextTryConnectPeriod int64) {
	a.NextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
}

//IsConnectDirectly is connect by ip directly
//false : no
//true : yes
func (a *AppConfig) IsConnectDirectly() bool {
	if a.NextTryConnTime >= 0 && a.NextTryConnTime > time.Now().Unix() {
		return true
	}

	return false
}

//SyncServerIPListSuccessCallBack 同步服务器列表成功后的回调
func (a *AppConfig) SyncServerIPListSuccessCallBack(responseBody []byte) (o interface{}, err error) {
	log.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*ServerInfo, 0)

	err = json.Unmarshal(responseBody, &tmpServerInfo)

	if err != nil {
		log.Error("Unmarshal json Fail,Error:", err)
		return
	}

	if len(tmpServerInfo) == 0 {
		log.Info("get no real server!")
		return
	}

	for _, server := range tmpServerInfo {
		if server == nil {
			continue
		}
		a.servers.Store(server.HomepageURL, server)
	}
	return
}

//SetDownNode 设置失效节点
func (a *AppConfig) SetDownNode(host string) {
	if host == "" {
		return
	}

	if host == a.GetHost() {
		a.SetNextTryConnTime(nextTryConnectPeriod)
	}

	a.GetServers().Range(func(k, v interface{}) bool {
		server := v.(*ServerInfo)
		// if some node has down then select next node
		if strings.Index(k.(string), host) > -1 {
			server.IsDown = true
			return false
		}
		return true
	})
}

//GetServers 获取服务器数组
func (a *AppConfig) GetServers() sync.Map {
	return a.servers
}

//GetServersLen 获取服务器数组长度
func (a *AppConfig) GetServersLen() int {
	s := a.GetServers()
	l := 0
	s.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}
