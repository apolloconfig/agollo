package timer

import (
	"time"
	"github.com/zouyx/agollo/config"
	"net/http"
	"io/ioutil"
	"github.com/cihub/seelog"
	"github.com/zouyx/agollo/utils/https"
	"errors"
	"github.com/zouyx/agollo/dto"
	"github.com/zouyx/agollo/repository"
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

	appConfig:=config.GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=config.GetConfigUrl(appConfig)
	seelog.Debug("url:",url)

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

	if responseBody==nil{
		return errors.New("response body is null!")
	}

	apolloConfig,err:=dto.CreateApolloConfigWithJson(responseBody)

	if err!=nil{
		seelog.Error("Unmarshal Msg Fail,Error:",err)
		return err
	}

	go updateAppConfig(apolloConfig)

	//repository.UpdateLocalConfigRepository(apolloConfig.Configurations)

	return nil
}

func updateAppConfig(apolloConfig *dto.ApolloConfig) {
	repository.UpdateApolloConfig(apolloConfig)
}