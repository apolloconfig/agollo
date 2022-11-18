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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/cluster/roundrobin"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	jsonFile "github.com/apolloconfig/agollo/v4/env/file/json"
	"github.com/apolloconfig/agollo/v4/env/server"
	"github.com/apolloconfig/agollo/v4/extension"
	http2 "github.com/apolloconfig/agollo/v4/protocol/http"
)

var asyncApollo *asyncApolloConfig

func init() {
	extension.SetLoadBalance(&roundrobin.RoundRobin{})
	extension.SetFileHandler(&jsonFile.FileHandler{})

	asyncApollo = &asyncApolloConfig{}
	asyncApollo.remoteApollo = asyncApollo
}

const configResponseStr = `{
	"appId": "100004458",
	"cluster": "default",
	"namespaceName": "application",
	"configurations": {
	  "key1":"value1",
	  "key2":"value2"
	},
	"releaseKey": "20170430092936-dee2d58e74515ff3"
}`

const grayConfigResponseStr = `{
	"appId": "100004458",
	"cluster": "default",
	"namespaceName": "application",
	"configurations": {
	  "key1":"gray_value1",
	  "key2":"gray_value2"
	},
	"releaseKey": "20170430092936-dee2d58e74515ff3"
}`

const configFilesResponseStr = `{
    "key1":"value1",
    "key2":"value2"
}`

const grayConfigFilesResponseStr = `{
    "key1":"gray_value1",
    "key2":"gray_value2"
}`

