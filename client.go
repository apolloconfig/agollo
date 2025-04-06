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

package agollo

import (
	"container/list"
	"errors"
	"strings"

	"github.com/apolloconfig/agollo/v4/agcache"
	"github.com/apolloconfig/agollo/v4/agcache/memory"
	"github.com/apolloconfig/agollo/v4/cluster/roundrobin"
	"github.com/apolloconfig/agollo/v4/component"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/component/notify"
	"github.com/apolloconfig/agollo/v4/component/remote"
	"github.com/apolloconfig/agollo/v4/component/serverlist"
	"github.com/apolloconfig/agollo/v4/constant"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	jsonFile "github.com/apolloconfig/agollo/v4/env/file/json"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/protocol/auth/sign"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/apolloconfig/agollo/v4/utils"
	"github.com/apolloconfig/agollo/v4/utils/parse/normal"
	"github.com/apolloconfig/agollo/v4/utils/parse/properties"
	"github.com/apolloconfig/agollo/v4/utils/parse/yaml"
	"github.com/apolloconfig/agollo/v4/utils/parse/yml"
)

// separator is used to split string values in configuration
const separator = ","

// init initializes the default components and extensions for the Apollo client
func init() {
	extension.SetCacheFactory(&memory.DefaultCacheFactory{})
	extension.SetLoadBalance(&roundrobin.RoundRobin{})
	extension.SetFileHandler(&jsonFile.FileHandler{})
	extension.SetHTTPAuth(&sign.AuthSignature{})

	// Register file parsers for different configuration formats
	extension.AddFormatParser(constant.DEFAULT, &normal.Parser{})
	extension.AddFormatParser(constant.Properties, &properties.Parser{})
	extension.AddFormatParser(constant.YML, &yml.Parser{})
	extension.AddFormatParser(constant.YAML, &yaml.Parser{})
}

// syncApolloConfig is used to synchronize Apollo configurations
var syncApolloConfig = remote.CreateSyncApolloConfig()

// Client defines the interface for Apollo configuration client
// It provides methods to access and manage Apollo configurations
type Client interface {
	// GetConfig retrieves the configuration for a specific namespace
	GetConfig(namespace string) *storage.Config
	// GetConfigAndInit retrieves and initializes the configuration for a specific namespace
	GetConfigAndInit(namespace string) *storage.Config
	// GetConfigCache returns the cache interface for a specific namespace
	GetConfigCache(namespace string) agcache.CacheInterface
	// GetDefaultConfigCache returns the cache interface for the default namespace
	GetDefaultConfigCache() agcache.CacheInterface
	// GetApolloConfigCache returns the cache interface for Apollo configurations
	GetApolloConfigCache() agcache.CacheInterface
	// GetValue retrieves a configuration value by key
	GetValue(key string) string
	// GetStringValue retrieves a string configuration value with default fallback
	GetStringValue(key string, defaultValue string) string
	// GetIntValue retrieves an integer configuration value with default fallback
	GetIntValue(key string, defaultValue int) int
	// GetFloatValue retrieves a float configuration value with default fallback
	GetFloatValue(key string, defaultValue float64) float64
	// GetBoolValue retrieves a boolean configuration value with default fallback
	GetBoolValue(key string, defaultValue bool) bool
	// GetStringSliceValue retrieves a string slice configuration value with default fallback
	GetStringSliceValue(key string, defaultValue []string) []string
	// GetIntSliceValue retrieves an integer slice configuration value with default fallback
	GetIntSliceValue(key string, defaultValue []int) []int
	// AddChangeListener adds a listener for configuration changes
	AddChangeListener(listener storage.ChangeListener)
	// RemoveChangeListener removes a configuration change listener
	RemoveChangeListener(listener storage.ChangeListener)
	// GetChangeListeners returns the list of configuration change listeners
	GetChangeListeners() *list.List
	// UseEventDispatch enables event dispatch for configuration changes
	UseEventDispatch()
	// Close stops the configuration polling
	Close()
}

// internalClient represents the internal implementation of the Apollo client
type internalClient struct {
	initAppConfigFunc func() (*config.AppConfig, error) // nolint: unused
	appConfig         *config.AppConfig
	cache             *storage.Cache
	configComponent   *notify.ConfigComponent
}

