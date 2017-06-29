package timer

import (
	"time"
	"github.com/zouyx/agollo/config"
	"fmt"
	//"net/http"
)

const (
	//max retries connect apollo
	MAX_RETRIES=5
)

type AutoRefreshConfigComponent struct {

}

func (this *AutoRefreshConfigComponent) Start()  {
	t2 := time.NewTimer(config.REFRESH_INTERVAL)
	for {
		select {
		case <-t2.C:
			fmt.Println(config.REFRESH_INTERVAL,"s timer")
			t2.Reset(config.REFRESH_INTERVAL)
		}
	}
}

func StartAutoRefreshConfig()  {
	auto:=&AutoRefreshConfigComponent{}
	auto.Start()
}

func updateConfigServices()  {
	//client := &http.Client{
	//	Timeout:config.CONNECT_TIMEOUT,
	//}
	//resp, err := client.Get("")
}


