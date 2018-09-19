package agollo

//start apollo
func Start() error {
	return StartWithLogger(nil)
}

func StartWithLogger(loggerInterface LoggerInterface) error {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}

	//first sync
	err := notifySyncConfigServices()

	//first sync fail then load config file
	if err !=nil{
		config, _ := loadConfigFile(appConfig.BackupConfigPath)
		if config!=nil{
			updateApolloConfig(config,false)
		}
	}

	//start auto refresh config
	go StartRefreshConfig(&AutoRefreshConfigComponent{})

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})
	
	return err
}