// getAppConfig returns the current application configuration
func (c *internalClient) getAppConfig() config.AppConfig {
	return *c.appConfig
}

// create initializes a new internal client instance
func create() *internalClient {
	appConfig := env.InitFileConfig()
	return &internalClient{
		appConfig: appConfig,
	}
}

// Start initializes the Apollo client with default configuration
// Returns a Client interface and any error that occurred during initialization
func Start() (Client, error) {
	return StartWithConfig(nil)
}

// StartWithConfig initializes the Apollo client with custom configuration
// loadAppConfig is a function that provides custom application configuration
// Returns a Client interface and any error that occurred during initialization
func StartWithConfig(loadAppConfig func() (*config.AppConfig, error)) (Client, error) {
	// Initialize configuration
	appConfig, err := env.InitConfig(loadAppConfig)
	if err != nil {
		return nil, err
	}

	c := create()
	if appConfig != nil {
		c.appConfig = appConfig
	}

	c.cache = storage.CreateNamespaceConfig(appConfig.NamespaceName)
	appConfig.Init()

	// Initialize server list synchronization
	serverlist.InitSyncServerIPList(c.getAppConfig)

	// First synchronization of configurations
	configs := syncApolloConfig.Sync(c.getAppConfig)
	if len(configs) == 0 && appConfig != nil && appConfig.MustStart {
		return nil, errors.New("start failed cause no config was read")
	}

	// Update cache with synchronized configurations
	for _, apolloConfig := range configs {
		c.cache.UpdateApolloConfig(apolloConfig, c.getAppConfig)
	}

	log.Debug("init notifySyncConfigServices finished")

	// Start long polling for configuration updates
	configComponent := &notify.ConfigComponent{}
	configComponent.SetAppConfig(c.getAppConfig)
	configComponent.SetCache(c.cache)
	go component.StartRefreshConfig(configComponent)
	c.configComponent = configComponent

	log.Info("agollo start finished ! ")

	return c, nil
}

// GetConfig retrieves the configuration for a specific namespace
// If the namespace is empty, returns nil
func (c *internalClient) GetConfig(namespace string) *storage.Config {
	return c.GetConfigAndInit(namespace)
}

// GetConfigAndInit retrieves and initializes the configuration for a specific namespace
// If the configuration doesn't exist, it will synchronize with the Apollo server
// Returns nil if the namespace is empty
func (c *internalClient) GetConfigAndInit(namespace string) *storage.Config {
	if namespace == "" {
		return nil
	}

	cfg := c.cache.GetConfig(namespace)

	if cfg == nil {
		// Synchronize configuration from Apollo server
		apolloConfig := syncApolloConfig.SyncWithNamespace(namespace, c.getAppConfig)
		if apolloConfig != nil {
			c.SyncAndUpdate(namespace, apolloConfig)
		}
	}

	cfg = c.cache.GetConfig(namespace)

	return cfg
}

// SyncAndUpdate synchronizes and updates the configuration for a specific namespace
// It updates the appConfig, notification map, and cache with the new configuration
func (c *internalClient) SyncAndUpdate(namespace string, apolloConfig *config.ApolloConfig) {
	// Update appConfig only if namespace does not exist yet
	namespaces := strings.Split(c.appConfig.NamespaceName, ",")
	exists := false
	for _, n := range namespaces {
		if n == namespace {
			exists = true
			break
		}
	}
	if !exists {
		c.appConfig.NamespaceName += "," + namespace
	}

	// Update notification map
	c.appConfig.GetNotificationsMap().UpdateNotify(namespace, 0)

	// Update cache with new configuration
	c.cache.UpdateApolloConfig(apolloConfig, c.getAppConfig)
}

// GetConfigCache returns the cache interface for a specific namespace
// Returns nil if the configuration doesn't exist
func (c *internalClient) GetConfigCache(namespace string) agcache.CacheInterface {
	config := c.GetConfigAndInit(namespace)
	if config == nil {
		return nil
	}

	return config.GetCache()
}

