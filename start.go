package agollo

//start apollo
func Start() error {
	return startAgollo()
}

func SetLogger(loggerInterface LoggerInterface)  {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}
}

func SetCache(cacheInterface CacheInterface)  {
	if cacheInterface != nil {
		initCache(cacheInterface)
	}
}

func StartWithLogger(loggerInterface LoggerInterface) error {
	SetLogger(loggerInterface)
	return startAgollo()
}

func StartWithCache(cacheInterface CacheInterface) error {
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
