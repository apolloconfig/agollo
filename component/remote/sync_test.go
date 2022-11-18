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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/constant"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/server"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/utils/parse/normal"
	"github.com/apolloconfig/agollo/v4/utils/parse/properties"
	"github.com/apolloconfig/agollo/v4/utils/parse/yaml"
	"github.com/apolloconfig/agollo/v4/utils/parse/yml"
)

var (
	grayLabel         = "gray"
	normalConfigCount = 1
	syncApollo        *syncApolloConfig
)

func init() {
	syncApollo = &syncApolloConfig{}
	syncApollo.remoteApollo = syncApollo

	// file parser
	extension.AddFormatParser(constant.DEFAULT, &normal.Parser{})
	extension.AddFormatParser(constant.Properties, &properties.Parser{})
	extension.AddFormatParser(constant.YML, &yml.Parser{})
	extension.AddFormatParser(constant.YAML, &yaml.Parser{})
}

// Normal response
// First request will hold 5s and response http.StatusNotModified
// Second request will hold 5s and response http.StatusNotModified
// Second request will response [{"namespaceName":"application","notificationId":3}]
func runNormalConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalConfigCount++
		if normalConfigCount%2 == 0 {
			label, ok := r.URL.Query()["label"]
			if ok && len(label) > 0 && label[0] == grayLabel {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(grayConfigFilesResponseStr))

				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(configFilesResponseStr))
		} else {
			time.Sleep(500 * time.Microsecond)
			w.WriteHeader(http.StatusNotModified)
		}
	}))

	return ts
}

// Error response
// will hold 5s and keep response 404
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

	apolloConfigs := syncApollo.Sync(func() config.AppConfig {
		return *newAppConfig
	})

	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(1))

	apolloConfig := apolloConfigs[0]
	newAppConfig.GetCurrentApolloConfig().Set(newAppConfig.NamespaceName, &apolloConfig.ApolloConnConfig)
	c := newAppConfig.GetCurrentApolloConfig().Get()[newAppConfig.NamespaceName]
	Assert(t, "application", Equal(c.NamespaceName))
	Assert(t, "value1", Equal(apolloConfig.Configurations["key1"]))
	Assert(t, "value2", Equal(apolloConfig.Configurations["key2"]))
}

func TestAutoSyncConfigServicesNoBackupFile(t *testing.T) {
	appConfig := initNotifications()
	server1 := runNormalConfigResponse()
	newAppConfig := initNotifications()
	newAppConfig.IP = server1.URL
	appConfig.IsBackupConfig = false
	configFilePath := extension.GetFileHandler().GetConfigFile(newAppConfig.GetBackupConfigPath(), newAppConfig.AppID, "application")
	os.Remove(configFilePath)

	time.Sleep(1 * time.Second)

	server.SetNextTryConnTime(appConfig.GetHost(), 0)
	newAppConfig.IsBackupConfig = false
	syncApollo.Sync(func() config.AppConfig {
		return *newAppConfig
	})

	server.SetNextTryConnTime(appConfig.GetHost(), 0)
	newAppConfig.IsBackupConfig = true
	configs := syncApollo.Sync(func() config.AppConfig {
		return *newAppConfig
	})

	Assert(t, len(configs), GreaterThan(0))
	checkNilBackupFile(t)

}

func checkNilBackupFile(t *testing.T) {
	appConfig := env.InitFileConfig()
	newConfig, e := extension.GetFileHandler().LoadConfigFile(appConfig.GetBackupConfigPath(), appConfig.AppID, "application")
	Assert(t, e, NotNilVal())
	Assert(t, newConfig, NilVal())
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	// reload app properties
	server := runErrorConfigResponse()
	newAppConfig := initNotifications()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	apolloConfigs := syncApollo.Sync(func() config.AppConfig {
		return *newAppConfig
	})

	Assert(t, len(apolloConfigs), Equal(0))
}

func TestClientLabelConfigService(t *testing.T) {
	server := runNormalConfigResponse()
	newAppConfig := initNotifications()
	newAppConfig.IP = server.URL
	newAppConfig.Label = grayLabel

	apolloConfigs := syncApollo.Sync(func() config.AppConfig {
		return *newAppConfig
	})

	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(1))

	apolloConfig := apolloConfigs[0]
	newAppConfig.GetCurrentApolloConfig().Set(newAppConfig.NamespaceName, &apolloConfig.ApolloConnConfig)
	c := newAppConfig.GetCurrentApolloConfig().Get()[newAppConfig.NamespaceName]
	Assert(t, "application", Equal(c.NamespaceName))
	Assert(t, "gray_value1", Equal(apolloConfig.Configurations["key1"]))
	Assert(t, "gray_value2", Equal(apolloConfig.Configurations["key2"]))
}
