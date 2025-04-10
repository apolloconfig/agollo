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

package storage

import (
	"container/list"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/apolloconfig/agollo/v4/agcache"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/utils"
)

const (
	// configCacheExpireTime defines the expiration time for configuration cache in seconds
	configCacheExpireTime = 120

	// defaultNamespace is the default namespace for Apollo configuration
	defaultNamespace = "application"

	// propertiesFormat is the format string for properties file output
	propertiesFormat = "%s=%v\n"
)

// Cache represents the Apollo configuration cache
// It maintains a thread-safe storage for configurations and manages change listeners
type Cache struct {
	apolloConfigCache sync.Map
	changeListeners   *list.List
	rw                sync.RWMutex
}

// GetConfig retrieves the configuration for a specific namespace
// Returns nil if the namespace is empty or not found
func (c *Cache) GetConfig(namespace string) *Config {
	if namespace == "" {
		return nil
	}

	config, ok := c.apolloConfigCache.Load(namespace)

	if !ok {
		return nil
	}

	return config.(*Config)
}

// CreateNamespaceConfig initializes Apollo configuration for the given namespace
// It creates a new cache instance and initializes configurations for each namespace
func CreateNamespaceConfig(namespace string) *Cache {
	// config from apollo
	var apolloConfigCache sync.Map
	config.SplitNamespaces(namespace, func(namespace string) {
		if _, ok := apolloConfigCache.Load(namespace); ok {
			return
		}
		c := initConfig(namespace, extension.GetCacheFactory())
		apolloConfigCache.Store(namespace, c)
	})
	return &Cache{
		apolloConfigCache: apolloConfigCache,
		changeListeners:   list.New(),
	}
}

// initConfig initializes a new Config instance for the given namespace
// It sets up the cache and initialization flags
func initConfig(namespace string, factory agcache.CacheFactory) *Config {
	c := &Config{
		namespace: namespace,
		cache:     factory.Create(),
	}
	c.isInit.Store(false)
	c.waitInit.Add(1)
	return c
}

// Config represents an Apollo configuration item
// It contains the namespace, cache, and initialization state
type Config struct {
	namespace string
	cache     agcache.CacheInterface
	isInit    atomic.Value
	waitInit  sync.WaitGroup
}

// GetIsInit returns the initialization status of the configuration
func (c *Config) GetIsInit() bool {
	return c.isInit.Load().(bool)
}

// GetWaitInit returns the WaitGroup for initialization synchronization
func (c *Config) GetWaitInit() *sync.WaitGroup {
	return &c.waitInit
}

// GetCache returns the cache interface for this configuration
func (c *Config) GetCache() agcache.CacheInterface {
	return c.cache
}

// getConfigValue retrieves a configuration value by key
// If waitInit is true, it will wait for initialization to complete
// Returns nil if the key doesn't exist or there's an error
func (c *Config) getConfigValue(key string, waitInit bool) interface{} {
	b := c.GetIsInit()
	if !b {
		if !waitInit {
			log.Errorf("getConfigValue fail, init not done, namespace:%s key:%s", c.namespace, key)
			return nil
		}
		c.waitInit.Wait()
	}
	if c.cache == nil {
		log.Errorf("get config value fail! namespace:%s not exist!", c.namespace)
		return nil
	}

	value, err := c.cache.Get(key)
	if err != nil {
		log.Errorf("get config value fail! key:%s, error:%v", key, err)
		return nil
	}

	return value
}

// GetValueImmediately retrieves a string configuration value without waiting for initialization
// Returns empty string if the key doesn't exist or there's an error
func (c *Config) GetValueImmediately(key string) string {
	value := c.getConfigValue(key, false)
	if value == nil {
		return utils.Empty
	}

	v, ok := value.(string)
	if !ok {
		log.Debugf("convert to string fail ! source type:%T", value)
		return utils.Empty
	}
	return v
}

// GetStringValueImmediately retrieves a string configuration value with default fallback
// Returns the default value if the key doesn't exist
func (c *Config) GetStringValueImmediately(key string, defaultValue string) string {
	value := c.GetValueImmediately(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

// GetStringSliceValueImmediately retrieves a string slice configuration value without waiting
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetStringSliceValueImmediately(key string, defaultValue []string) []string {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]string)
	if !ok {
		log.Debugf("convert to []string fail ! source type:%T", value)
		return defaultValue
	}
	return v
}

// GetIntSliceValueImmediately retrieves an integer slice configuration value without waiting
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetIntSliceValueImmediately(key string, defaultValue []int) []int {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]int)
	if !ok {
		log.Debugf("convert to []int fail ! source type:%T", value)
		return defaultValue
	}
	return v
}

// GetSliceValueImmediately retrieves an interface slice configuration value without waiting
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetSliceValueImmediately(key string, defaultValue []interface{}) []interface{} {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]interface{})
	if !ok {
		log.Debugf("convert to []interface{} fail ! source type:%T", value)
		return defaultValue
	}
	return v
}

