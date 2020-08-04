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
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/zouyx/agollo/v3/agcache"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"
)

const (
	//1 minute
	configCacheExpireTime = 120

	defaultNamespace = "application"

	propertiesFormat = "%s=%v\n"
)

var (
	//config from apollo
	apolloConfigCache sync.Map
)

//InitConfigCache 获取程序配置初始化agollo内润配置
func InitConfigCache() {
	if env.GetPlainAppConfig() == nil {
		log.Warn("Config is nil,can not init agollo.")
		return
	}
	CreateNamespaceConfig(env.GetPlainAppConfig().NamespaceName)
}

//CreateNamespaceConfig 根据namespace初始化agollo内润配置
func CreateNamespaceConfig(namespace string) {
	env.SplitNamespaces(namespace, func(namespace string) {
		if _, ok := apolloConfigCache.Load(namespace); ok {
			return
		}
		c := initConfig(namespace, extension.GetCacheFactory())
		apolloConfigCache.Store(namespace, c)
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

//Config apollo配置项
type Config struct {
	namespace string
	cache     agcache.CacheInterface
	isInit    atomic.Value
	waitInit  sync.WaitGroup
}

//GetIsInit 获取标志
func (c *Config) GetIsInit() bool {
	return c.isInit.Load().(bool)
}

//GetWaitInit 获取标志
func (c *Config) GetWaitInit() *sync.WaitGroup {
	return &c.waitInit
}

//GetCache 获取cache
func (c *Config) GetCache() agcache.CacheInterface {
	return c.cache
}

//getConfigValue 获取配置值
func (c *Config) getConfigValue(key string) interface{} {
	b := c.GetIsInit()
	if !b {
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

//GetValue 获取配置值（string）
func (c *Config) GetValue(key string) string {
	value := c.getConfigValue(key)
	if value == nil {
		return utils.Empty
	}

	return value.(string)
}

//GetStringValue 获取配置值（string），获取不到则取默认值
func (c *Config) GetStringValue(key string, defaultValue string) string {
	value := c.GetValue(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

//GetStringSliceValue 获取配置值（[]string）
func (c *Config) GetStringSliceValue(key string) []string {
	value := c.getConfigValue(key)
	if value == nil {
		return []string{}
	}
	return value.([]string)
}

//GetIntSliceValue 获取配置值（[]int)
func (c *Config) GetIntSliceValue(key string) []int {
	value := c.getConfigValue(key)
	if value == nil {
		return []int{}
	}
	return value.([]int)
}

//GetSliceValue 获取配置值（[]interface)
func (c *Config) GetSliceValue(key string) []interface{} {
	value := c.getConfigValue(key)
	if value == nil {
		return []interface{}{}
	}
	return value.([]interface{})
}

//GetIntValue 获取配置值（int），获取不到则取默认值
func (c *Config) GetIntValue(key string, defaultValue int) int {
	value := c.GetValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Debug("convert to int fail!error:", err)
		return defaultValue
	}
	return i
}

//GetFloatValue 获取配置值（float），获取不到则取默认值
func (c *Config) GetFloatValue(key string, defaultValue float64) float64 {
	value := c.GetValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

//GetBoolValue 获取配置值（bool），获取不到则取默认值
func (c *Config) GetBoolValue(key string, defaultValue bool) bool {
	value := c.GetValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}

//UpdateApolloConfig 根据config server返回的内容更新内存
//并判断是否需要写备份文件
func UpdateApolloConfig(apolloConfig *env.ApolloConfig, isBackupConfig bool) {
	if apolloConfig == nil {
		log.Error("apolloConfig is null,can't update!")
		return
	}

	//update apollo connection config
	env.SetCurrentApolloConfig(apolloConfig.NamespaceName, &apolloConfig.ApolloConnConfig)

	//get change list
	changeList := UpdateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime, apolloConfig.NamespaceName)

	//push all newest changes
	pushNewestChanges(apolloConfig.NamespaceName, apolloConfig.Configurations)

	if len(changeList) > 0 {
		//create config change event base on change list
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName)

		//push change event to channel
		pushChangeEvent(event)
	}

	if isBackupConfig {
		//write config file async
		go extension.GetFileHandler().WriteConfigFile(apolloConfig, env.GetPlainAppConfig().GetBackupConfigPath())
	}
}

//UpdateApolloConfigCache 根据conf[ig server返回的内容更新内存
func UpdateApolloConfigCache(configurations map[string]interface{}, expireTime int, namespace string) map[string]*ConfigChange {
	config := GetConfig(namespace)
	if config == nil {
		config = initConfig(namespace, extension.GetCacheFactory())
		apolloConfigCache.Store(namespace, config)
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

	//get old keys
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
			//key state insert or update
			//insert
			if !mp[key] {
				changes[key] = createAddConfigChange(value)
			} else {
				//update
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
		//get old value and del
		oldValue, _ := config.cache.Get(key)
		changes[key] = createDeletedConfigChange(oldValue)

		config.cache.Del(key)
	}
	isInit = true

	return changes
}

//GetContent 获取配置文件内容
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

//GetApolloConfigCache 获取默认namespace的apollo配置
func GetApolloConfigCache() *sync.Map {
	return &apolloConfigCache
}

//GetDefaultNamespace 获取默认命名空间
func GetDefaultNamespace() string {
	return defaultNamespace
}

//GetConfig 根据namespace获取apollo配置
func GetConfig(namespace string) *Config {
	if namespace == "" {
		return nil
	}

	config, ok := GetApolloConfigCache().Load(namespace)

	if !ok {
		return nil
	}

	return config.(*Config)
}
