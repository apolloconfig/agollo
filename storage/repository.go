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
	"container/list"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/qshuai/agollo/v4/agcache"
	"github.com/qshuai/agollo/v4/component/log"
	"github.com/qshuai/agollo/v4/env/config"
	"github.com/qshuai/agollo/v4/extension"
	"github.com/qshuai/agollo/v4/utils"
)

const (
	// 1 minute
	configCacheExpireTime = 120

	defaultNamespace = "application"

	propertiesFormat = "%s=%v\n"
)

// Cache apollo 配置缓存
type Cache struct {
	// type: map[string]*Config
	// business: map[namespace]*Config
	apolloConfigCache sync.Map
	changeListeners   *list.List
	rw                sync.RWMutex
}

// GetConfig 根据namespace获取apollo配置
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

// CreateNamespaceConfig 根据namespace初始化agollo内部配置
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

// AddNamespaceConfig Add new namespace at runtime
func AddNamespaceConfig(cache *Cache, namespace string) {
	config.SplitNamespaces(namespace, func(namespace string) {
		if _, ok := cache.apolloConfigCache.Load(namespace); ok {
			return
		}
		c := initConfig(namespace, extension.GetCacheFactory())
		cache.apolloConfigCache.Store(namespace, c)
	})
}

func initConfig(namespace string, factory agcache.CacheFactory) *Config {
	c := &Config{
		namespace: namespace,
		cache:     factory.Create(),
	}
	c.isInit.Store(false)
	c.waitInit.Add(1)
	return c
}

// Config apollo配置项
type Config struct {
	namespace string
	cache     agcache.CacheInterface
	isInit    atomic.Value
	waitInit  sync.WaitGroup
}

// GetIsInit 获取标志
func (c *Config) GetIsInit() bool {
	return c.isInit.Load().(bool)
}

// GetWaitInit 获取标志
func (c *Config) GetWaitInit() *sync.WaitGroup {
	return &c.waitInit
}

// GetCache 获取cache
func (c *Config) GetCache() agcache.CacheInterface {
	return c.cache
}

// getConfigValue 获取配置值
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
		log.Errorf("get config value fail!namespace:%s is not exist!", c.namespace)
		return nil
	}

	value, err := c.cache.Get(key)
	if err != nil {
		log.Errorf("get config value fail!key:%s,err:%s", key, err)
		return nil
	}

	return value
}

// GetValueImmediately 获取配置值（string），立即返回，初始化未完成直接返回错误
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

// GetStringValueImmediately 获取配置值（string），立即返回，初始化未完成直接返回错误
func (c *Config) GetStringValueImmediately(key string, defaultValue string) string {
	value := c.GetValueImmediately(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

// GetStringSliceValueImmediately 获取配置值（[]string），立即返回，初始化未完成直接返回错误
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

// GetIntSliceValueImmediately 获取配置值（[]int)，立即返回，初始化未完成直接返回错误
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

// GetSliceValueImmediately 获取配置值（[]interface)，立即返回，初始化未完成直接返回错误
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

// GetIntValueImmediately 获取配置值（int），获取不到则取默认值，立即返回，初始化未完成直接返回错误
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
		log.Debug("Atoi fail err:%s", err.Error())
		return defaultValue
	}

	return v
}

// GetFloatValueImmediately 获取配置值（float），获取不到则取默认值，立即返回，初始化未完成直接返回错误
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
		log.Debug("ParseFloat fail err:%s", err.Error())
		return defaultValue
	}

	return v
}

// GetBoolValueImmediately 获取配置值（bool），获取不到则取默认值，立即返回，初始化未完成直接返回错误
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
		log.Debug("ParseBool fail err:%s", err.Error())
		return defaultValue
	}

	return v
}

// GetValue 获取配置值（string）
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

// GetStringValue 获取配置值（string），获取不到则取默认值
func (c *Config) GetStringValue(key string, defaultValue string) string {
	value := c.GetValue(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

// GetStringSliceValue 获取配置值（[]string）
func (c *Config) GetStringSliceValue(key string, defaultValue []string) []string {
	value := c.getConfigValue(key, true)
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

// GetIntSliceValue 获取配置值（[]int)
func (c *Config) GetIntSliceValue(key string, defaultValue []int) []int {
	value := c.getConfigValue(key, true)
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

// GetSliceValue 获取配置值（[]interface)
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

// GetIntValue 获取配置值（int），获取不到则取默认值
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
		log.Debug("Atoi fail  err:%s", err.Error())
		return defaultValue
	}
	return v
}

// GetFloatValue 获取配置值（float），获取不到则取默认值
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
		log.Debug("ParseFloat fail err:%s", err.Error())
		return defaultValue
	}
	return v
}

// GetBoolValue 获取配置值（bool），获取不到则取默认值
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
		log.Debug("ParseBool fail err:%s", err.Error())
		return defaultValue
	}
	return v
}

// UpdateApolloConfig 根据config server返回的内容更新内存
// 并判断是否需要写备份文件
func (c *Cache) UpdateApolloConfig(apolloConfig *config.ApolloConfig, appConfigFunc func() config.AppConfig) {
	if apolloConfig == nil {
		log.Error("apolloConfig is null,can't update!")
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

// UpdateApolloConfigCache 根据conf[ig server返回的内容更新内存
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

	if (configurations == nil || len(configurations) == 0) && config.cache.EntryCount() == 0 {
		return nil
	}

	// get old keys
	mp := map[string]bool{}
	config.cache.Range(func(key, value interface{}) bool {
		mp[key.(string)] = true
		return true
	})

	changes := make(map[string]*ConfigChange)

	if configurations != nil {
		// update new
		// keys
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
				log.Errorf("set key %s to cache error %s", key, err)
			}
			delete(mp, key)
		}
	}

	// remove del keys
	for key := range mp {
		// get old value and del
		oldValue, _ := config.cache.Get(key)
		changes[key] = createDeletedConfigChange(oldValue)

		config.cache.Del(key)
	}
	isInit = true

	return changes
}

// GetContent 获取配置文件内容
func (c *Config) GetContent() string {
	return convertToProperties(c.cache)
}

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

// GetDefaultNamespace 获取默认命名空间
func GetDefaultNamespace() string {
	return defaultNamespace
}

// AddChangeListener 增加变更监控
func (c *Cache) AddChangeListener(listener ChangeListener) {
	if listener == nil {
		return
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	c.changeListeners.PushBack(listener)
}

// RemoveChangeListener 增加变更监控
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

// GetChangeListeners 获取配置修改监听器列表
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

// push config change event
func (c *Cache) pushChangeEvent(event *ChangeEvent) {
	c.pushChange(func(listener ChangeListener) {
		go listener.OnChange(event)
	})
}

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

func (c *Cache) pushChange(f func(ChangeListener)) {
	// if channel is null ,mean no listener,don't need to push msg
	listeners := c.GetChangeListeners()
	if listeners == nil || listeners.Len() == 0 {
		return
	}

	for i := listeners.Front(); i != nil; i = i.Next() {
		listener := i.Value.(ChangeListener)
		f(listener)
	}
}