// GetIntValueImmediately retrieves an integer configuration value without waiting
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetIntValueImmediately(key string, defaultValue int) int {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(int)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to int fail ! source type:%T", value)
		return defaultValue
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		log.Debugf("Atoi fail, error:%v", err)
		return defaultValue
	}

	return v
}

// GetFloatValueImmediately retrieves a float configuration value without waiting
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetFloatValueImmediately(key string, defaultValue float64) float64 {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(float64)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to float64 fail ! source type:%T", value)
		return defaultValue
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Debugf("ParseFloat fail, error:%v", err)
		return defaultValue
	}

	return v
}

// GetBoolValueImmediately retrieves a boolean configuration value without waiting
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetBoolValueImmediately(key string, defaultValue bool) bool {
	value := c.getConfigValue(key, false)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(bool)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to bool fail ! source type:%T", value)
		return defaultValue
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Debugf("ParseBool fail, error:%v", err)
		return defaultValue
	}

	return v
}

// GetValue retrieves a string configuration value, waiting for initialization if necessary
// Returns empty string if the key doesn't exist or conversion fails
func (c *Config) GetValue(key string) string {
	value := c.getConfigValue(key, true)
	if value == nil {
		return utils.Empty
	}

	v, ok := value.(string)
	if !ok {
		log.Debugf("convert to string fail ! source type:%T", value)
		return utils.Empty
	}
	return v
}

// GetStringValue retrieves a string configuration value with default fallback
// Returns the default value if the key doesn't exist
func (c *Config) GetStringValue(key string, defaultValue string) string {
	value := c.GetValue(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

// GetStringSliceValue retrieves a string slice configuration value
// The values are split by the separator if the value is a string
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetStringSliceValue(key, separator string, defaultValue []string) []string {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]string)
	if !ok {
		s, ok := value.(string)
		if !ok {
			log.Debugf("convert to []string fail ! source type:%T", value)
			return defaultValue
		}
		return strings.Split(s, separator)
	}
	return v
}

// GetIntSliceValue retrieves an integer slice configuration value
// The values are split by the separator if the value is a string
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetIntSliceValue(key, separator string, defaultValue []int) []int {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]int)
	if !ok {
		sl := c.GetStringSliceValue(key, separator, nil)
		if sl == nil {
			return defaultValue
		}
		v = make([]int, 0, len(sl))
		for index := range sl {
			i, err := strconv.Atoi(sl[index])
			if err != nil {
				log.Debugf("convert to []int fail! value:%s,  source type:%T", sl[index], sl[index])
				return defaultValue
			}
			v = append(v, i)
		}
	}
	return v
}

// GetSliceValue retrieves an interface slice configuration value
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetSliceValue(key string, defaultValue []interface{}) []interface{} {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.([]interface{})
	if !ok {
		log.Debugf("convert to []interface{} fail ! source type:%T", value)
		return defaultValue
	}
	return v
}

// GetIntValue retrieves an integer configuration value with default fallback
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetIntValue(key string, defaultValue int) int {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(int)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to int fail ! source type:%T", value)
		return defaultValue
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		log.Debugf("Atoi fail, error:%v", err)
		return defaultValue
	}
	return v
}

// GetFloatValue retrieves a float configuration value with default fallback
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetFloatValue(key string, defaultValue float64) float64 {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(float64)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to float64 fail ! source type:%T", value)
		return defaultValue
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Debugf("ParseFloat fail, error:%v", err)
		return defaultValue
	}
	return v
}

// GetBoolValue retrieves a boolean configuration value with default fallback
// Returns the default value if the key doesn't exist or conversion fails
func (c *Config) GetBoolValue(key string, defaultValue bool) bool {
	value := c.getConfigValue(key, true)
	if value == nil {
		return defaultValue
	}

	v, ok := value.(bool)
	if ok {
		return v
	}

	s, ok := value.(string)
	if !ok {
		log.Debugf("convert to bool fail ! source type:%T", value)
		return defaultValue
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		log.Debugf("ParseBool fail, error:%v", err)
		return defaultValue
	}
	return v
}

// UpdateApolloConfig updates the in-memory configuration based on the server response
// It also determines whether to write a backup file
func (c *Cache) UpdateApolloConfig(apolloConfig *config.ApolloConfig, appConfigFunc func() config.AppConfig) {
	if apolloConfig == nil {
		log.Error("apolloConfig is null, can't update!")
		return
	}

	appConfig := appConfigFunc()
	// update apollo connection config
	appConfig.SetCurrentApolloConfig(&apolloConfig.ApolloConnConfig)

	// get change list
	changeList := c.UpdateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime, apolloConfig.NamespaceName)

	notify := appConfig.GetNotificationsMap().GetNotify(apolloConfig.NamespaceName)

	// push all newest changes
	c.pushNewestChanges(apolloConfig.NamespaceName, apolloConfig.Configurations, notify)

	if len(changeList) > 0 {
		// create config change event base on change list
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName, notify)

		// push change event to channel
		c.pushChangeEvent(event)
	}

	if appConfig.GetIsBackupConfig() {
		// write config file async
		apolloConfig.AppID = appConfig.AppID
		go extension.GetFileHandler().WriteConfigFile(apolloConfig, appConfig.GetBackupConfigPath())
	}
}

