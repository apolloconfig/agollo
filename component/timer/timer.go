package timer

import (
	"time"
	"github.com/zouyx/agollo/config"
	"fmt"
)

type AutoRefreshConfigComponent struct {

}

func (this *AutoRefreshConfigComponent) Start()  {
	t2 := time.NewTimer(config.REFRESH_INTERVAL*config.REFRESH_INTERVAL_TIME_UNIT)
	for {
		select {
		case <-t2.C:
			fmt.Println(config.REFRESH_INTERVAL,"s timer")
			t2.Reset(config.REFRESH_INTERVAL*config.REFRESH_INTERVAL_TIME_UNIT)
		}
	}
}


