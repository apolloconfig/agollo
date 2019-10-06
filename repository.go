package agollo

import (
	"github.com/zouyx/agollo/agcache"
	"strconv"
	"sync"
)

const (
	empty = ""

	//1 minute
	configCacheExpireTime = 120

	defaultNamespace="DEFAULT"
)


var (
	currentConnApolloConfig = &currentApolloConfig{}

	//config from apollo
	apolloConfigCache = agcache.DefaultCacheFactory{}.Create()

	apolloConfigLocalCache = make(map[string]*Config,0)
)

func init() {
	s, i := createDefaultConfig()
	apolloConfigLocalCache[s]=i
}

func initCache(cacheInterface agcache.CacheInterface)  {
	apolloConfigCache=cacheInterface
}

func createDefaultConfig() (string,*Config) {
	c:=&Config{
		namespace:defaultNamespace,
		cache:agcache.DefaultCacheFactory{}.Create(),
	}

	return c.namespace,c
}

type currentApolloConfig struct {
	l      sync.RWMutex
	config *ApolloConnConfig
}

type Config struct {
	namespace string
	cache agcache.CacheInterface
}

func (this *Config) getConfigValue(key string) interface{} {
	value, err := this.cache.Get([]byte(key))
	if err != nil {
		logger.Errorf("get config value fail!key:%s,err:%s", key, err)
		return empty
	}

	return string(value)
}

func (this *Config) getValue(key string) string {
	value := getConfigValue(key)
	if value == nil {
		return empty
	}

	return value.(string)
}

func (this *Config) GetStringValue(key string, defaultValue string) string {
	value := getValue(key)
	if value == empty {
		return defaultValue
	}

	return value
}

func (this *Config) GetIntValue(key string, defaultValue int) int {
	value := getValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		logger.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

func (this *Config) GetFloatValue(key string, defaultValue float64) float64 {
	value := getValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logger.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

func (this *Config) GetBoolValue(key string, defaultValue bool) bool {
	value := getValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		logger.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}

func updateApolloConfig(apolloConfig *ApolloConfig, isBackupConfig bool) {
	if apolloConfig == nil {
		logger.Error("apolloConfig is null,can't update!")
		return
	}
	//get change list
	changeList := updateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime)

	if len(changeList) > 0 {
		//create config change event base on change list
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName)

		//push change event to channel
		pushChangeEvent(event)
	}

	//update apollo connection config
	currentConnApolloConfig.l.Lock()
	defer currentConnApolloConfig.l.Unlock()

	currentConnApolloConfig.config = &apolloConfig.ApolloConnConfig

	if isBackupConfig {
		//write config file async
		go writeConfigFile(apolloConfig, appConfig.getBackupConfigPath())
	}
}

func updateApolloConfigCache(configurations map[string]string, expireTime int) map[string]*ConfigChange {
	if (configurations == nil || len(configurations) == 0) && apolloConfigCache.EntryCount() == 0 {
		return nil
	}

	//get old keys
	mp := map[string]bool{}
	it := apolloConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		mp[string(en.Key)] = true
	}

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
				oldValue, _ := apolloConfigCache.Get([]byte(key))
				if string(oldValue) != value {
					changes[key] = createModifyConfigChange(string(oldValue), value)
				}
			}

			apolloConfigCache.Set([]byte(key), []byte(value), expireTime)
			delete(mp, string(key))
		}
	}

	// remove del keys
	for key := range mp {
		//get old value and del
		oldValue, _ := apolloConfigCache.Get([]byte(key))
		changes[key] = createDeletedConfigChange(string(oldValue))

		apolloConfigCache.Del([]byte(key))
	}

	return changes
}

//base on changeList create Change event
func createConfigChangeEvent(changes map[string]*ConfigChange, nameSpace string) *ChangeEvent {
	return &ChangeEvent{
		Namespace: nameSpace,
		Changes:   changes,
	}
}

func touchApolloConfigCache() error {
	updateApolloConfigCacheTime(configCacheExpireTime)
	return nil
}

func updateApolloConfigCacheTime(expireTime int) {
	it := apolloConfigCache.NewIterator()
	for i := int64(0); i < apolloConfigCache.EntryCount(); i++ {
		entry := it.Next()
		if entry == nil {
			break
		}
		apolloConfigCache.Set([]byte(entry.Key), []byte(entry.Value), expireTime)
	}
}

func GetApolloConfigCache() agcache.CacheInterface {
	return apolloConfigCache
}

func GetCurrentApolloConfig() *ApolloConnConfig {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()

	return currentConnApolloConfig.config

}

func getConfigValue(key string) interface{} {
	value, err := apolloConfigCache.Get([]byte(key))
	if err != nil {
		logger.Errorf("get config value fail!key:%s,err:%s", key, err)
		return empty
	}

	return string(value)
}

func getValue(key string) string {
	value := getConfigValue(key)
	if value == nil {
		return empty
	}

	return value.(string)
}

func GetStringValue(key string, defaultValue string) string {
	value := getValue(key)
	if value == empty {
		return defaultValue
	}

	return value
}

func GetIntValue(key string, defaultValue int) int {
	value := getValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		logger.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

func GetFloatValue(key string, defaultValue float64) float64 {
	value := getValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logger.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

func GetBoolValue(key string, defaultValue bool) bool {
	value := getValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		logger.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}