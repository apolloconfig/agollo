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
	"fmt"
	"net/http"
	"os"
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

var (
	jsonConfigFile = &jsonConfig.ConfigFile{}
	isAsync        = true
)

const responseStr = `[{"namespaceName":"application","notificationId":%d}]`

func onlyNormalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configResponseStr)
}

func onlyNormalTwoConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configAbc1ResponseStr)
}

func onlyNormalResponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseStr, 3)
	fmt.Fprintf(rw, "%s", result)
}

func initMockNotifyAndConfigServer() {
	//clear
	initNotifications()
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap["abc1"] = onlyNormalTwoConfigResponse
	server := runMockConfigServer(handlerMap, onlyNormalResponse)
	appConfig := env.InitFileConfig()
	env.InitConfig(func() (*config.AppConfig, error) {
		appConfig.IP = server.URL
		appConfig.NextTryConnTime = 0
		return appConfig, nil
	})
}

func TestSyncConfigServices(t *testing.T) {
	initMockNotifyAndConfigServer()
	appConfig := env.InitFileConfig()
	err := AsyncConfigs(appConfig)
	//err keep nil
	Assert(t, err, NilVal())
}

func TestGetRemoteConfig(t *testing.T) {
	initMockNotifyAndConfigServer()

	time.Sleep(1 * time.Second)

	var remoteConfigs []*config.Notification
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil, EMPTY, isAsync)

	//err keep nil
	Assert(t, err, NilVal())

	Assert(t, err, NilVal())
	Assert(t, remoteConfigs, NotNilVal())
	Assert(t, 1, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	notify := remoteConfigs[0]

	Assert(t, "application", Equal(notify.NamespaceName))
	Assert(t, true, Equal(notify.NotificationID > 0))
}

func TestErrorGetRemoteConfig(t *testing.T) {
	//clear
	initNotifications()
	appConfig := env.InitFileConfig()
	server := runErrorResponse()
	appConfig.IP = server.URL
	env.InitConfig(func() (*config.AppConfig, error) {
		appConfig.IP = server.URL
		appConfig.NextTryConnTime = 0
		return appConfig, nil
	})

	time.Sleep(1 * time.Second)

	var remoteConfigs []*config.Notification
	var err error
	remoteConfigs, err = notifyRemoteConfig(nil, EMPTY, isAsync)

	Assert(t, err, NotNilVal())
	Assert(t, remoteConfigs, NilVal())
	Assert(t, 0, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	Assert(t, "over Max Retry Still Error", Equal(err.Error()))
}

func initNotifications() *config.AppConfig {
	appConfig := env.InitFileConfig()
	appConfig.NamespaceName = "application,abc1"
	appConfig.InitAllNotifications(nil)
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
	appConfig := env.InitFileConfig()

	notifyJson := `ffffff`
	notifies := make([]*config.Notification, 0)

	err := json.Unmarshal([]byte(notifyJson), &notifies)

	Assert(t, err, NotNilVal())
	Assert(t, true, Equal(len(notifies) == 0))

	appConfig.GetNotificationsMap().UpdateAllNotifications(notifies)

	Assert(t, true, Equal(appConfig.GetNotificationsMap().GetNotifyLen() == 0))
}

func TestToApolloConfigError(t *testing.T) {

	notified, err := toApolloConfig([]byte("jaskldfjaskl"))
	Assert(t, notified, NilVal())
	Assert(t, err, NotNilVal())
}

func TestAutoSyncConfigServices(t *testing.T) {
	appConfig := initNotifications()
	server := runNormalConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	err := AutoSyncConfigServices(newAppConfig)
	err = AutoSyncConfigServices(newAppConfig)

	Assert(t, err, NilVal())

	config := appConfig.GetCurrentApolloConfig().Get()[newAppConfig.NamespaceName]

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
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL
	appConfig.IsBackupConfig = false
	configFilePath := extension.GetFileHandler().GetConfigFile(newAppConfig.GetBackupConfigPath(), "application")
	err := os.Remove(configFilePath)

	time.Sleep(1 * time.Second)

	appConfig.NextTryConnTime = 0

	configs := AutoSyncConfigServices(newAppConfig)

	Assert(t, err, NilVal())
	Assert(t, len(configs), GreaterThan(0))
	checkNilBackupFile(t)
	appConfig.IsBackupConfig = true
}

func checkNilBackupFile(t *testing.T) {
	appConfig := env.InitFileConfig()
	newConfig, e := extension.GetFileHandler().LoadConfigFile(appConfig.GetBackupConfigPath(), "application")
	Assert(t, e, NotNilVal())
	Assert(t, newConfig, NilVal())
}

func TestAutoSyncConfigServicesError(t *testing.T) {
	//reload app properties
	go env.InitConfig(nil)
	server := runErrorConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	err := AutoSyncConfigServices(nil)

	Assert(t, err, NotNilVal())

	config := newAppConfig.GetCurrentApolloConfig().Get()[newAppConfig.NamespaceName]

	//still properties config
	Assert(t, "100004458", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
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
	appConfig.InitAllNotifications(nil)
	appConfig.Init()
	return appConfig
}

func TestCreateApolloConfigWithJson(t *testing.T) {
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))

	Assert(t, err, NilVal())
	Assert(t, config, NotNilVal())

	Assert(t, "100004458", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
	Assert(t, "value1", Equal(config.Configurations["key1"]))
	Assert(t, "value2", Equal(config.Configurations["key2"]))

}

func TestCreateApolloConfigWithJsonError(t *testing.T) {
	jsonStr := `jklasdjflasjdfa`

	config, err := createApolloConfigWithJSON([]byte(jsonStr))

	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())
}
