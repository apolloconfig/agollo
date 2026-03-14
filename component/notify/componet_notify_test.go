// Copyright 2025 Apollo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notify

import (
	"encoding/json"
	"testing"

	"github.com/apolloconfig/agollo/v5/cluster/roundrobin"
	_ "github.com/apolloconfig/agollo/v5/cluster/roundrobin"
	"github.com/apolloconfig/agollo/v5/env"
	"github.com/apolloconfig/agollo/v5/env/config"
	jsonConfig "github.com/apolloconfig/agollo/v5/env/config/json"
	_ "github.com/apolloconfig/agollo/v5/env/file/json"
	jsonFile "github.com/apolloconfig/agollo/v5/env/file/json"
	"github.com/apolloconfig/agollo/v5/extension"
	"github.com/apolloconfig/agollo/v5/storage"
	. "github.com/tevid/gohamcrest"
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

// TestConfigComponent_Stop 测试重复调用stop()和stopCh为空的场景
func TestConfigComponent_Stop(t *testing.T) {
	type fields struct {
		appConfigFunc func() config.AppConfig
		cache         *storage.Cache
		stopCh        chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test_component_stop",
			fields: fields{
				stopCh: make(chan struct{}),
			},
		},
		{
			name:   "test_component_stop_chan_nil",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigComponent{
				appConfigFunc: tt.fields.appConfigFunc,
				cache:         tt.fields.cache,
				stopCh:        tt.fields.stopCh,
			}
			c.Stop()
			c.Stop()
		})
	}
}
