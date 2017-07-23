package agollo

//start apollo
func Start() {
	//start auto refresh config
	go StartRefreshConfig(&AutoRefreshConfigComponent{})

	//start long poll sync config
	go StartRefreshConfig(&NotifyConfigComponent{})
}