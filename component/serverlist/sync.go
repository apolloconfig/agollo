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
	"fmt"
	"strconv"
	"time"

	"github.com/apolloconfig/agollo/v4/component"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/server"
	"github.com/apolloconfig/agollo/v4/perror"
	"github.com/apolloconfig/agollo/v4/protocol/http"
)

const (
	// refresh ip list
	refreshIPListInterval = 20 * time.Minute // 20m
)

// InitSyncServerIPList 初始化同步服务器信息列表
func InitSyncServerIPList(appConfig func() config.AppConfig) error {
	// 先同步执行一次，后续定时异步执行
	if _, err := SyncServerIPList(appConfig); err != nil {
		return err
	}
	if err := CheckSecretOK(appConfig); err != nil {
		return err
	}
	go component.StartRefreshConfig(&SyncServerIPListComponent{
		appConfig: appConfig,
		stopCh:    make(chan struct{}),
	})
	return nil
}

// SyncServerIPListComponent set timer for update ip list
// interval : 20m
type SyncServerIPListComponent struct {
	appConfig func() config.AppConfig
	stopCh    chan struct{}
}

// Start 启动向apollo服务列表同步
func (s *SyncServerIPListComponent) Start() {
	if s.stopCh == nil {
		s.stopCh = make(chan struct{})
	}

	t2 := time.NewTimer(refreshIPListInterval)
loop:
	for {
		select {
		case <-t2.C:
			if _, err := SyncServerIPList(s.appConfig); err != nil {
				log.Errorf("同步Apollo服务信息失败. err: %+v", err)
			}
			t2.Reset(refreshIPListInterval)
		case <-s.stopCh:
			break loop
		}
	}
}

// Stop 停止向apollo服务长轮询
func (s *SyncServerIPListComponent) Stop() {
	if s.stopCh != nil {
		close(s.stopCh)
	}
}

// SyncServerIPList 同步apollo服务信息
func SyncServerIPList(appConfigFunc func() config.AppConfig) (map[string]*config.ServerInfo, error) {
	if appConfigFunc == nil {
		return nil, fmt.Errorf("can not find apollo config! please confirm")
	}

	appConfig := appConfigFunc()
	c := &env.ConnectConfig{
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	if appConfigFunc().SyncServerTimeout > 0 {
		duration, err := time.ParseDuration(strconv.Itoa(appConfigFunc().SyncServerTimeout) + "s")
		if err != nil {
			return nil, err
		}
		c.Timeout = duration
	}
	serverMap, err := http.Request(appConfig.GetServicesConfigURL(), c, &http.CallBack{
		SuccessCallBack: SyncServerIPListSuccessCallBack,
		AppConfigFunc:   appConfigFunc,
	})
	if err != nil {
		if err == perror.ErrOverMaxRetryStill {
			return nil, fmt.Errorf("获取Apollo服务列表失败")
		}
		return nil, err
	}

	m := serverMap.(map[string]*config.ServerInfo)
	server.SetServers(appConfig.GetHost(), m)
	return m, err
}

// SyncServerIPListSuccessCallBack 同步服务器列表成功后的回调
func SyncServerIPListSuccessCallBack(responseBody []byte, callback http.CallBack) (serversInfoMap interface{}, err error) {
	log.Debugf("get all server info: %s", string(responseBody))

	tmpServerInfo := make([]*config.ServerInfo, 0)
	if err = json.Unmarshal(responseBody, &tmpServerInfo); err != nil {
		log.Errorf("unmarshal json failed. err: %v", err)
		return
	}
	if len(tmpServerInfo) == 0 {
		log.Info("get no real server!")
		return
	}

	m := make(map[string]*config.ServerInfo, len(tmpServerInfo))
	for _, svr := range tmpServerInfo {
		if svr == nil {
			continue
		}
		m[svr.HomepageURL] = svr
	}
	serversInfoMap = m
	return
}
