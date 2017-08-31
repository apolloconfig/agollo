package agollo

//start apollo
func Start() {
	//first sync
	notifySyncConfigServices()

	//start auto refresh config
	go StartRefreshConfig(&AutoRefreshConfigComponent{})

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})
}

func StartWithConfig(config *AppConfig) {
	//init common
	initCommon()

	appConfig = config

	//init config
	go func(appConfig *AppConfig) {
		apolloConfig := &ApolloConfig{}
		apolloConfig.AppId = appConfig.AppId
		apolloConfig.Cluster = appConfig.Cluster
		apolloConfig.NamespaceName = appConfig.NamespaceName

		updateApolloConfig(apolloConfig)
	}(config)

	// notify init
	initNotify()

	//first sync
	notifySyncConfigServices()

	//start auto refresh config
	go StartRefreshConfig(&AutoRefreshConfigComponent{})

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})
}