const configAbc1ResponseStr = `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "abc1",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

const responseStr = `[{"namespaceName":"application","notificationId":%d}]`
const tworesponseStr = `[{"namespaceName":"application","notificationId":%d},{"namespaceName":"abc1","notificationId":%d}]`

func onlyNormalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	label, ok := req.URL.Query()["label"]
	if ok && len(label) > 0 && label[0] == grayLabel {
		fmt.Fprintf(rw, grayConfigResponseStr)
		return
	}

	fmt.Fprintf(rw, configResponseStr)
}

func onlyNormalTwoConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configAbc1ResponseStr)
}

func serverErrorTwoConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusInternalServerError)
}

func onlynormalresponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseStr, 3)
	fmt.Fprintf(rw, "%s", result)
}

func onlynormaltworesponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(tworesponseStr, 3, 3)
	fmt.Fprintf(rw, "%s", result)
}

func initMockNotifyAndConfigServer() *httptest.Server {
	// clear
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap["abc1"] = onlyNormalTwoConfigResponse
	return runMockConfigServer(handlerMap, onlynormalresponse)
}

func initMockNotifyAndConfigServerWithTwo() *httptest.Server {
	// clear
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap["abc1"] = onlyNormalTwoConfigResponse
	return runMockConfigServer(handlerMap, onlynormaltworesponse)
}

func initMockNotifyAndConfigServerWithTwoErrResponse() *httptest.Server {
	// clear
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap["abc1"] = serverErrorTwoConfigResponse
	return runMockConfigServer(handlerMap, onlynormaltworesponse)
}

// run mock config server
func runMockConfigServer(handlerMap map[string]func(http.ResponseWriter, *http.Request),
	notifyHandler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	appConfig := env.InitFileConfig()
	uriHandlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 0)
	for namespace, handler := range handlerMap {
		uri := fmt.Sprintf("/configs/%s/%s/%s", appConfig.AppID, appConfig.Cluster, namespace)
		uriHandlerMap[uri] = handler
	}
	uriHandlerMap["/notifications/v2"] = notifyHandler

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		for path, handler := range uriHandlerMap {
			if strings.HasPrefix(uri, path) {
				handler(w, r)
				break
			}
		}
	}))

	return ts
}

func initNotifications() *config.AppConfig {
	appConfig := env.InitFileConfig()
	appConfig.NamespaceName = "application,abc1"
	appConfig.Init()
	return appConfig
}

// Error response
// will hold 5s and keep response 404
func runErrorResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	return ts
}

func TestApolloConfig_Sync(t *testing.T) {
	server := initMockNotifyAndConfigServer()
	appConfig := initNotifications()
	appConfig.IP = server.URL
	apolloConfigs := asyncApollo.Sync(func() config.AppConfig {
		return *appConfig
	})
	// err keep nil
	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(1))
	Assert(t, appConfig.GetNotificationsMap().GetNotify("application"), Equal(int64(3)))
	Assert(t, appConfig.GetNotificationsMap().GetNotify("abc1"), Equal(int64(-1)))
}

func TestApolloConfig_SyncTwoOk(t *testing.T) {
	server := initMockNotifyAndConfigServerWithTwo()
	appConfig := initNotifications()
	appConfig.IP = server.URL
	apolloConfigs := asyncApollo.Sync(func() config.AppConfig {
		return *appConfig
	})
	// err keep nil
	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(2))
	Assert(t, appConfig.GetNotificationsMap().GetNotify("application"), Equal(int64(3)))
	Assert(t, appConfig.GetNotificationsMap().GetNotify("abc1"), Equal(int64(3)))
}

func TestApolloConfig_GraySync(t *testing.T) {
	server := initMockNotifyAndConfigServer()
	appConfig := initNotifications()
	appConfig.IP = server.URL
	appConfig.Label = grayLabel
	apolloConfigs := asyncApollo.Sync(func() config.AppConfig {
		return *appConfig
	})
	// err keep nil
	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(1))

	apolloConfig := apolloConfigs[0]
	Assert(t, "gray_value1", Equal(apolloConfig.Configurations["key1"]))
	Assert(t, "gray_value2", Equal(apolloConfig.Configurations["key2"]))
}

func TestApolloConfig_SyncABC1Error(t *testing.T) {
	server := initMockNotifyAndConfigServerWithTwoErrResponse()
	appConfig := initNotifications()
	appConfig.IP = server.URL
	apolloConfigs := asyncApollo.Sync(func() config.AppConfig {
		return *appConfig
	})
	// err keep nil
	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(1))
	Assert(t, appConfig.GetNotificationsMap().GetNotify("application"), Equal(int64(3)))
	Assert(t, appConfig.GetNotificationsMap().GetNotify("abc1"), Equal(int64(-1)))
}

func TestToApolloConfigError(t *testing.T) {

	notified, err := toApolloConfig([]byte("jaskldfjaskl"))
	Assert(t, notified, NilVal())
	Assert(t, err, NotNilVal())
}

func TestGetRemoteConfig(t *testing.T) {
	server := initMockNotifyAndConfigServer()

	time.Sleep(1 * time.Second)

	var remoteConfigs []*config.Notification
	var err error
	appConfig := initNotifications()
	appConfig.IP = server.URL
	remoteConfigs, err = asyncApollo.notifyRemoteConfig(func() config.AppConfig {
		return *appConfig
	}, EMPTY)

	// err keep nil
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
	// clear
	initNotifications()
	appConfig := initNotifications()
	server1 := runErrorResponse()
	appConfig.IP = server1.URL
	server.SetNextTryConnTime(appConfig.GetHost(), 0)

	time.Sleep(1 * time.Second)

	var remoteConfigs []*config.Notification
	var err error

	remoteConfigs, err = asyncApollo.notifyRemoteConfig(func() config.AppConfig {
		return *appConfig
	}, EMPTY)

	Assert(t, err, NotNilVal())
	Assert(t, remoteConfigs, NilVal())
	Assert(t, 0, Equal(len(remoteConfigs)))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	Assert(t, "over Max Retry Still Error", Equal(err.Error()))
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
	o, err := createApolloConfigWithJSON([]byte(jsonStr), http2.CallBack{})
	c := o.(*config.ApolloConfig)

	Assert(t, err, NilVal())
	Assert(t, c, NotNilVal())

	Assert(t, "100004458", Equal(c.AppID))
	Assert(t, "default", Equal(c.Cluster))
	Assert(t, "application", Equal(c.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(c.ReleaseKey))
	Assert(t, "value1", Equal(c.Configurations["key1"]))
	Assert(t, "value2", Equal(c.Configurations["key2"]))

}

func TestCreateApolloConfigWithJsonError(t *testing.T) {
	jsonStr := `jklasdjflasjdfa`

	config, err := createApolloConfigWithJSON([]byte(jsonStr), http2.CallBack{})

	Assert(t, err, NotNilVal())
	Assert(t, config, NilVal())
}

func TestGetConfigURLSuffix(t *testing.T) {
	appConfig := &config.AppConfig{}
	appConfig.Init()
	uri := asyncApollo.GetSyncURI(*appConfig, "kk")
	Assert(t, "", NotEqual(uri))
}
