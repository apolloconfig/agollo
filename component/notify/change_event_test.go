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
	"sync"
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/agcache/memory"
	_ "github.com/apolloconfig/agollo/v4/agcache/memory"
	"github.com/apolloconfig/agollo/v4/cluster/roundrobin"
	"github.com/apolloconfig/agollo/v4/component/remote"
	"github.com/apolloconfig/agollo/v4/env/config"
	_ "github.com/apolloconfig/agollo/v4/env/file/json"
	jsonFile "github.com/apolloconfig/agollo/v4/env/file/json"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/storage"
)

func init() {
	extension.SetLoadBalance(&roundrobin.RoundRobin{})
	extension.SetFileHandler(&jsonFile.FileHandler{})
	extension.SetCacheFactory(&memory.DefaultCacheFactory{})
}

type CustomChangeListener struct {
	t     *testing.T
	group *sync.WaitGroup
}

func (c *CustomChangeListener) OnChange(changeEvent *storage.ChangeEvent) {
	if c.group == nil {
		return
	}
	defer c.group.Done()
	bytes, _ := json.Marshal(changeEvent)
	fmt.Println("event:", string(bytes))

	Assert(c.t, "application", Equal(changeEvent.Namespace))

	Assert(c.t, "string", Equal(changeEvent.Changes["string"].NewValue))
	Assert(c.t, nil, Equal(changeEvent.Changes["string"].OldValue))
	Assert(c.t, storage.ADDED, Equal(changeEvent.Changes["string"].ChangeType))

	Assert(c.t, "value1", Equal(changeEvent.Changes["key1"].NewValue))
	Assert(c.t, nil, Equal(changeEvent.Changes["key2"].OldValue))
	Assert(c.t, storage.ADDED, Equal(changeEvent.Changes["key1"].ChangeType))

	Assert(c.t, "value2", Equal(changeEvent.Changes["key2"].NewValue))
	Assert(c.t, nil, Equal(changeEvent.Changes["key2"].OldValue))
	Assert(c.t, storage.ADDED, Equal(changeEvent.Changes["key2"].ChangeType))
}

func (c *CustomChangeListener) OnNewestChange(event *storage.FullChangeEvent) {

}

func buildNotifyResult(t *testing.T) {
	server := runChangeConfigResponse()
	defer server.Close()

	time.Sleep(1 * time.Second)

	newAppConfig := getTestAppConfig()
	newAppConfig.IP = server.URL

	syncApolloConfig := remote.CreateSyncApolloConfig()
	apolloConfigs := syncApolloConfig.Sync(func() config.AppConfig {
		return *newAppConfig
	})
	apolloConfigs = syncApolloConfig.Sync(func() config.AppConfig {
		return *newAppConfig
	})

	Assert(t, apolloConfigs, NotNilVal())
	Assert(t, len(apolloConfigs), Equal(1))

	newAppConfig.GetCurrentApolloConfig().Set(newAppConfig.NamespaceName, &apolloConfigs[0].ApolloConnConfig)
	config := newAppConfig.GetCurrentApolloConfig().Get()[newAppConfig.NamespaceName]

	Assert(t, "100004458", Equal(config.AppID))
	Assert(t, "default", Equal(config.Cluster))
	Assert(t, "application", Equal(config.NamespaceName))
	Assert(t, "20170430092936-dee2d58e74515ff3", Equal(config.ReleaseKey))
}

func TestListenChangeEvent(t *testing.T) {
	t.SkipNow()
	cache := storage.CreateNamespaceConfig("abc")
	buildNotifyResult(t)
	group := sync.WaitGroup{}
	group.Add(1)

	listener := &CustomChangeListener{
		t:     t,
		group: &group,
	}
	cache.AddChangeListener(listener)
	group.Wait()
	// 运行完清空变更队列
	cache.RemoveChangeListener(listener)
}

func TestRemoveChangeListener(t *testing.T) {
	cache := storage.CreateNamespaceConfig("abc")
	go buildNotifyResult(t)

	listener := &CustomChangeListener{}
	cache.AddChangeListener(listener)
	Assert(t, 1, Equal(cache.GetChangeListeners().Len()))
	cache.RemoveChangeListener(listener)
	Assert(t, 0, Equal(cache.GetChangeListeners().Len()))
}
