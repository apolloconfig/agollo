package agollo

import (
	"time"
	"github.com/cihub/seelog"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"sync"
)

type NotifyConfigComponent struct {

}

type ApolloNotify struct {
	NotificationId int64 `json:"notificationId"`
	NamespaceName string `json:"namespaceName"`
}


func (this *NotifyConfigComponent) Start()  {
	t2 := time.NewTimer(long_poll_interval)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			syncConfigServices()
			t2.Reset(long_poll_interval)
		}
	}
}

func toApolloConfig(resBody []byte) ([]*ApolloNotify,error) {
	remoteConfig:=make([]*ApolloNotify,0)

	err:=json.Unmarshal(resBody,&remoteConfig)

	if err!=nil{
		seelog.Error("Unmarshal Msg Fail,Error:",err)
		return nil,err
	}
	return remoteConfig,nil
}

func getRemoteConfig() ([]*ApolloNotify,error) {
	client := &http.Client{
		Timeout:long_poll_connect_timeout,

	}
	appConfig:=GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=GetNotifyUrl(allNotifications.getNotifies(),appConfig)

	seelog.Debugf("sync config url:%s",url)
	seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

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

		//not modified break
		if res.StatusCode==NOT_MODIFIED {
			seelog.Warn("Config Not Modified:",err)
			return nil,nil
		}

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
	SyncConfig()

	return nil
}

func updateAllNotifications(remoteConfigs []*ApolloNotify) {
	for _,remoteConfig:=range remoteConfigs{
		if remoteConfig.NamespaceName==""{
			continue
		}

		allNotifications.setNotify(remoteConfig.NamespaceName,remoteConfig.NotificationId)
	}
}


const(
	DEFAULT_NOTIFICATION_ID=-1
)

var(
	allNotifications *notificationsMap
)

func init()  {
	allNotifications=&notificationsMap{
		notifications:make(map[string]int64,1),
	}
	appConfig:=GetAppConfig()

	allNotifications.setNotify(appConfig.NamespaceName,DEFAULT_NOTIFICATION_ID)
}

type notification struct {
	NamespaceName string `json:"namespaceName"`
	NotificationId int64 `json:"notificationId"`
}

type notificationsMap struct {
	notifications map[string]int64
	sync.RWMutex
}

func (this *notificationsMap) setNotify(namespaceName string,notificationId int64) {
	this.Lock()
	defer this.Unlock()
	this.notifications[namespaceName]=notificationId
}
func (this *notificationsMap) getNotifies() string {
	this.RLock()
	defer this.RUnlock()

	notificationArr:=make([]*notification,0)
	for namespaceName,notificationId:=range this.notifications{
		notificationArr=append(notificationArr,
			&notification{
				NamespaceName:namespaceName,
				NotificationId:notificationId,
			})
	}

	j,err:=json.Marshal(notificationArr)

	if err!=nil{
		return ""
	}

	return string(j)
}
