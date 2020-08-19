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
	"net/url"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/env/config"
	"github.com/zouyx/agollo/v4/env/config/json"
	"github.com/zouyx/agollo/v4/utils"
)

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

	return c.(*config.AppConfig)
}

func TestRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIPList(t)
	server := runNormalBackupConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := getConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		URI: urlSuffix,
	}, &CallBack{
		SuccessCallBack: nil,
	})

	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func TestHttpsRequestRecovery(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIPList(t)
	server := runNormalBackupConfigResponseWithHTTPS()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := getConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

	o, err := RequestRecovery(appConfig, &env.ConnectConfig{
		URI: urlSuffix,
	}, &CallBack{
		SuccessCallBack: nil,
	})

	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func TestCustomTimeout(t *testing.T) {
	time.Sleep(1 * time.Second)
	mockIPList(t)
	server := runLongTimeResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	startTime := time.Now().Unix()
	appConfig := env.GetAppConfig(newAppConfig)
	urlSuffix := getConfigURLSuffix(appConfig, newAppConfig.NamespaceName)

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
	Assert(t, int64(10), Equal(duration))
	Assert(t, err, NilVal())
	Assert(t, o, NilVal())
}

func mockIPList(t *testing.T) {
	time.Sleep(1 * time.Second)

	_, err := env.SyncServerIPListSuccessCallBack([]byte(servicesResponseStr))

	Assert(t, err, NilVal())

	serverLen := env.GetServersLen()

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
		url.QueryEscape(env.GetCurrentApolloConfigReleaseKey(namespaceName)),
		utils.GetInternal())
}
