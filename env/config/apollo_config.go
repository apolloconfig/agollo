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

package config

import (
	"sync"

	"github.com/apolloconfig/agollo/v4/utils"
)

// CurrentApolloConfig represents the current configuration information returned by Apollo
// It maintains a thread-safe map of namespace to connection configurations
type CurrentApolloConfig struct {
	l       sync.RWMutex                 // Lock for thread-safe operations
	configs map[string]*ApolloConnConfig // Map of namespace to connection configurations
}

// CreateCurrentApolloConfig creates a new instance of CurrentApolloConfig
// Returns:
//   - *CurrentApolloConfig: A new instance with initialized configuration map
func CreateCurrentApolloConfig() *CurrentApolloConfig {
	return &CurrentApolloConfig{
		configs: make(map[string]*ApolloConnConfig, 1),
	}
}

// Set stores the Apollo connection configuration for a specific namespace
// Parameters:
//   - namespace: The configuration namespace
//   - connConfig: The connection configuration to be stored
//
// Thread-safe operation using mutex lock
func (c *CurrentApolloConfig) Set(namespace string, connConfig *ApolloConnConfig) {
	c.l.Lock()
	defer c.l.Unlock()

	c.configs[namespace] = connConfig
}

// Get retrieves all Apollo connection configurations
// Returns:
//   - map[string]*ApolloConnConfig: Map of namespace to connection configurations
//
// Thread-safe operation using read lock
func (c *CurrentApolloConfig) Get() map[string]*ApolloConnConfig {
	c.l.RLock()
	defer c.l.RUnlock()

	return c.configs
}

// GetReleaseKey retrieves the release key for a specific namespace
// Parameters:
//   - namespace: The configuration namespace
//
// Returns:
//   - string: The release key if found, empty string otherwise
//
// Thread-safe operation using read lock
func (c *CurrentApolloConfig) GetReleaseKey(namespace string) string {
	c.l.RLock()
	defer c.l.RUnlock()
	config := c.configs[namespace]
	if config == nil {
		return utils.Empty
	}

	return config.ReleaseKey
}

// ApolloConnConfig defines the connection configuration for Apollo
// This structure contains the basic information needed to connect to Apollo server
type ApolloConnConfig struct {
	AppID         string `json:"appId"`         // Application ID
	Cluster       string `json:"cluster"`       // Cluster name
	NamespaceName string `json:"namespaceName"` // Configuration namespace
	ReleaseKey    string `json:"releaseKey"`    // Release key for version control
	sync.RWMutex         // Embedded mutex for thread-safe operations
}

// ApolloConfig represents a complete Apollo configuration
// It extends ApolloConnConfig with actual configuration values
type ApolloConfig struct {
	ApolloConnConfig                        // Embedded connection configuration
	Configurations   map[string]interface{} `json:"configurations"` // Key-value pairs of configurations
}

// Init initializes an Apollo configuration with basic connection information
// Parameters:
//   - appID: Application identifier
//   - cluster: Cluster name
//   - namespace: Configuration namespace
func (a *ApolloConfig) Init(appID string, cluster string, namespace string) {
	a.AppID = appID
	a.Cluster = cluster
	a.NamespaceName = namespace
}
