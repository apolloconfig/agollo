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
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v4/cluster/roundrobin"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/env/config"
	"github.com/zouyx/agollo/v4/env/config/json"
	"github.com/zouyx/agollo/v4/extension"
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
	trySyncServerIPList(appConfig)
	host := "http://localhost:8888/"
	Assert(t, host, Equal(appConfig.GetHost()))
	load := extension.GetLoadBalance().Load(*appConfig.GetServers())
	Assert(t, load, NotNilVal())
	Assert(t, host, NotEqual(load.HomepageURL))
}

func TestGetConfigURLSuffix(t *testing.T) {
	appConfig := &config.AppConfig{}
	appConfig.Init()
	uri := GetConfigURLSuffix(appConfig, "kk")
	Assert(t, "", NotEqual(uri))

	uri = GetConfigURLSuffix(nil, "kk")
	Assert(t, "", Equal(uri))
}

type testComponent struct {
}

//Start 启动同步服务器列表
func (s *testComponent) Start() {
}

func TestStartRefreshConfig(t *testing.T) {
	StartRefreshConfig(&testComponent{})
}

func TestName(t *testing.T) {

}

func trySyncServerIPList(appConfig *config.AppConfig) {
	appConfig.SyncServerIPListSuccessCallBack(appConfig, []byte(servicesConfigResponseStr))
}
