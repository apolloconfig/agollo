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
			notifySyncConfigServices()
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

		if res==nil||err!=nil{
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err!=nil{
				seelog.Error("Connect Apollo Server Fail,Error:",err)
				continue
			}

			apolloConfig,err:=createApolloConfigWithJson(responseBody)

			if err!=nil{
				seelog.Error("Unmarshal Msg Fail,Error:",err)
				return err
			}

			updateApolloConfig(apolloConfig)

			return nil
		default:
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			if res!=nil{
				seelog.Error("Connect Apollo Server Fail,StatusCode:",res.StatusCode)
			}
			// if error then sleep
			time.Sleep(ON_ERROR_RETRY_INTERVAL)
			continue
		}
	}

	seelog.Error("Over Max Retry Still Error,Error:",err)
	if err==nil{
		err=errors.New("Over Max Retry Still Error!")
	}
	return err
	//repository.UpdateLocalConfigRepository(apolloConfig.Configurations)
}
