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
	error := notifySyncConfigServices()

	//start auto refresh config
	go StartRefreshConfig(&AutoRefreshConfigComponent{})

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})
	
	return error
}
