package agollo

//start apollo
func Start() error {
	return StartWithLogger(nil)
}

func StartWithLogger(loggerInterface LoggerInterface) error {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}

  //init server ip list
  go initServerIpList()

	//first sync
	err := notifySyncConfigServices()

	//first sync fail then load config file
	if err !=nil{
		config, _ := loadConfigFile(appConfig.BackupConfigPath)
		if config!=nil{
			updateApolloConfig(config,false)
		}
	}

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})

	logger.Info("agollo start finished , error:",err)
	
	return err
}
