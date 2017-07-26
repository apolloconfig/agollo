package agollo

//start apollo
func Start() {
	//first sync
	syncConfigServices()

	//start auto refresh config
	go StartRefreshConfig(&AutoRefreshConfigComponent{})

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})
}