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
