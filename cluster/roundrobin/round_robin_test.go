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

package roundrobin

import (
	"testing"

	"github.com/apolloconfig/agollo/v4/component/serverlist"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/server"
	"github.com/apolloconfig/agollo/v4/protocol/http"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/env"
)

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

func TestSelectHost(t *testing.T) {
	balanace := &RoundRobin{}

	appConfig := env.InitFileConfig()
	// mock ip data
	trySyncServerIPList(*appConfig)

	t.Log("appconfig host:" + appConfig.GetHost())
	t.Log("appconfig select host:", balanace.Load(server.GetServers(appConfig.GetHost())).HomepageURL)

	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	Assert(t, host, NotEqual(balanace.Load(server.GetServers(appConfig.GetHost())).HomepageURL))

	// check select next time
	server.SetNextTryConnTime(appConfig.GetHost(), 5)
	Assert(t, host, NotEqual(balanace.Load(server.GetServers(appConfig.GetHost())).HomepageURL))

	// check servers
	server.SetNextTryConnTime(appConfig.GetHost(), 5)
	firstHost := balanace.Load(server.GetServers(appConfig.GetHost())).HomepageURL
	Assert(t, host, NotEqual(firstHost))
	server.SetDownNode(appConfig.GetHost(), firstHost)

	secondHost := balanace.Load(server.GetServers(appConfig.GetHost())).HomepageURL
	Assert(t, host, NotEqual(secondHost))
	Assert(t, firstHost, NotEqual(secondHost))
	server.SetDownNode(appConfig.GetHost(), secondHost)

	thirdHost := balanace.Load(server.GetServers(appConfig.GetHost())).HomepageURL
	Assert(t, host, NotEqual(thirdHost))
	Assert(t, firstHost, NotEqual(thirdHost))
	Assert(t, secondHost, NotEqual(thirdHost))

	for _, info := range server.GetServers(appConfig.GetHost()) {
		info.IsDown = true
	}

	Assert(t, balanace.Load(server.GetServers(appConfig.GetHost())), NilVal())

	// no servers
	// servers = make(map[string]*serverInfo, 0)
	deleteServers(appConfig)
	Assert(t, balanace.Load(server.GetServers(appConfig.GetHost())), NilVal())
}

func deleteServers(appConfig *config.AppConfig) {
	servers := make(map[string]*config.ServerInfo)
	server.SetServers(appConfig.GetHost(), servers)
}

func trySyncServerIPList(appConfig config.AppConfig) {
	// 里面已经设置了
	serverMap, _ := serverlist.SyncServerIPListSuccessCallBack([]byte(servicesConfigResponseStr), http.CallBack{AppConfigFunc: func() config.AppConfig {
		return appConfig
	}})
	m := serverMap.(map[string]*config.ServerInfo)
	server.SetServers(appConfig.GetHost(), m)
}
