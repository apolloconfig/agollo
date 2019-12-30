package agollo

import (
	"github.com/zouyx/agollo/v2/agcache"
	"github.com/zouyx/agollo/v2/component/notify"
	"github.com/zouyx/agollo/v2/storage"
	"github.com/zouyx/agollo/v2/utils"
	"strconv"
	. "github.com/zouyx/agollo/v2/component/log"
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
		storage.CreateNamespaceConfig(storage.GetDefaultCacheFactory(), namespace)

		notify.NotifySimpleSyncConfigServices(namespace)
	}

	if config, ok = storage.GetApolloConfigCache().Load(namespace); !ok {
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
	if !config.GetIsInit() {
		config.GetWaitInit().Wait()
	}

	return config.GetCache()
}


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


func GetValue(key string) string {
	value := getConfigValue(key)
	if value == nil {
		return utils.Empty
	}

	return value.(string)
}

func GetStringValue(key string, defaultValue string) string {
	value := GetValue(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

func GetIntValue(key string, defaultValue int) int {
	value := GetValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		Logger.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

func GetFloatValue(key string, defaultValue float64) float64 {
	value := GetValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		Logger.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

func GetBoolValue(key string, defaultValue bool) bool {
	value := GetValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		Logger.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}


func getConfigValue(key string) interface{} {
	value, err := GetDefaultConfigCache().Get(key)
	if err != nil {
		Logger.Errorf("get config value fail!key:%s,err:%s", key, err)
		return utils.Empty
	}

	return string(value)
}