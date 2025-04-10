// Copyright 2025 Apollo Authors
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

package serverlist

import (
	"encoding/json"
	"time"

	"github.com/apolloconfig/agollo/v4/component"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/server"
	"github.com/apolloconfig/agollo/v4/protocol/http"
)

const (
	// refreshIPListInterval defines the interval for refreshing server IP list
	// The client will update the server list every 20 minutes
	refreshIPListInterval = 20 * time.Minute
)

func init() {
}

// InitSyncServerIPList initializes the synchronization of server information list
// Parameters:
//   - appConfig: Function that provides application configuration
//
// This function starts a goroutine to periodically refresh the server list
func InitSyncServerIPList(appConfig func() config.AppConfig) {
	go component.StartRefreshConfig(&SyncServerIPListComponent{appConfig})
}

// SyncServerIPListComponent implements periodic server list synchronization
// It maintains a timer to update the IP list every 20 minutes
type SyncServerIPListComponent struct {
	appConfig func() config.AppConfig
}

// Start begins the server list synchronization process
// This method:
// 1. Performs initial synchronization
// 2. Sets up a timer for periodic updates
// 3. Continuously monitors for timer events
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
func SyncServerIPList(appConfigFunc func() config.AppConfig) (map[string]*config.ServerInfo, error) {
	if appConfigFunc == nil {
		panic("can not find apollo config!please confirm!")
	}

	appConfig := appConfigFunc()
	c := &env.ConnectConfig{
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	if appConfig.SyncServerTimeout > 0 {
		c.Timeout = time.Duration(appConfig.SyncServerTimeout) * time.Second
	}
	serverMap, err := http.Request(appConfig.GetServicesConfigURL(), c, &http.CallBack{
		SuccessCallBack: SyncServerIPListSuccessCallBack,
		AppConfigFunc:   appConfigFunc,
	})
	if serverMap == nil {
		return nil, err
	}

	m := serverMap.(map[string]*config.ServerInfo)
	server.SetServers(appConfig.GetHost(), m)
	return m, err
}

// SyncServerIPListSuccessCallBack handles the successful response from server list synchronization
// Parameters:
//   - responseBody: Raw response bytes from the server
//   - callback: HTTP callback handler containing context information
//
// Returns:
//   - interface{}: Map of server information processed from response
//   - error: Any error during response processing
//
// This function:
// 1. Logs the received server information
// 2. Unmarshals JSON response into server info structures
// 3. Validates the server list
// 4. Creates a map of servers indexed by homepage URL
func SyncServerIPListSuccessCallBack(responseBody []byte, callback http.CallBack) (o interface{}, err error) {
	log.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*config.ServerInfo, 0)

	err = json.Unmarshal(responseBody, &tmpServerInfo)

	if err != nil {
		log.Errorf("Unmarshal json Fail, error: %v", err)
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
