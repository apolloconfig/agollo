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

package config

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/utils"
)

var (
	appConfig = getTestAppConfig()
)

func getTestAppConfig() *AppConfig {
	jsonStr := `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "releaseKey": "1"
	}`
	c, _ := Unmarshal([]byte(jsonStr))

	return c.(*AppConfig)
}

func TestGetIsBackupConfig(t *testing.T) {
	config := appConfig.GetIsBackupConfig()
	Assert(t, config, Equal(true))
}

func TestGetBackupConfigPath(t *testing.T) {
	config := appConfig.GetBackupConfigPath()
	Assert(t, config, Equal("/app/"))
}

func TestSetNextTryConnTime(t *testing.T) {
	appConfig.SetNextTryConnTime(10)

	Assert(t, int(appConfig.NextTryConnTime), GreaterThan(int(time.Now().Unix())))
}

func Unmarshal(b []byte) (interface{}, error) {
	appConfig := &AppConfig{
		Cluster:          "default",
		NamespaceName:    "application",
		IsBackupConfig:   true,
		BackupConfigPath: "/app/",
	}
	err := json.Unmarshal(b, appConfig)
	if utils.IsNotNil(err) {
		return nil, err
	}

	return appConfig, nil
}

func TestGetHost(t *testing.T) {
	ip := appConfig.IP
	host := appConfig.GetHost()
	Assert(t, host, Equal("http://localhost:8888/"))

	appConfig.IP = "http://baidu.com"
	host = appConfig.GetHost()
	Assert(t, host, Equal("http://baidu.com/"))

	appConfig.IP = "http://163.com/"
	host = appConfig.GetHost()
	Assert(t, host, Equal("http://163.com/"))

	appConfig.IP = ip
}

func TestAppConfig_IsConnectDirectly(t *testing.T) {
	backup := appConfig.NextTryConnTime

	appConfig.NextTryConnTime = 0
	isConnectDirectly := appConfig.IsConnectDirectly()
	Assert(t, isConnectDirectly, Equal(false))

	appConfig.NextTryConnTime = time.Now().Unix() + 10
	isConnectDirectly = appConfig.IsConnectDirectly()
	Assert(t, isConnectDirectly, Equal(true))

	appConfig.NextTryConnTime = backup
}
