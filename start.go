package agollo

//start apollo
func Start() error {
	return StartWithLogger(nil)
}

func StartWithLogger(loggerInterface LoggerInterface) error {
	return StartWithParams(loggerInterface,nil)
}

func StartWithCache(cacheInterface CacheInterface) error {
	return StartWithParams(nil,cacheInterface)
}

func StartWithParams(loggerInterface LoggerInterface,cacheInterface CacheInterface) error {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}
	if cacheInterface != nil {
		initCache(cacheInterface)
	}

	//init server ip list
	go initServerIpList()

	//first sync
	err := notifySyncConfigServices()

	//first sync fail then load config file
	if err != nil {
		config, _ := loadConfigFile(appConfig.BackupConfigPath)
		if config != nil {
			updateApolloConfig(config, false)
		}
	}

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})

	logger.Info("agollo start finished , error:", err)

	return err
}