// GetDefaultConfigCache returns the cache interface for the default namespace
// Returns nil if the default configuration doesn't exist
func (c *internalClient) GetDefaultConfigCache() agcache.CacheInterface {
	config := c.GetConfigAndInit(storage.GetDefaultNamespace())
	if config != nil {
		return config.GetCache()
	}
	return nil
}

// GetApolloConfigCache returns the cache interface for Apollo configurations
// This is an alias for GetDefaultConfigCache
func (c *internalClient) GetApolloConfigCache() agcache.CacheInterface {
	return c.GetDefaultConfigCache()
}

// GetValue retrieves a configuration value by key from the default namespace
// Returns the value as a string
func (c *internalClient) GetValue(key string) string {
	return c.GetConfig(storage.GetDefaultNamespace()).GetValue(key)
}

// GetStringValue retrieves a string configuration value with default fallback
// Returns the default value if the key doesn't exist
func (c *internalClient) GetStringValue(key string, defaultValue string) string {
	return c.GetConfig(storage.GetDefaultNamespace()).GetStringValue(key, defaultValue)
}

// GetIntValue retrieves an integer configuration value with default fallback
// Returns the default value if the key doesn't exist or cannot be converted to int
func (c *internalClient) GetIntValue(key string, defaultValue int) int {
	return c.GetConfig(storage.GetDefaultNamespace()).GetIntValue(key, defaultValue)
}

// GetFloatValue retrieves a float configuration value with default fallback
// Returns the default value if the key doesn't exist or cannot be converted to float
func (c *internalClient) GetFloatValue(key string, defaultValue float64) float64 {
	return c.GetConfig(storage.GetDefaultNamespace()).GetFloatValue(key, defaultValue)
}

// GetBoolValue retrieves a boolean configuration value with default fallback
// Returns the default value if the key doesn't exist or cannot be converted to bool
func (c *internalClient) GetBoolValue(key string, defaultValue bool) bool {
	return c.GetConfig(storage.GetDefaultNamespace()).GetBoolValue(key, defaultValue)
}

// GetStringSliceValue retrieves a string slice configuration value with default fallback
// The values are split by the separator constant
// Returns the default value if the key doesn't exist
func (c *internalClient) GetStringSliceValue(key string, defaultValue []string) []string {
	return c.GetConfig(storage.GetDefaultNamespace()).GetStringSliceValue(key, separator, defaultValue)
}

// GetIntSliceValue retrieves an integer slice configuration value with default fallback
// The values are split by the separator constant
// Returns the default value if the key doesn't exist or values cannot be converted to integers
func (c *internalClient) GetIntSliceValue(key string, defaultValue []int) []int {
	return c.GetConfig(storage.GetDefaultNamespace()).GetIntSliceValue(key, separator, defaultValue)
}

// getConfigValue retrieves a raw configuration value from the default namespace cache
// Returns utils.Empty if the key doesn't exist or there's an error
func (c *internalClient) getConfigValue(key string) interface{} {
	cache := c.GetDefaultConfigCache()
	if cache == nil {
		return utils.Empty
	}

	value, err := cache.Get(key)
	if err != nil {
		log.Errorf("get config value fail! key:%s, error:%v", key, err)
		return utils.Empty
	}

	return value
}

// AddChangeListener adds a listener for configuration changes
// The listener will be notified when configuration changes occur
func (c *internalClient) AddChangeListener(listener storage.ChangeListener) {
	c.cache.AddChangeListener(listener)
}

// RemoveChangeListener removes a configuration change listener
// The listener will no longer receive configuration change notifications
func (c *internalClient) RemoveChangeListener(listener storage.ChangeListener) {
	c.cache.RemoveChangeListener(listener)
}

// GetChangeListeners returns the list of configuration change listeners
// Returns a list.List containing all registered listeners
func (c *internalClient) GetChangeListeners() *list.List {
	return c.cache.GetChangeListeners()
}

// UseEventDispatch enables event dispatch for configuration changes
// This will add a default event dispatcher as a change listener
func (c *internalClient) UseEventDispatch() {
	c.AddChangeListener(storage.UseEventDispatch())
}

// Close stops the configuration polling and cleanup resources
// This should be called when the client is no longer needed
func (c *internalClient) Close() {
	c.configComponent.Stop()
}
