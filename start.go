package agollo

import "github.com/zouyx/agollo/agcache"

//start apollo
func Start() error {
	return startAgollo()
}

//SetLogger 设置自定义logger组件
func SetLogger(loggerInterface LoggerInterface)  {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}
}

//SetCache 设置自定义cache组件
func SetCache(cacheInterface agcache.CacheInterface)  {
	if cacheInterface != nil {
		initConfigCache(cacheInterface)
	}
}

//StartWithLogger 通过自定义logger启动agollo
func StartWithLogger(loggerInterface LoggerInterface) error {
	SetLogger(loggerInterface)
	return startAgollo()
}

//StartWithCache 通过自定义cache启动agollo
func StartWithCache(cacheInterface agcache.CacheInterface) error {
	SetCache(cacheInterface)
	return startAgollo()
}

func startAgollo() error {
	//init server ip list
	go initServerIpList()

	//first sync
	err := notifySyncConfigServices()

	//first sync fail then load config file
	if err != nil {
		splitNamespaces(appConfig.NamespaceName, func(namespace string) {
			config, _ := loadConfigFile(appConfig.BackupConfigPath,namespace)
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
