package agollo

import "github.com/zouyx/agollo/component/timer"

//start apollo
func Start() {
	//start auto refresh config
	go timer.StartAutoRefreshConfig()
}
