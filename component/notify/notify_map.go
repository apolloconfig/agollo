package notify

import (
	"sync"
	"encoding/json"
	"github.com/zouyx/agollo/config"
)

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
	appConfig:=config.GetAppConfig()

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
