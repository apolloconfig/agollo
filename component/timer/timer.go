package timer

import (
	"time"
	"github.com/zouyx/agollo/config"
	"fmt"
	//"net/http"
	"net/http"
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
			updateConfigServices()
			t2.Reset(config.REFRESH_INTERVAL)
		}
	}
}

func StartAutoRefreshConfig()  {
	auto:=&AutoRefreshConfigComponent{}
	auto.Start()
}

func updateConfigServices()  {
	client := &http.Client{
		Timeout:config.CONNECT_TIMEOUT,
	}
	if config.ApolloConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=fmt.Sprint("http://10.16.4.193:8080/configfiles/json/%s/%s/%s",
		config.ApolloConfig.AppId,
		config.ApolloConfig.Cluster,
		config.ApolloConfig.NamespaceName)
	client.Get(url)
}


