package agollo

import (
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"sync"
	"errors"
	"github.com/cihub/seelog"
)


const(
	default_notification_id=-1
)

var(
	allNotifications *notificationsMap
)

type NotifyConfigComponent struct {

}

type apolloNotify struct {
	NotificationId int64 `json:"notificationId"`
	NamespaceName string `json:"namespaceName"`
}


func (this *NotifyConfigComponent) Start()  {
	t2 := time.NewTimer(long_poll_interval)
	//long poll for sync
	for {
		select {
		case <-t2.C:
			notifySyncConfigServices()
			t2.Reset(long_poll_interval)
		}
	}
}

func toApolloConfig(resBody []byte) ([]*apolloNotify,error) {
	remoteConfig:=make([]*apolloNotify,0)

	err:=json.Unmarshal(resBody,&remoteConfig)

	if err!=nil{
		seelog.Error("Unmarshal Msg Fail,Error:",err)
		return nil,err
	}
	return remoteConfig,nil
}

func getRemoteConfig() ([]*apolloNotify,error) {
	client := &http.Client{
		Timeout:long_poll_connect_timeout,

	}
	appConfig:=GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=getNotifyUrl(allNotifications.getNotifies(),appConfig)

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

		if res==nil||err!=nil{
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err!=nil{
				seelog.Error("Read Apollo Server response Fail,Error:",err)
				continue
			}
			return toApolloConfig(responseBody)

		case http.StatusNotModified:
			seelog.Warn("Config Not Modified:", err)
			return nil, nil

		default:
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			if res!=nil{
				seelog.Error("Connect Apollo Server Fail,StatusCode:",res.StatusCode)
			}
			// if error then sleep
			time.Sleep(on_error_retry_interval)
			continue
		}
	}

	seelog.Error("Over Max Retry Still Error,Error:",err)
	if err==nil{
		err=errors.New("Over Max Retry Still Error!")
	}
	return nil,err
}

func notifySyncConfigServices() error {

	remoteConfigs,err:=getRemoteConfig()

	if err!=nil||len(remoteConfigs)==0{
		return err
	}

	updateAllNotifications(remoteConfigs)

	//sync all config
	SyncConfig()

	return nil
}

func updateAllNotifications(remoteConfigs []*apolloNotify) {
	for _,remoteConfig:=range remoteConfigs{
		if remoteConfig.NamespaceName==""{
			continue
		}

		allNotifications.setNotify(remoteConfig.NamespaceName,remoteConfig.NotificationId)
	}
}


func init()  {
	allNotifications=&notificationsMap{
		notifications:make(map[string]int64,1),
	}
	appConfig:=GetAppConfig()

	allNotifications.setNotify(appConfig.NamespaceName,default_notification_id)
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
