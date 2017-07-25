package agollo

import (
	"time"
	"net/http"
	"io/ioutil"
	"github.com/cihub/seelog"
	"errors"
)

type AutoRefreshConfigComponent struct {

}

func (this *AutoRefreshConfigComponent) Start()  {
	t2 := time.NewTimer(refresh_interval)
	for {
		select {
		case <-t2.C:
			syncConfigServices()
			t2.Reset(refresh_interval)
		}
	}
}

func SyncConfig() error {
	return autoSyncConfigServices()
}

func autoSyncConfigServices() error {
	client := &http.Client{
		Timeout:connect_timeout,
	}

	appConfig:=GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=GetConfigUrl(appConfig)
	seelog.Debug("url:",url)

	retry:=0
	var responseBody []byte
	var err error
	var res *http.Response
	for{
		retry++

		if retry>max_retries{
			break
		}

		res,err=client.Get(url)

		if err != nil || res.StatusCode != SUCCESS{
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			if res!=nil{
				seelog.Error("Connect Apollo Server Fail,StatusCode:",res.StatusCode)
			}
			// if error then sleep
			time.Sleep(ON_ERROR_RETRY_INTERVAL)
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

	if responseBody==nil{
		return errors.New("response body is null!")
	}

	apolloConfig,err:=CreateApolloConfigWithJson(responseBody)

	if err!=nil{
		seelog.Error("Unmarshal Msg Fail,Error:",err)
		return err
	}

	go updateAppConfig(apolloConfig)

	//repository.UpdateLocalConfigRepository(apolloConfig.Configurations)

	return nil
}

func updateAppConfig(apolloConfig *ApolloConfig) {
	UpdateApolloConfig(apolloConfig)
}