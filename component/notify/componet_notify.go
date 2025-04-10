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
	"time"

	"github.com/apolloconfig/agollo/v4/component/remote"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
)

const (
	// longPollInterval defines the interval for long polling operations
	// The client will check for configuration updates every 2 seconds
	longPollInterval = 2 * time.Second
)

// ConfigComponent handles the configuration synchronization process
// It maintains a long-polling mechanism to fetch updates from Apollo server
type ConfigComponent struct {
	// appConfigFunc is a function that returns the current application configuration
	appConfigFunc func() config.AppConfig
	// cache stores the configuration data
	cache *storage.Cache
	// stopCh is used to signal the termination of the polling process
	stopCh chan interface{}
}

// SetAppConfig sets the function that provides application configuration
// Parameters:
//   - appConfigFunc: A function that returns the current AppConfig
func (c *ConfigComponent) SetAppConfig(appConfigFunc func() config.AppConfig) {
	c.appConfigFunc = appConfigFunc
}

// SetCache sets the cache instance for storing configuration data
// Parameters:
//   - cache: The storage cache instance to be used
func (c *ConfigComponent) SetCache(cache *storage.Cache) {
	c.cache = cache
}

// Start initiates the configuration synchronization process
// This method:
// 1. Initializes the stop channel if not already initialized
// 2. Creates a timer for periodic synchronization
// 3. Starts a loop that continuously checks for configuration updates
// 4. Updates the local cache when new configurations are received
func (c *ConfigComponent) Start() {
	if c.stopCh == nil {
		c.stopCh = make(chan interface{})
	}

	t2 := time.NewTimer(longPollInterval)
	instance := remote.CreateAsyncApolloConfig()
	//long poll for sync
loop:
	for {
		select {
		case <-t2.C:
			configs := instance.Sync(c.appConfigFunc)
			for _, apolloConfig := range configs {
				c.cache.UpdateApolloConfig(apolloConfig, c.appConfigFunc)
			}
			t2.Reset(longPollInterval)
		case <-c.stopCh:
			break loop
		}
	}
}

// Stop terminates the configuration synchronization process
// This method closes the stop channel, which will cause the polling loop to exit
func (c *ConfigComponent) Stop() {
	if c.stopCh != nil {
		close(c.stopCh)
	}
}
