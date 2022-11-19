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
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/qshuai/agollo/v4/env"
	"github.com/qshuai/agollo/v4/env/config"
	"github.com/qshuai/agollo/v4/env/server"
	"github.com/qshuai/agollo/v4/protocol/http"
)

func TestSyncServerIPList(t *testing.T) {
	trySyncServerIPList(t)
}

func trySyncServerIPList(t *testing.T) {
	server := runMockServicesConfigServer()
	defer server.Close()

	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL
	serverMap, err := SyncServerIPList(func() config.AppConfig {
		return *newAppConfig
	})

	Assert(t, err, NilVal())

	Assert(t, 10, Equal(len(serverMap)))

}

func getTestAppConfig() *config.AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
	"syncServerTimeout":2,
    "releaseKey": "1"
	}`
	c, _ := env.Unmarshal([]byte(jsonStr))

	return c.(*config.AppConfig)
}

func TestSyncServerIpListSuccessCallBack(t *testing.T) {
	appConfig := getTestAppConfig()
	serverMap, _ := SyncServerIPListSuccessCallBack([]byte(servicesConfigResponseStr), http.CallBack{AppConfigFunc: func() config.AppConfig {
		return *appConfig
	}})
	m := serverMap.(map[string]*config.ServerInfo)
	Assert(t, len(m), Equal(10))
}

func TestSetDownNode(t *testing.T) {
	t.SkipNow()
	appConfig := getTestAppConfig()
	SyncServerIPListSuccessCallBack([]byte(servicesConfigResponseStr), http.CallBack{AppConfigFunc: func() config.AppConfig {
		return *appConfig
	}})

	downNode := "10.15.128.102:8080"
	server.SetDownNode(appConfig.GetHost(), downNode)

	info, ok := server.GetServers(appConfig.IP)["http://10.15.128.102:8080/"]
	Assert(t, ok, Equal(true))
	Assert(t, info.IsDown, Equal(true))
}
