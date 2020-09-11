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

package notify

import (
	"encoding/json"
	"github.com/zouyx/agollo/v4/cluster/roundrobin"
	"github.com/zouyx/agollo/v4/component/log"
	"github.com/zouyx/agollo/v4/constant"
	jsonFile "github.com/zouyx/agollo/v4/env/file/json"
	"github.com/zouyx/agollo/v4/utils"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	_ "github.com/zouyx/agollo/v4/cluster/roundrobin"
	"github.com/zouyx/agollo/v4/env"
	"github.com/zouyx/agollo/v4/env/config"
	jsonConfig "github.com/zouyx/agollo/v4/env/config/json"
	_ "github.com/zouyx/agollo/v4/env/file/json"
	"github.com/zouyx/agollo/v4/extension"
)

func init() {
	extension.SetLoadBalance(&roundrobin.RoundRobin{})
	extension.SetFileHandler(&jsonFile.FileHandler{})
}

var (
	jsonConfigFile = &jsonConfig.ConfigFile{}
)

func initNotifications() *config.AppConfig {
	appConfig := env.InitFileConfig()
	appConfig.NamespaceName = "application,abc1"
	appConfig.Init()
	return appConfig
}

func TestUpdateAllNotifications(t *testing.T) {
	//clear
	c := initNotifications()

	notifyJson := `[
  {
    "namespaceName": "application",
    "notificationId": 101
  }
]`
	notifies := make([]*config.Notification, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	Assert(t, err, NilVal())
	Assert(t, true, Equal(len(notifies) > 0))

	c.GetNotificationsMap().UpdateAllNotifications(notifies)

	Assert(t, true, Equal(c.GetNotificationsMap().GetNotifyLen() > 0))
	Assert(t, int64(101), Equal(c.GetNotificationsMap().GetNotify("application")))
}

func TestUpdateAllNotificationsError(t *testing.T) {
	//clear
	appConfig := initNotifications()

	notifyJson := `ffffff`
	notifies := make([]*config.Notification, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	Assert(t, err, NotNilVal())
	Assert(t, true, Equal(len(notifies) == 0))

	appConfig.GetNotificationsMap().UpdateAllNotifications(notifies)

	Assert(t, appConfig.GetNotificationsMap().GetNotifyLen(), Equal(2))
}

func TestAutoSyncConfigServices(t *testing.T) {
	server := runNormalConfigResponse()
	newAppConfig := initNotifications()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	apolloConfigs := AutoSyncConfigServices(newAppConfig)

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
	AutoSyncConfigServices(newAppConfig)

	newAppConfig.NextTryConnTime = 0
	newAppConfig.IsBackupConfig = true
	configs := AutoSyncConfigServices(newAppConfig)

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

	apolloConfigs := AutoSyncConfigServices(newAppConfig)

	Assert(t, len(apolloConfigs), Equal(0))
}

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
