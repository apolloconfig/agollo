package storage

import (
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/zouyx/agollo/v2/agcache"
	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/utils"

	. "github.com/zouyx/agollo/v2/component/log"
)

//ConfigFileFormat 配置文件类型
type ConfigFileFormat string

const (
	//Properties
	Properties ConfigFileFormat = "properties"
	//XML
	XML ConfigFileFormat = "xml"
	//JSON
	JSON ConfigFileFormat = "json"
	//YML
	YML ConfigFileFormat = "yml"
	//YAML
	YAML ConfigFileFormat = "yaml"
)

const (

	//1 minute
	configCacheExpireTime = 120

	defaultNamespace = "application"
)

var (

	//config from apollo
	apolloConfigCache sync.Map
	//apolloConfigCache = make(map[string]*Config, 0)

	formatParser        = make(map[ConfigFileFormat]utils.ContentParser, 0)
	defaultFormatParser = &utils.DefaultParser{}

	cacheFactory = &agcache.DefaultCacheFactory{}
)

func init() {
	formatParser[Properties] = &utils.PropertiesParser{}
}

func InitDefaultConfig() {
	InitConfigCache(cacheFactory)
}

func InitConfigCache(cacheFactory *agcache.DefaultCacheFactory) {
	if env.GetPlainAppConfig() == nil {
		Logger.Warn("Config is nil,can not init agollo.")
		return
	}
	createNamespaceConfig(cacheFactory, env.GetPlainAppConfig().NamespaceName)
}

func createNamespaceConfig(cacheFactory *agcache.DefaultCacheFactory, namespace string) {
	env.SplitNamespaces(namespace, func(namespace string) {
		if _, ok := apolloConfigCache.Load(namespace); ok {
			return
		}
		c := &Config{
			namespace: namespace,
			cache:     cacheFactory.Create(),
		}
		c.isInit.Store(false)
		c.waitInit.Add(1)
		apolloConfigCache.Store(namespace, c)
	})
}

//Config apollo配置项
type Config struct {
	namespace string
	cache     agcache.CacheInterface
	isInit    atomic.Value
	waitInit  sync.WaitGroup
}

//getIsInit 获取标志
func (this *Config) getIsInit() bool {
	return this.isInit.Load().(bool)
}

//getConfigValue 获取配置值
func (this *Config) getConfigValue(key string) interface{} {
	b := this.getIsInit()
	if !b {
		this.waitInit.Wait()
	}
	if this.cache == nil {
		Logger.Errorf("get config value fail!namespace:%s is not exist!", this.namespace)
		return utils.Empty
	}

	value, err := this.cache.Get(key)
	if err != nil {
		Logger.Errorf("get config value fail!key:%s,err:%s", key, err)
		return utils.Empty
	}

	return string(value)
}

//GetValue 获取配置值（string）
func (this *Config) GetValue(key string) string {
	value := this.getConfigValue(key)
	if value == nil {
		return utils.Empty
	}

	return value.(string)
}

//GetStringValue 获取配置值（string），获取不到则取默认值
func (this *Config) GetStringValue(key string, defaultValue string) string {
	value := this.GetValue(key)
	if value == utils.Empty {
		return defaultValue
	}

	return value
}

//GetIntValue 获取配置值（int），获取不到则取默认值
func (this *Config) GetIntValue(key string, defaultValue int) int {
	value := this.GetValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		Logger.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

//GetFloatValue 获取配置值（float），获取不到则取默认值
func (this *Config) GetFloatValue(key string, defaultValue float64) float64 {
	value := this.GetValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		Logger.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

//GetBoolValue 获取配置值（bool），获取不到则取默认值
func (this *Config) GetBoolValue(key string, defaultValue bool) bool {
	value := this.GetValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		Logger.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}

//GetConfig 根据namespace获取apollo配置
func GetConfig(namespace string) *Config {
	return GetConfigAndInit(namespace)
}

//GetConfigAndInit 根据namespace获取apollo配置
func GetConfigAndInit(namespace string) *Config {
	if namespace == "" {
		return nil
	}

	config, ok := apolloConfigCache.Load(namespace)

	if !ok {
		createNamespaceConfig(cacheFactory, namespace)

		//notify.NotifySimpleSyncConfigServices(namespace)
	}

	if config, ok = apolloConfigCache.Load(namespace); !ok {
		return nil
	}

	return config.(*Config)
}

//GetConfigCache 根据namespace获取apollo配置的缓存
func GetConfigCache(namespace string) agcache.CacheInterface {
	config := GetConfigAndInit(namespace)
	if config == nil {
		return nil
	}
	if !config.getIsInit() {
		config.waitInit.Wait()
	}

	return config.cache
}

func GetDefaultConfigCache() agcache.CacheInterface {
	config := GetConfigAndInit(defaultNamespace)
	if config != nil {
		return config.cache
	}
	return nil
}

func UpdateApolloConfig(apolloConfig *env.ApolloConfig, isBackupConfig bool) {
	if apolloConfig == nil {
		Logger.Error("apolloConfig is null,can't update!")
		return
	}
	//get change list
	changeList := updateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime, apolloConfig.NamespaceName)

	if len(changeList) > 0 {
		//create config change event base on change list
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName)

		//push change event to channel
		pushChangeEvent(event)
	}

	//update apollo connection config
	env.SetCurrentApolloConfig(apolloConfig.NamespaceName, &apolloConfig.ApolloConnConfig)

	if isBackupConfig {
		//write config file async
		go env.WriteConfigFile(apolloConfig, env.GetPlainAppConfig().GetBackupConfigPath())
	}
}

func updateApolloConfigCache(configurations map[string]string, expireTime int, namespace string) map[string]*ConfigChange {
	config := GetConfig(namespace)
	if config == nil {
		return nil
	}

	isInit := false
	defer func(c *Config) {
		if !isInit {
			return
		}
		b := c.getIsInit()
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
				if string(oldValue) != value {
					changes[key] = createModifyConfigChange(string(oldValue), value)
				}
			}

			config.cache.Set(key, []byte(value), expireTime)
			delete(mp, string(key))
		}
	}

	// remove del keys
	for key := range mp {
		//get old value and del
		oldValue, _ := config.cache.Get(key)
		changes[key] = createDeletedConfigChange(string(oldValue))

		config.cache.Del(key)
	}
	isInit = true

	return changes
}

//base on changeList create Change event
func createConfigChangeEvent(changes map[string]*ConfigChange, nameSpace string) *ChangeEvent {
	return &ChangeEvent{
		Namespace: nameSpace,
		Changes:   changes,
	}
}

//GetApolloConfigCache 获取默认namespace的apollo配置
func GetApolloConfigCache() agcache.CacheInterface {
	return GetDefaultConfigCache()
}

func getConfigValue(key string) interface{} {
	value, err := GetDefaultConfigCache().Get(key)
	if err != nil {
		Logger.Errorf("get config value fail!key:%s,err:%s", key, err)
		return utils.Empty
	}

	return string(value)
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

//GetContent 获取配置文件内容
func (c *Config) GetContent(format ConfigFileFormat) string {
	parser := formatParser[format]
	if parser == nil {
		parser = defaultFormatParser
	}
	s, err := parser.Parse(c.cache)
	if err != nil {
		Logger.Debug("GetContent fail ! error:", err)
	}
	return s
}
