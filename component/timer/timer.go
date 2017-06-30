package timer

import (
	"time"
	"github.com/zouyx/agollo/config"
	"fmt"
	//"net/http"
	"net/http"
	"io/ioutil"
	"github.com/cihub/seelog"
	"encoding/json"
	"github.com/zouyx/agollo/repository"
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

func updateConfigServices() error {
	client := &http.Client{
		Timeout:config.CONNECT_TIMEOUT,
	}
	if config.ApolloConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=fmt.Sprintf("http://%s/configfiles/json/%s/%s/%s",
		config.ApolloConfig.Ip,
		config.ApolloConfig.AppId,
		config.ApolloConfig.Cluster,
		config.ApolloConfig.NamespaceName)

	retry:=0
	var responseBody []byte
	var err error
	var res *http.Response
	for{
		retry++

		if retry>MAX_RETRIES{
			break
		}

		res,err=client.Get(url)

		if err != nil {
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}
		responseBody, err = ioutil.ReadAll(res.Body)
		if err!=nil{
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}
	}

	if err !=nil {
		seelog.Error("Over Max Retry Still Error,Error:",err)
		return err
	}

	remoteConfig:=make(map[string]interface{})

	err=json.Unmarshal(responseBody,&remoteConfig)

	if err!=nil{
		seelog.Error("Unmarshal Msg Fail,Error:",err)
		return err
	}

	repository.UpdateConfig(remoteConfig)

	return nil
}


