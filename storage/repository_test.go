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

package storage

import (
	"github.com/zouyx/agollo/v4/agcache/memory"
	"github.com/zouyx/agollo/v4/env/config"
	jsonFile "github.com/zouyx/agollo/v4/env/file/json"
	"github.com/zouyx/agollo/v4/extension"
	"strings"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"
	_ "github.com/zouyx/agollo/v4/agcache/memory"
	"github.com/zouyx/agollo/v4/env"
	_ "github.com/zouyx/agollo/v4/env/file/json"

	_ "github.com/zouyx/agollo/v4/utils/parse/normal"
	_ "github.com/zouyx/agollo/v4/utils/parse/properties"
)

//init param
func init() {
	extension.SetCacheFactory(&memory.DefaultCacheFactory{})
	extension.SetFileHandler(&jsonFile.FileHandler{})
}

func creatTestApolloConfig(configurations map[string]interface{}, namespace string) *Cache {
	c := CreateNamespaceConfig(namespace)
	appConfig := env.InitFileConfig()
	apolloConfig := &config.ApolloConfig{}
	apolloConfig.NamespaceName = namespace
	apolloConfig.AppID = "test"
	apolloConfig.Cluster = "dev"
	apolloConfig.Configurations = configurations
	c.UpdateApolloConfig(apolloConfig, func() config.AppConfig {
		return *appConfig
	}, true)
	return c

}

func TestUpdateApolloConfigNull(t *testing.T) {
	time.Sleep(1 * time.Second)
	c := CreateNamespaceConfig(defaultNamespace)
	appConfig := env.InitFileConfig()

	configurations := make(map[string]interface{})
	configurations["string"] = "string"
	configurations["int"] = "1"
	configurations["float"] = "1"
	configurations["bool"] = "true"
	configurations["slice"] = []int{1, 2}

	apolloConfig := &config.ApolloConfig{}
	apolloConfig.NamespaceName = defaultNamespace
	apolloConfig.AppID = "test"
	apolloConfig.Cluster = "dev"
	apolloConfig.Configurations = configurations
	c.UpdateApolloConfig(apolloConfig, func() config.AppConfig {
		return *appConfig
	}, true)

	currentConnApolloConfig := appConfig.GetCurrentApolloConfig().Get()
	config := currentConnApolloConfig[defaultNamespace]

	Assert(t, config, NotNilVal())
	Assert(t, defaultNamespace, Equal(config.NamespaceName))
	Assert(t, apolloConfig.AppID, Equal(config.AppID))
	Assert(t, apolloConfig.Cluster, Equal(config.Cluster))
	Assert(t, "", Equal(config.ReleaseKey))
	Assert(t, len(apolloConfig.Configurations), Equal(5))

}

func TestGetDefaultNamespace(t *testing.T) {
	namespace := GetDefaultNamespace()
	Assert(t, namespace, Equal(defaultNamespace))
}

func TestGetConfig(t *testing.T) {
	configurations := make(map[string]interface{})
	configurations["string"] = "string2"
	configurations["int"] = "2"
	configurations["float"] = "1"
	configurations["bool"] = "false"
	configurations["sliceString"] = []string{"1", "2", "3"}
	configurations["sliceInt"] = []int{1, 2, 3}
	configurations["sliceInter"] = []interface{}{1, "2", 3}
	c := creatTestApolloConfig(configurations, "test")
	config := c.GetConfig("test")
	Assert(t, config, NotNilVal())

	//string
	s := config.GetStringValue("string", "s")
	Assert(t, s, Equal(configurations["string"]))

	s = config.GetStringValue("s", "s")
	Assert(t, s, Equal("s"))

	//int
	i := config.GetIntValue("int", 3)
	Assert(t, i, Equal(2))
	i = config.GetIntValue("i", 3)
	Assert(t, i, Equal(3))

	//float
	f := config.GetFloatValue("float", 2)
	Assert(t, f, Equal(float64(1)))
	f = config.GetFloatValue("f", 2)
	Assert(t, f, Equal(float64(2)))

	//bool
	b := config.GetBoolValue("bool", true)
	Assert(t, b, Equal(false))

	b = config.GetBoolValue("b", false)
	Assert(t, b, Equal(false))

	slice := config.GetStringSliceValue("sliceString")
	Assert(t, slice, Equal([]string{"1", "2", "3"}))

	sliceInt := config.GetIntSliceValue("sliceInt")
	Assert(t, sliceInt, Equal([]int{1, 2, 3}))

	sliceInter := config.GetSliceValue("sliceInter")
	Assert(t, sliceInter, Equal([]interface{}{1, "2", 3}))

	//content
	content := config.GetContent()
	hasFloat := strings.Contains(content, "float=1")
	Assert(t, hasFloat, Equal(true))

	hasInt := strings.Contains(content, "int=2")
	Assert(t, hasInt, Equal(true))

	hasString := strings.Contains(content, "string=string2")
	Assert(t, hasString, Equal(true))

	hasBool := strings.Contains(content, "bool=false")
	Assert(t, hasBool, Equal(true))

	hasSlice := strings.Contains(content, "sliceString=[1 2 3]")
	Assert(t, hasSlice, Equal(true))
	hasSlice = strings.Contains(content, "sliceInt=[1 2 3]")
	Assert(t, hasSlice, Equal(true))
}

func createChangeEvent() *ChangeEvent {
	addConfig := createAddConfigChange("new")
	deleteConfig := createDeletedConfigChange("old")
	modifyConfig := createModifyConfigChange("old", "new")
	changes := make(map[string]*ConfigChange)
	changes["add"] = addConfig
	changes["adx"] = addConfig
	changes["delete"] = deleteConfig
	changes["modify"] = modifyConfig
	cEvent := &ChangeEvent{
		Changes: changes,
	}
	cEvent.Namespace = "a"
	return cEvent
}

func TestRegDispatchInRepository(t *testing.T) {
	dispatch := UseEventDispatch()
	err := dispatch.RegisterListener(nil, "ad.*")
	Assert(t, err, NotNilVal())
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err = dispatch.RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	cEvent := createChangeEvent()
	cache := CreateNamespaceConfig("abc")
	cache.AddChangeListener(dispatch)
	cache.pushChangeEvent(cEvent)
	time.Sleep(1 * time.Second)
	Assert(t, len(l.Keys), Equal(2))
	v, ok := l.Keys["add"]
	Assert(t, v, Equal("new"))
	Assert(t, ok, Equal(true))
	v, ok = l.Keys["adx"]
	Assert(t, v, Equal("new"))
	Assert(t, ok, Equal(true))
}

func TestDispatchInRepository(t *testing.T) {
	dispatch := UseEventDispatch()
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err := dispatch.RegisterListener(l, "add", "delete")
	Assert(t, err, NilVal())
	Assert(t, len(dispatch.listeners), Equal(2))
	cEvent := createChangeEvent()
	cache := CreateNamespaceConfig("abc")
	cache.AddChangeListener(dispatch)
	cache.pushChangeEvent(cEvent)
	time.Sleep(1 * time.Second)
	Assert(t, len(l.Keys), Equal(2))
	v, ok := l.Keys["add"]
	Assert(t, v, Equal("new"))
	Assert(t, ok, Equal(true))
	v, ok = l.Keys["delete"]
	Assert(t, ok, Equal(true))
	Assert(t, v, Equal("old"))
	_, ok = l.Keys["modify"]
	Assert(t, ok, Equal(false))
}
