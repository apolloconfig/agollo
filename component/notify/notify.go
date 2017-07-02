package notify

import (
	"time"
	"github.com/zouyx/agollo/config"
	"github.com/cihub/seelog"
	"net/http"
	"github.com/zouyx/agollo/utils/https"
	"io/ioutil"
	"github.com/zouyx/agollo/dto"
	"encoding/json"
	"github.com/zouyx/agollo/component/timer"
)

type NotifyConfigComponent struct {

}

func (this *NotifyConfigComponent) Start()  {
	t2 := time.NewTimer(config.LONG_POLL_INTERVAL)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			go syncConfigServices()
			t2.Reset(config.LONG_POLL_INTERVAL)
		}
	}
}

func toApolloConfig(resBody []byte) ([]*dto.ApolloNotify,error) {
	remoteConfig:=make([]*dto.ApolloNotify,0)

	err:=json.Unmarshal(resBody,&remoteConfig)

	if err!=nil{
		seelog.Error("Unmarshal Msg Fail,Error:",err)
		return nil,err
	}
	return remoteConfig,nil
}

func getRemoteConfig() ([]*dto.ApolloNotify,error) {
	client := &http.Client{
		Timeout:config.CONNECT_TIMEOUT,

	}
	if config.AppConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=config.GetNotifyUrl(allNotifications.getNotifies())

	seelog.Debugf("sync config url:%s",url)
	seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

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

		if res.StatusCode==https.NOT_MODIFIED {
			seelog.Warn("Config Not Modified:",err)
			return nil,nil
		}

		responseBody, err = ioutil.ReadAll(res.Body)
		if err!=nil{
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}
	}

	if err !=nil {
		seelog.Error("Over Max Retry Still Error,Error:",err)
		return nil,err
	}

	return toApolloConfig(responseBody)
}

func syncConfigServices() error {

	remoteConfigs,err:=getRemoteConfig()

	if err!=nil||len(remoteConfigs)==0{
		return err
	}

	updateAllNotifications(remoteConfigs)

	//sync all config
	timer.SyncConfig()

	return nil
}

func updateAllNotifications(remoteConfigs []*dto.ApolloNotify) {
	for _,remoteConfig:=range remoteConfigs{
		if remoteConfig.NamespaceName==""{
			continue
		}

		allNotifications.setNotify(remoteConfig.NamespaceName,remoteConfig.NotificationId)
	}
}