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

package agollo

import (
	"strconv"

	"github.com/zouyx/agollo/v3/agcache"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/component/notify"
	"github.com/zouyx/agollo/v3/storage"
	"github.com/zouyx/agollo/v3/utils"
)

//GetConfig 根据namespace获取apollo配置
func GetConfig(namespace string) *storage.Config {
	return GetConfigAndInit(namespace)
}

//GetConfigAndInit 根据namespace获取apollo配置
func GetConfigAndInit(namespace string) *storage.Config {
	if namespace == "" {
		return nil
	}

	config, ok := storage.GetApolloConfigCache().Load(namespace)

	if !ok {
		//init cache
		storage.CreateNamespaceConfig(namespace)

		//sync config
		notify.SyncNamespaceConfig(namespace)
	}

	config, ok = storage.GetApolloConfigCache().Load(namespace)

	if !ok {
		return nil
	}

	return config.(*storage.Config)
}

//GetConfigCache 根据namespace获取apollo配置的缓存
func GetConfigCache(namespace string) agcache.CacheInterface {
	config := GetConfigAndInit(namespace)
	if config == nil {
		return nil
	}

	return config.GetCache()
}

//GetDefaultConfigCache 获取默认缓存
func GetDefaultConfigCache() agcache.CacheInterface {
	config := GetConfigAndInit(storage.GetDefaultNamespace())
	if config != nil {
		return config.GetCache()
	}
	return nil
}

//GetApolloConfigCache 获取默认namespace的apollo配置
func GetApolloConfigCache() agcache.CacheInterface {
	return GetDefaultConfigCache()
}

//GetValue 获取配置
func GetValue(key string) string {
	value := getConfigValue(key)
	if value == nil {
		return utils.Empty
	}

	return value.(string)
}

//GetStringValue 获取string配置值
func GetStringValue(key string, defaultValue string) string {
	value := GetValue(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

//GetIntValue 获取int配置值
func GetIntValue(key string, defaultValue int) int {
	value := GetValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

//GetFloatValue 获取float配置值
func GetFloatValue(key string, defaultValue float64) float64 {
	value := GetValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

//GetBoolValue 获取bool 配置值
func GetBoolValue(key string, defaultValue bool) bool {
	value := GetValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}

//GetStringSliceValue 获取[]string 配置值
func GetStringSliceValue(key string, defaultValue []string) []string {
	value := getConfigValue(key)

	if value == nil {
		return defaultValue
	}
	s, ok := value.([]string)
	if !ok {
		return defaultValue
	}
	return s
}

//GetIntSliceValue 获取[]int 配置值
func GetIntSliceValue(key string, defaultValue []int) []int {
	value := getConfigValue(key)

	if value == nil {
		return defaultValue
	}
	s, ok := value.([]int)
	if !ok {
		return defaultValue
	}
	return s
}

func getConfigValue(key string) interface{} {
	cache := GetDefaultConfigCache()
	if cache == nil {
		return utils.Empty
	}

	value, err := cache.Get(key)
	if err != nil {
		log.Errorf("get config value fail!key:%s,err:%s", key, err)
		return utils.Empty
	}

	return value
}
