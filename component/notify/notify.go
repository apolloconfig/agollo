package notify

import (
	"time"
	"github.com/zouyx/agollo/config"
)

type NotifyConfigComponent struct {

}

func (this *NotifyConfigComponent) Start()  {
	t2 := time.NewTimer(config.REFRESH_INTERVAL)
	for {
		select {
		case <-t2.C:
			t2.Reset(config.REFRESH_INTERVAL)
		}
	}
}

func StartNotifyConfig()  {
	auto:=&NotifyConfigComponent{}
	auto.Start()
}
