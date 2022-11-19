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

package serverlist

import (
	"encoding/json"
	"time"

	"github.com/qshuai/agollo/v4/component"
	"github.com/qshuai/agollo/v4/component/log"
	"github.com/qshuai/agollo/v4/env"
	"github.com/qshuai/agollo/v4/env/config"
	"github.com/qshuai/agollo/v4/env/server"
	"github.com/qshuai/agollo/v4/protocol/http"
)

const (
	// refresh ip list
	refreshIPListInterval = 20 * time.Minute // 20m
)

// InitSyncServerIPList 初始化同步服务器信息列表
func InitSyncServerIPList(appConfig func() config.AppConfig) {
	go component.StartRefreshConfig(&SyncServerIPListComponent{appConfig})
}

// SyncServerIPListComponent set timer for update ip list
// interval : 20m
type SyncServerIPListComponent struct {
	appConfig func() config.AppConfig
}

// Start 启动同步服务器列表
func (s *SyncServerIPListComponent) Start() {
	SyncServerIPList(s.appConfig)
	log.Debug("syncServerIpList started")

	t2 := time.NewTimer(refreshIPListInterval)
	for {
		select {
		case <-t2.C:
			SyncServerIPList(s.appConfig)
			t2.Reset(refreshIPListInterval)
		}
	}
}

// SyncServerIPList sync ip list from server
// then
// 1.update agcache
// 2.store in disk
func SyncServerIPList(appConfigFn func() config.AppConfig) (map[string]*config.ServerInfo, error) {
	if appConfigFn == nil {
		panic("can not find apollo config!please confirm!")
	}

	appConfig := appConfigFn()
	c := &env.ConnectConfig{
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	if appConfig.SyncServerTimeout > 0 {
		c.Timeout = time.Duration(appConfig.SyncServerTimeout) * time.Second
	}
	serverMap, err := http.Request(appConfig.GetServicesConfigURL(), c, &http.CallBack{
		SuccessCallBack: SyncServerIPListSuccessCallBack,
		AppConfigFunc:   appConfigFn,
	})
	if serverMap == nil {
		return nil, err
	}

	m := serverMap.(map[string]*config.ServerInfo)
	server.SetServers(appConfig.GetHost(), m)
	return m, err
}

// SyncServerIPListSuccessCallBack 同步服务器列表成功后的回调
func SyncServerIPListSuccessCallBack(responseBody []byte, callback http.CallBack) (o interface{}, err error) {
	log.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*config.ServerInfo, 0)
	err = json.Unmarshal(responseBody, &tmpServerInfo)
	if err != nil {
		log.Error("Unmarshal json Fail,Error: %v", err)
		return
	}

	if len(tmpServerInfo) == 0 {
		log.Info("get no real server!")
		return
	}

	m := make(map[string]*config.ServerInfo)
	for _, server := range tmpServerInfo {
		if server == nil {
			continue
		}
		m[server.HomepageURL] = server
	}
	o = m
	return
}
