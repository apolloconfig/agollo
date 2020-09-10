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
	"github.com/zouyx/agollo/v4/cluster/roundrobin"
	jsonFile "github.com/zouyx/agollo/v4/env/file/json"
	"net/http"
	"net/http/httptest"
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

func init() {
	extension.SetLoadBalance(&roundrobin.RoundRobin{})
	extension.SetFileHandler(&jsonFile.FileHandler{})
}

var (
	jsonConfigFile = &jsonConfig.ConfigFile{}
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

func initMockNotifyAndConfigServer() *httptest.Server {
	//clear
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap["abc1"] = onlyNormalTwoConfigResponse
	return runMockConfigServer(handlerMap, onlyNormalResponse)
}

func TestSyncConfigServices(t *testing.T) {
	server := initMockNotifyAndConfigServer()
	appConfig := initNotifications()
	appConfig.IP = server.URL
	apolloConfigs := AsyncConfigs(appConfig)
	//err keep nil
	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(2))
}

func TestGetRemoteConfig(t *testing.T) {
	server := initMockNotifyAndConfigServer()

	time.Sleep(1 * time.Second)

	var remoteConfigs []*config.Notification
	var err error
	appConfig := initNotifications()
	appConfig.IP = server.URL
	remoteConfigs, err = notifyRemoteConfig(appConfig, EMPTY)

	//err keep nil
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
	appConfig := initNotifications()
	server := runErrorResponse()
	appConfig.IP = server.URL
	appConfig.NextTryConnTime = 0

	time.Sleep(1 * time.Second)

	var remoteConfigs []*config.Notification
	var err error
	remoteConfigs, err = notifyRemoteConfig(appConfig, EMPTY)

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

func TestToApolloConfigError(t *testing.T) {

	notified, err := toApolloConfig([]byte("jaskldfjaskl"))
	Assert(t, notified, NilVal())
	Assert(t, err, NotNilVal())
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
