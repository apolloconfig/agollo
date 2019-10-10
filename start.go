package agollo

import "github.com/zouyx/agollo/agcache"

//start apollo
func Start() error {
	return startAgollo()
}

func SetLogger(loggerInterface LoggerInterface)  {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}
}

func SetCache(cacheInterface agcache.CacheInterface)  {
	if cacheInterface != nil {
		initConfigCache(cacheInterface)
	}
}

func StartWithLogger(loggerInterface LoggerInterface) error {
	SetLogger(loggerInterface)
	return startAgollo()
}

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
