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

package http

import (
	"fmt"
	"github.com/zouyx/agollo/v4/cluster/roundrobin"
	"github.com/zouyx/agollo/v4/extension"
	"net/url"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/env/config"
	"github.com/zouyx/agollo/v4/env/config/json"
	"github.com/zouyx/agollo/v4/utils"
)

func init() {
	extension.SetLoadBalance(&roundrobin.RoundRobin{})
}

var (
	jsonConfigFile = &json.ConfigFile{}
)

func getTestAppConfig() *config.AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`

	c, _ := env.Unmarshal([]byte(jsonStr))
	appConfig := c.(*config.AppConfig)
	appConfig.Init()
	return appConfig
}

func TestRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	server := runNormalBackupConfigResponse()
	appConfig := getTestAppConfig()
	appConfig.IP = server.URL

	mockIPList(t, appConfig)
	urlSuffix := getConfigURLSuffix(appConfig, appConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		URI:     urlSuffix,
		IsRetry: true,
	}, &CallBack{
		SuccessCallBack: nil,
	})

	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func TestHttpsRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	server := runNormalBackupConfigResponseWithHTTPS()
	appConfig := getTestAppConfig()
	appConfig.IP = server.URL

	mockIPList(t, appConfig)
	urlSuffix := getConfigURLSuffix(appConfig, appConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		URI:     urlSuffix,
		IsRetry: true,
	}, &CallBack{
		SuccessCallBack: nil,
	})

	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func TestCustomTimeout(t *testing.T) {
	time.Sleep(1 * time.Second)
	server := runLongTimeResponse()
	appConfig := getTestAppConfig()
	appConfig.IP = server.URL

	startTime := time.Now().Unix()
	mockIPList(t, appConfig)
	urlSuffix := getConfigURLSuffix(appConfig, appConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		URI:     urlSuffix,
		Timeout: 11 * time.Second,
	}, &CallBack{
		SuccessCallBack: nil,
	})

	endTime := time.Now().Unix()
	duration := endTime - startTime
	t.Log("start time:", startTime)
	t.Log("endTime:", endTime)
	t.Log("duration:", duration)
	Assert(t, int64(11), Equal(duration))
	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func mockIPList(t *testing.T, appConfig *config.AppConfig) {
	time.Sleep(1 * time.Second)

	_, err := appConfig.SyncServerIPListSuccessCallBack(nil, []byte(servicesResponseStr))

	Assert(t, err, NilVal())

	serverLen := appConfig.GetServersLen()

	Assert(t, 2, Equal(serverLen))
}

func getConfigURLSuffix(config *config.AppConfig, namespaceName string) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(config.GetCurrentApolloConfig().GetReleaseKey(namespaceName)),
		utils.GetInternal())
}