// UpdateApolloConfigCache updates the in-memory cache based on the server response
// It returns a map of configuration changes
func (c *Cache) UpdateApolloConfigCache(configurations map[string]interface{}, expireTime int, namespace string) map[string]*ConfigChange {
	config := c.GetConfig(namespace)
	if config == nil {
		config = initConfig(namespace, extension.GetCacheFactory())
		c.apolloConfigCache.Store(namespace, config)
	}

	isInit := false
	defer func(c *Config) {
		if !isInit {
			return
		}
		b := c.GetIsInit()
		if b {
			return
		}
		c.isInit.Store(isInit)
		c.waitInit.Done()
	}(config)

	if (len(configurations) == 0) && config.cache.EntryCount() == 0 {
		return nil
	}

	// get old keys
	mp := map[string]bool{}
	config.cache.Range(func(key, value interface{}) bool {
		mp[key.(string)] = true
		return true
	})

	changes := make(map[string]*ConfigChange)

	// update new keys
	for key, value := range configurations {
		// key state insert or update
		// insert
		if !mp[key] {
			changes[key] = createAddConfigChange(value)
		} else {
			// update
			oldValue, _ := config.cache.Get(key)
			if !reflect.DeepEqual(oldValue, value) {
				changes[key] = createModifyConfigChange(oldValue, value)
			}
		}

		if err := config.cache.Set(key, value, expireTime); err != nil {
			log.Errorf("set key %s to cache, error: %v", key, err)
		}

		delete(mp, key)
	}

	// remove deleted keys
	for key := range mp {
		// get old value and delete
		oldValue, _ := config.cache.Get(key)
		changes[key] = createDeletedConfigChange(oldValue)

		config.cache.Del(key)
	}
	isInit = true

	return changes
}

// GetContent retrieves the configuration content in properties format
func (c *Config) GetContent() string {
	return convertToProperties(c.cache)
}

// convertToProperties converts the cache content to properties format
func convertToProperties(cache agcache.CacheInterface) string {
	properties := utils.Empty
	if cache == nil {
		return properties
	}
	cache.Range(func(key, value interface{}) bool {
		properties += fmt.Sprintf(propertiesFormat, key, value)
		return true
	})
	return properties
}

// GetDefaultNamespace returns the default namespace for Apollo configuration
func GetDefaultNamespace() string {
	return defaultNamespace
}

// AddChangeListener adds a change listener to the cache
// The listener will be notified when configuration changes occur
func (c *Cache) AddChangeListener(listener ChangeListener) {
	if listener == nil {
		return
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	c.changeListeners.PushBack(listener)
}

// RemoveChangeListener removes a change listener from the cache
// The listener will no longer receive configuration change notifications
func (c *Cache) RemoveChangeListener(listener ChangeListener) {
	if listener == nil {
		return
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	for i := c.changeListeners.Front(); i != nil; i = i.Next() {
		apolloListener := i.Value.(ChangeListener)
		if listener == apolloListener {
			c.changeListeners.Remove(i)
		}
	}
}

// GetChangeListeners returns the list of configuration change listeners
// Returns a new list containing all registered listeners
func (c *Cache) GetChangeListeners() *list.List {
	if c.changeListeners == nil {
		return nil
	}
	c.rw.RLock()
	defer c.rw.RUnlock()
	l := list.New()
	l.PushBackList(c.changeListeners)
	return l
}

// pushChangeEvent pushes a configuration change event to all listeners
func (c *Cache) pushChangeEvent(event *ChangeEvent) {
	c.pushChange(func(listener ChangeListener) {
		go listener.OnChange(event)
	})
}

// pushNewestChanges pushes the latest configuration changes to all listeners
func (c *Cache) pushNewestChanges(namespace string, configuration map[string]interface{}, notificationID int64) {
	e := &FullChangeEvent{
		Changes: configuration,
	}
	e.Namespace = namespace
	e.NotificationID = notificationID
	c.pushChange(func(listener ChangeListener) {
		go listener.OnNewestChange(e)
	})
}

// pushChange executes the given function for each change listener
// It handles the case when there are no listeners
func (c *Cache) pushChange(f func(ChangeListener)) {
	// if channel is null, mean no listener, don't need to push msg
	listeners := c.GetChangeListeners()
	if listeners == nil || listeners.Len() == 0 {
		return
	}

	for i := listeners.Front(); i != nil; i = i.Next() {
		listener := i.Value.(ChangeListener)
		f(listener)
	}
}
