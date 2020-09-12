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

package remote

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/extension"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	normalConfigCount = 1
	syncApollo        *syncApolloConfig
)

func init() {
	syncApollo = &syncApolloConfig{}
	syncApollo.remoteApollo = syncApollo
}

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func runNormalConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalConfigCount++
		if normalConfigCount%2 == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(configResponseStr))
		} else {
			time.Sleep(500 * time.Microsecond)
			w.WriteHeader(http.StatusNotModified)
		}
	}))

	return ts
}

//Error response
//will hold 5s and keep response 404
func runErrorConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Microsecond)
		w.WriteHeader(http.StatusNotFound)
	}))

	return ts
}

func TestAutoSyncConfigServices(t *testing.T) {
	server := runNormalConfigResponse()
	newAppConfig := initNotifications()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	apolloConfigs := syncApollo.Sync(newAppConfig)

	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(1))

	newAppConfig.GetCurrentApolloConfig().Set(newAppConfig.NamespaceName, &apolloConfigs[0].ApolloConnConfig)
	config := newAppConfig.GetCurrentApolloConfig().Get()[newAppConfig.NamespaceName]

	Assert(t, "100004458", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	//Assert(t,"value1",config.Configurations["key1"])
	//Assert(t,"value2",config.Configurations["key2"])
}

func TestAutoSyncConfigServicesNoBackupFile(t *testing.T) {
	appConfig := initNotifications()
	server := runNormalConfigResponse()
	newAppConfig := initNotifications()
	newAppConfig.IP = server.URL
	appConfig.IsBackupConfig = false
	configFilePath := extension.GetFileHandler().GetConfigFile(newAppConfig.GetBackupConfigPath(), "application")
	os.Remove(configFilePath)

	time.Sleep(1 * time.Second)

	newAppConfig.NextTryConnTime = 0
	newAppConfig.IsBackupConfig = false
	syncApollo.Sync(newAppConfig)

	newAppConfig.NextTryConnTime = 0
	newAppConfig.IsBackupConfig = true
	configs := syncApollo.Sync(newAppConfig)

	Assert(t, len(configs), GreaterThan(0))
	checkNilBackupFile(t)

}

func checkNilBackupFile(t *testing.T) {
	appConfig := env.InitFileConfig()
	newConfig, e := extension.GetFileHandler().LoadConfigFile(appConfig.GetBackupConfigPath(), "application")
	Assert(t, e, NotNilVal())
	Assert(t, newConfig, NilVal())
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	server := runErrorConfigResponse()
	newAppConfig := initNotifications()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	apolloConfigs := syncApollo.Sync(newAppConfig)

	Assert(t, len(apolloConfigs), Equal(0))
}
