package timer

import (
	"time"
	"github.com/zouyx/agollo/config"
	"net/http"
	"io/ioutil"
	"github.com/cihub/seelog"
	"encoding/json"
	"github.com/zouyx/agollo/repository"
	"github.com/zouyx/agollo/utils/https"
)

type AutoRefreshConfigComponent struct {

}

func (this *AutoRefreshConfigComponent) Start()  {
	t2 := time.NewTimer(config.REFRESH_INTERVAL)
	for {
		select {
		case <-t2.C:
			syncConfigServices()
			t2.Reset(config.REFRESH_INTERVAL)
		}
	}
}

func SyncConfig() error {
	return syncConfigServices()
}

func syncConfigServices() error {
	client := &http.Client{
		Timeout:config.CONNECT_TIMEOUT,
	}
	if config.AppConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=config.GetConfigUrl()

	retry:=0
	var responseBody []byte
	var err error
	var res *http.Response
	for{
		retry++

		if retry>config.MAX_RETRIES{
			break
		}

		res,err=client.Get(url)

		if err != nil || res.StatusCode != https.SUCCESS{
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			if res!=nil{
				seelog.Error("Connect Apollo Server Fail,StatusCode:",res.StatusCode)
			}
			// if error then sleep
			time.Sleep(config.ON_ERROR_RETRY_INTERVAL)
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


