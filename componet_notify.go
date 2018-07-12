package agollo

import (
	"time"
	"encoding/json"
	"sync"
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
		logger.Error("Unmarshal Msg Fail,Error:",err)
		return nil,err
	}
	return remoteConfig,nil
}

func getRemoteConfigSuccessCallBack(responseBody []byte)(o interface{},err error){
	return toApolloConfig(responseBody)
}

func getRemoteConfig(newAppConfig *AppConfig) ([]*apolloNotify,error) {
	appConfig:=GetAppConfig(newAppConfig)
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	urlSuffix:=getNotifyUrlSuffix(allNotifications.getNotifies(),appConfig,newAppConfig)

	//seelog.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

	notifies ,err:=requestRecovery(appConfig,&ConnectConfig{
		Uri:urlSuffix,
		Timeout:nofity_connect_timeout,
	},&CallBack{
		SuccessCallBack:getRemoteConfigSuccessCallBack,
		NotModifyCallBack:touchApolloConfigCache,
	})

	if notifies==nil{
		return nil,err
	}

	return notifies.([]*apolloNotify),err
}

func notifySyncConfigServices() error {

	remoteConfigs,err:=getRemoteConfig(nil)

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

func init() {
	initAllNotifications()
}

func initAllNotifications()  {
	appConfig:=GetAppConfig(nil)

	if appConfig!=nil {
		allNotifications = &notificationsMap{
			notifications: make(map[string]int64, 1),
		}

		allNotifications.setNotify(appConfig.NamespaceName, default_notification_id)
	}
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
