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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"

	"github.com/qshuai/agollo/v4/agcache/memory"
	"github.com/qshuai/agollo/v4/env/config"
	_ "github.com/qshuai/agollo/v4/env/file/json"
	"github.com/qshuai/agollo/v4/env/server"
	"github.com/qshuai/agollo/v4/extension"
	"github.com/qshuai/agollo/v4/storage"
)

const testDefaultNamespace = "application"

// init param
func init() {
	extension.SetCacheFactory(&memory.DefaultCacheFactory{})
}

func createMockApolloConfig(expireTime int) *internalClient {
	client := create()
	client.cache = storage.CreateNamespaceConfig(client.appConfig.GetNamespace())
	configs := make(map[string]interface{}, 0)
	// string
	configs["string"] = "value"
	// int
	configs["int"] = 1
	// float
	configs["float"] = 190.3
	// bool
	configs["bool"] = true
	// string slice
	configs["stringSlice"] = []string{"1", "2"}

	// int slice
	configs["intSlice"] = []int{1, 2}

	client.cache.UpdateApolloConfigCache(configs, expireTime, storage.GetDefaultNamespace())

	return client
}

func TestGetConfigValueNullApolloConfig(t *testing.T) {
	client := createMockApolloConfig(120)
	// test getValue
	value := client.GetValue("joe")

	Assert(t, "", Equal(value))

	// test GetStringValue
	defaultValue := "j"

	// test default
	v := client.GetStringValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

}

func TestGetIntValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := 100000

	// test default
	v := client.GetIntValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = client.GetIntValue("int", defaultValue)

	Assert(t, 1, Equal(v))

	// error type
	v = client.GetIntValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetIntSliceValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := []int{100}

	// test default
	v := client.GetIntSliceValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = client.GetIntSliceValue("intSlice", defaultValue)

	Assert(t, []int{1, 2}, Equal(v))

	// error type
	v = client.GetIntSliceValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetStringSliceValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := []string{"100"}

	// test default
	v := client.GetStringSliceValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = client.GetStringSliceValue("stringSlice", defaultValue)

	Assert(t, []string{"1", "2"}, Equal(v))

	// error type
	v = client.GetStringSliceValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetFloatValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := 100000.1

	// test default
	v := client.GetFloatValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = client.GetFloatValue("float", defaultValue)

	Assert(t, 190.3, Equal(v))

	// error type
	v = client.GetFloatValue("int", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetBoolValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := false

	// test default
	v := client.GetBoolValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = client.GetBoolValue("bool", defaultValue)

	Assert(t, true, Equal(v))

	// error type
	v = client.GetBoolValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetStringValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := "j"

	// test default
	v := client.GetStringValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = client.GetStringValue("string", defaultValue)

	Assert(t, "value", Equal(v))
}

func TestAutoSyncConfigServicesNormal2NotModified(t *testing.T) {
	client := createMockApolloConfig(120)
	serverResponse := runLongNotmodifiedConfigResponse()
	newAppConfig := getTestAppConfig()
	newAppConfig.IP = serverResponse.URL
	time.Sleep(1 * time.Second)
	server.SetServers(newAppConfig.GetHost(), nil)
	client.appConfig = newAppConfig

	apolloConfig, _ := createApolloConfigWithJSON([]byte(configResponseStr))
	client.cache.UpdateApolloConfig(apolloConfig.(*config.ApolloConfig), func() config.AppConfig {
		return *newAppConfig
	})

	config := newAppConfig.GetCurrentApolloConfig().Get()[newAppConfig.GetNamespace()]

	fmt.Println("sleeping 10s")

	time.Sleep(10 * time.Second)

	fmt.Println("checking agcache time left")
	defaultConfigCache := client.GetDefaultConfigCache()

	defaultConfigCache.Range(func(key, value interface{}) bool {
		Assert(t, value, NotNilVal())
		return true
	})

	Assert(t, config, NotNilVal())
	Assert(t, testDefaultNamespace, Equal(config.NamespaceName))
	Assert(t, "value1", Equal(client.GetStringValue("key1", "")))
	Assert(t, "value2", Equal(client.GetStringValue("key2", "")))
	checkBackupFile(client, t)
}

func createApolloConfigWithJSON(b []byte) (o interface{}, err error) {
	apolloConfig := &config.ApolloConfig{}
	apolloConfig.NamespaceName = testDefaultNamespace

	configurations := make(map[string]interface{}, 0)
	apolloConfig.Configurations = configurations
	err = json.Unmarshal(b, &apolloConfig.Configurations)
	return apolloConfig, nil
}

func checkBackupFile(client *internalClient, t *testing.T) {
	newConfig, e := extension.GetFileHandler().LoadConfigFile(client.appConfig.GetBackupConfigPath(), client.appConfig.AppID, testDefaultNamespace)
	Assert(t, newConfig, NotNilVal())
	Assert(t, e, NilVal())
	Assert(t, newConfig.Configurations, NotNilVal())
	for k, v := range newConfig.Configurations {
		Assert(t, client.GetStringValue(k, ""), Equal(v))
	}
}

func runLongNotmodifiedConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Microsecond)
		w.WriteHeader(http.StatusNotModified)
	}))

	return ts
}

func TestConfig_GetStringValue(t *testing.T) {
	client := createMockApolloConfig(120)
	config := client.GetConfig(testDefaultNamespace)

	defaultValue := "j"
	// test default
	v := config.GetStringValue("joe", defaultValue)
	Assert(t, defaultValue, Equal(v))

	// normal value
	v = config.GetStringValue("string", defaultValue)

	Assert(t, "value", Equal(v))
}

func TestConfig_GetBoolValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := false
	config := client.GetConfig(testDefaultNamespace)

	// test default
	v := config.GetBoolValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = config.GetBoolValue("bool", defaultValue)

	Assert(t, true, Equal(v))

	// error type
	v = config.GetBoolValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestConfig_GetFloatValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := 100000.1
	config := client.GetConfig(testDefaultNamespace)

	// test default
	v := config.GetFloatValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = config.GetFloatValue("float", defaultValue)

	Assert(t, 190.3, Equal(v))

	// error type
	v = config.GetFloatValue("int", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestConfig_GetIntValue(t *testing.T) {
	client := createMockApolloConfig(120)
	defaultValue := 100000
	config := client.GetConfig(testDefaultNamespace)

	// test default
	v := config.GetIntValue("joe", defaultValue)

	Assert(t, defaultValue, Equal(v))

	// normal value
	v = config.GetIntValue("int", defaultValue)

	Assert(t, 1, Equal(v))

	// error type
	v = config.GetIntValue("float", defaultValue)

	Assert(t, defaultValue, Equal(v))
}

func TestGetApolloConfigCache(t *testing.T) {
	client := createMockApolloConfig(120)
	cache := client.GetApolloConfigCache()
	Assert(t, cache, NotNilVal())
}

func TestUseEventDispatch(t *testing.T) {
	dispatch := storage.UseEventDispatch()
	cache := storage.CreateNamespaceConfig("abc")
	cache.AddChangeListener(dispatch)
	l := cache.GetChangeListeners()
	Assert(t, l.Len(), Equal(1))
}
