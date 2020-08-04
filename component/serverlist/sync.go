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
	"time"

	"github.com/zouyx/agollo/v3/component"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/protocol/http"
)

const (
	//refresh ip list
	refreshIPListInterval = 20 * time.Minute //20m
)

func init() {

}

//InitSyncServerIPList 初始化同步服务器信息列表
func InitSyncServerIPList() {
	go component.StartRefreshConfig(&SyncServerIPListComponent{})
}

//SyncServerIPListComponent set timer for update ip list
//interval : 20m
type SyncServerIPListComponent struct {
}

//Start 启动同步服务器列表
func (s *SyncServerIPListComponent) Start() {
	SyncServerIPList(nil)
	log.Debug("syncServerIpList started")

	t2 := time.NewTimer(refreshIPListInterval)
	for {
		select {
		case <-t2.C:
			SyncServerIPList(nil)
			t2.Reset(refreshIPListInterval)
		}
	}
}

//SyncServerIPList sync ip list from server
//then
//1.update agcache
//2.store in disk
func SyncServerIPList(newAppConfig *config.AppConfig) error {
	appConfig := env.GetAppConfig(newAppConfig)
	if appConfig == nil {
		panic("can not find apollo config!please confirm!")
	}

	_, err := http.Request(env.GetServicesConfigURL(appConfig), &env.ConnectConfig{
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}, &http.CallBack{
		SuccessCallBack: env.SyncServerIPListSuccessCallBack,
	})

	return err
}
