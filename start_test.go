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

package agollo

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/agcache/memory"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	jsonFile "github.com/apolloconfig/agollo/v4/env/config/json"
	"github.com/apolloconfig/agollo/v4/extension"
)

var (
	jsonConfigFile = &jsonFile.ConfigFile{}
	appConfigFile  = `{
    "appId": "test",
    "cluster": "dev",
    "namespaceName": "application",
    "ip": "localhost:8888",
    "backupConfigPath":""
}`
	appConfig = &config.AppConfig{
		AppID:         "test",
		Cluster:       "dev",
		NamespaceName: "application",
		IP:            "localhost:8888",
	}
)

func writeFile(content []byte, configPath string) {
	file, e := os.Create(configPath)
	if e != nil {
		log.Errorf("writeConfigFile fail,error:", e)
	}
	defer file.Close()
	file.Write(content)
}

func TestStart(t *testing.T) {
	c := appConfig
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	server := runMockConfigFilesServer(handlerMap, nil, c)
	c.IP = server.URL

	b, _ := json.Marshal(c)
	writeFile(b, "app.properties")

	client, _ := Start()

	value := client.GetValue("key1")
	Assert(t, "value1", Equal(value))
	handler := extension.GetFileHandler()
	Assert(t, handler, NotNilVal())
}

func TestStartWithMultiNamespace(t *testing.T) {
	c := appConfig
	app1 := "abc1"
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	handlerMap[app1] = onlyNormalSecondConfigResponse
	server := runMockConfigFilesServer(handlerMap, nil, appConfig)

	c.NamespaceName = "application,abc1"
	c.IP = server.URL
	b, _ := json.Marshal(c)
	writeFile(b, "app.properties")

	client, _ := Start()

	value := client.GetValue("key1")
	Assert(t, "value1", Equal(value))

	config := client.GetConfig(app1)
	Assert(t, config, NotNilVal())
	Assert(t, config.GetValue("key1-1"), Equal("value1-1"))

	rollbackFile()
}

func rollbackFile() {
	writeFile([]byte(appConfigFile), "app.properties")
}

func TestErrorStart(t *testing.T) {
	t.SkipNow()
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	time.Sleep(1 * time.Second)

	client, _ := Start()

	value := client.GetValue("key1")
	Assert(t, "value1", Equal(value))

	value2 := client.GetValue("key2")
	Assert(t, "value2", Equal(value2))
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

	c2 := c.(*config.AppConfig)
	c2.Init()
	return c2
}

func TestStructInit(t *testing.T) {
	t.SkipNow()
	readyConfig := &config.AppConfig{
		AppID:         "test1",
		Cluster:       "dev1",
		NamespaceName: "application1",
		IP:            "localhost:8889",
	}

	client, _ := StartWithConfig(func() (*config.AppConfig, error) {
		return readyConfig, nil
	})

	time.Sleep(1 * time.Second)

	c := client.(*internalClient).appConfig
	Assert(t, c, NotNilVal())
	Assert(t, "test1", Equal(c.AppID))
	Assert(t, "dev1", Equal(c.Cluster))
	Assert(t, "application1", Equal(c.NamespaceName))
	Assert(t, "localhost:8889", Equal(c.IP))

	apolloConfig := c.GetCurrentApolloConfig().Get()[c.NamespaceName]
	Assert(t, "test1", Equal(apolloConfig.AppID))
	Assert(t, "dev1", Equal(apolloConfig.Cluster))
	Assert(t, "application1", Equal(apolloConfig.NamespaceName))

	// revert file config
	env.InitFileConfig()
}

func TestSetLogger(t *testing.T) {
	// TODO log.Logger data race
	// logger := &log.DefaultLogger{}
	// SetLogger(logger)
	// Assert(t, log.Logger, Equal(logger))
}

func TestSetCache(t *testing.T) {
	defaultCacheFactory := &memory.DefaultCacheFactory{}
	SetCache(defaultCacheFactory)
	Assert(t, extension.GetCacheFactory(), Equal(defaultCacheFactory))
}

type TestLoadBalance struct{}

// Load 负载均衡
func (r *TestLoadBalance) Load(servers map[string]*config.ServerInfo) *config.ServerInfo {
	return nil
}

func TestSetLoadBalance(t *testing.T) {
	balance := extension.GetLoadBalance()
	Assert(t, balance, NotNilVal())

	t2 := &TestLoadBalance{}
	SetLoadBalance(t2)
	Assert(t, t2, Equal(extension.GetLoadBalance()))
}

// testFileHandler 默认备份文件读写
type testFileHandler struct{}

// WriteConfigFile write config to file
func (fileHandler *testFileHandler) WriteConfigFile(config *config.ApolloConfig, configPath string) error {
	return nil
}

// GetConfigFile get real config file
func (fileHandler *testFileHandler) GetConfigFile(configDir string, appID string, namespace string) string {
	return ""
}

// LoadConfigFile load config from file
func (fileHandler *testFileHandler) LoadConfigFile(configDir string, appID string, namespace string) (*config.ApolloConfig, error) {
	return nil, nil
}

func TestSetBackupFileHandler(t *testing.T) {
	fileHandler := extension.GetFileHandler()
	Assert(t, fileHandler, NotNilVal())

	t2 := &testFileHandler{}
	SetBackupFileHandler(t2)
	Assert(t, t2, Equal(extension.GetFileHandler()))
}

type TestAuth struct{}

func (a *TestAuth) HTTPHeaders(url string, appID string, secret string) map[string][]string {
	return nil
}

func TestSetSignature(t *testing.T) {
	Assert(t, extension.GetHTTPAuth(), NotNilVal())

	t2 := &TestAuth{}
	SetSignature(t2)

	Assert(t, t2, Equal(extension.GetHTTPAuth()))
}

func TestErrorStartWithConfigMustReadFromRemote(t *testing.T) {
	server := runErrorResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL
	newAppConfig.MustStart = true

	client, err := StartWithConfig(func() (*config.AppConfig, error) {
		return newAppConfig, nil
	})

	Assert(t, client, Equal(nil))
	Assert(t, err, NotNilVal())
}

func TestStartWithConfigMustReadFromRemote(t *testing.T) {
	c := appConfig
	handlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 1)
	handlerMap["application"] = onlyNormalConfigResponse
	server := runMockConfigFilesServer(handlerMap, nil, c)
	c.IP = server.URL
	c.MustStart = true

	client, err := StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	Assert(t, err, Equal(nil))

	value := client.GetValue("key1")
	Assert(t, "value1", Equal(value))
	handler := extension.GetFileHandler()
	Assert(t, handler, NotNilVal())
}
