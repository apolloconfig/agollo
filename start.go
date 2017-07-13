package agollo

import (
	"github.com/zouyx/agollo/component/timer"
	"github.com/zouyx/agollo/component"
	"github.com/zouyx/agollo/component/notify"
)

//start apollo
func Start() {
	//start auto refresh config
	go component.StartRefreshConfig(&timer.AutoRefreshConfigComponent{})

	//start long poll sync config
	go component.StartRefreshConfig(&notify.NotifyConfigComponent{})
}