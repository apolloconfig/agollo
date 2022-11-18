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

package component

import (
	json2 "encoding/json"
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/cluster/roundrobin"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/config/json"
	"github.com/apolloconfig/agollo/v4/env/server"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/protocol/http"
)

func init() {
	extension.SetLoadBalance(&roundrobin.RoundRobin{})
}

const servicesConfigResponseStr = `[{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.102:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.102:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.125:apollo-configservice:8080",
"homepageUrl": "http://10.15.88.125:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.11:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.11:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.193:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.193:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.101:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.101:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.192:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.192:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.124:apollo-configservice:8080",
"homepageUrl": "http://10.15.88.124:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.103:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.103:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "localhost:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.12:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.194:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.194:8080/"
}
]`

var (
	jsonConfigFile = &json.ConfigFile{}
)

func TestSelectOnlyOneHost(t *testing.T) {
	appConfig := env.InitFileConfig()
	trySyncServerIPList(func() config.AppConfig {
		return *appConfig
	})
	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	load := extension.GetLoadBalance().Load(server.GetServers(appConfig.GetHost()))
	Assert(t, load, NotNilVal())
	Assert(t, host, NotEqual(load.HomepageURL))

	appConfig.IP = host
	Assert(t, host, Equal(appConfig.GetHost()))
	load = extension.GetLoadBalance().Load(server.GetServers(appConfig.GetHost()))
	Assert(t, load, NotNilVal())
	Assert(t, host, NotEqual(load.HomepageURL))

	appConfig.IP = "https://localhost:8888"
	https := "https://localhost:8888/"
	Assert(t, https, Equal(appConfig.GetHost()))
	load = extension.GetLoadBalance().Load(server.GetServers(appConfig.GetHost()))
	Assert(t, load, NilVal())
}

type testComponent struct {
}

// Start 启动同步服务器列表
func (s *testComponent) Start() {
}

func TestStartRefreshConfig(t *testing.T) {
	StartRefreshConfig(&testComponent{})
}

func TestName(t *testing.T) {

}

func trySyncServerIPList(appConfigFunc func() config.AppConfig) {
	SyncServerIPListSuccessCallBack([]byte(servicesConfigResponseStr), http.CallBack{AppConfigFunc: appConfigFunc})
}

// SyncServerIPListSuccessCallBack 同步服务器列表成功后的回调
func SyncServerIPListSuccessCallBack(responseBody []byte, callback http.CallBack) (o interface{}, err error) {
	log.Debug("get all server info:", string(responseBody))

	tmpServerInfo := make([]*config.ServerInfo, 0)

	err = json2.Unmarshal(responseBody, &tmpServerInfo)

	if err != nil {
		log.Errorf("Unmarshal json Fail,Error: %s", err)
		return
	}

	if len(tmpServerInfo) == 0 {
		log.Info("get no real server!")
		return
	}

	serverMap := make(map[string]*config.ServerInfo)
	for _, server1 := range tmpServerInfo {
		if server1 == nil {
			continue
		}
		serverMap[server1.HomepageURL] = server1
	}
	configFunc := callback.AppConfigFunc()
	c := &configFunc
	server.SetServers(c.GetHost(), serverMap)
	return
}
