package agollo

//start apollo
func Start() {
	StartWithLogger(nil)
}

func StartWithLogger(loggerInterface LoggerInterface) {
	if loggerInterface != nil {
		initLogger(loggerInterface)
	}

	//first sync
	notifySyncConfigServices()

	//start auto refresh config
	go StartRefreshConfig(&AutoRefreshConfigComponent{})

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})
}

func StartWithConfig(config *AppConfig) {
	initConfig(func() (*AppConfig, error) {
		return config, nil
	})
	StartWithLogger(nil)
}

func StartWithConfigFile(fileName string) {
	initConfig(func() (*AppConfig, error) {
		return loadJsonConfig(fileName)
	})
	StartWithLogger(nil)
}