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
	"sync"
	"time"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/component/remote"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
)

const (
	longPollInterval = 2 * time.Second //2s
)

// ConfigComponent 配置组件
type ConfigComponent struct {
	appConfigFunc func() config.AppConfig
	cache         *storage.Cache
	stopCh        chan struct{}
	stopOnce      sync.Once
	stopMu        sync.Mutex
}

func NewConfigComponent(appConfigFunc func() config.AppConfig, cache *storage.Cache) *ConfigComponent {
	return &ConfigComponent{
		appConfigFunc: appConfigFunc,
		cache:         cache,
		stopCh:        make(chan struct{}),
	}
}

// Start 启动配置组件定时器
func (c *ConfigComponent) Start() {
	stopCh := c.ensureStopCh()
	t2 := time.NewTimer(longPollInterval)
	defer t2.Stop()
	instance := remote.CreateAsyncApolloConfig()
	log.Debug("ConfigComponent started")
	//long poll for sync
	for {
		select {
		case <-t2.C:
			configs := instance.Sync(c.appConfigFunc)
			for _, apolloConfig := range configs {
				c.cache.UpdateApolloConfig(apolloConfig, c.appConfigFunc)
			}
			t2.Reset(longPollInterval)
		case <-stopCh:
			log.Debug("ConfigComponent stopped")
			return
		}
	}
}

// Stop 停止配置组件定时器
func (c *ConfigComponent) Stop() {
	c.stopOnce.Do(func() {
		c.stopMu.Lock()
		defer c.stopMu.Unlock()
		if c.stopCh != nil {
			close(c.stopCh)
		}
	})
}

func (c *ConfigComponent) ensureStopCh() chan struct{} {
	c.stopMu.Lock()
	defer c.stopMu.Unlock()
	if c.stopCh == nil {
		c.stopCh = make(chan struct{})
	}
	return c.stopCh
}
