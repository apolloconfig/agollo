package agollo

import (
	"github.com/zouyx/agollo/agcache"
)

func init() {
	//init config
	initFileConfig()

	initCommon()
}

func initCommon()  {
	initDefaultConfig()

	initAllNotifications()
}

//InitCustomConfig init config by custom
func InitCustomConfig(loadAppConfig func() (*AppConfig, error)) {

	initConfig(loadAppConfig)

	initCommon()
}

//start apollo
func Start() error {
	return startAgollo()
}

//SetLogger 设置自定义logger组件
func SetLogger(loggerInterface LoggerInterface) {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}
}

//SetCache 设置自定义cache组件
func SetCache(cacheFactory *agcache.DefaultCacheFactory) {
	if cacheFactory != nil {
		initConfigCache(cacheFactory)
	}
}

//StartWithLogger 通过自定义logger启动agollo
func StartWithLogger(loggerInterface LoggerInterface) error {
	SetLogger(loggerInterface)
	return startAgollo()
}

//StartWithCache 通过自定义cache启动agollo
func StartWithCache(cacheFactory *agcache.DefaultCacheFactory) error {
	SetCache(cacheFactory)
	return startAgollo()
}

func startAgollo() error {
	//init server ip list
	go initServerIpList()
	//first sync
	err := notifySyncConfigServices()
	logger.Debug("init notifySyncConfigServices finished")

	//first sync fail then load config file
	if err != nil {
		splitNamespaces(appConfig.NamespaceName, func(namespace string) {
			config, _ := loadConfigFile(appConfig.BackupConfigPath, namespace)
			if config != nil {
				updateApolloConfig(config, false)
			}
		})
	}

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})

	logger.Info("agollo start finished , error:", err)

	return err
}
