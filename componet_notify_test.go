package agollo

import (
	"testing"
	"time"
	"github.com/zouyx/agollo/test"
	"encoding/json"
)

func TestSyncConfigServices(t *testing.T) {
	notifySyncConfigServices()
}

func TestGetRemoteConfig(t *testing.T) {
	go runMockNotifyServer(normalResponse)
	defer closeMockNotifyServer()

	time.Sleep(1*time.Second)

	count:=1
	var remoteConfigs []*apolloNotify
	var err error
	for{
		count++
		remoteConfigs,err=getRemoteConfig()

		//err keep nil
		test.Nil(t,err)

		//if remote config is nil then break
		if remoteConfigs!=nil&&len(remoteConfigs)>0 {
			break
		}
	}

	test.Equal(t,count>1,true)
	test.Nil(t,err)
	test.NotNil(t,remoteConfigs)
	test.Equal(t,1,len(remoteConfigs))
	t.Log("remoteConfigs:",remoteConfigs)
	t.Log("remoteConfigs size:",len(remoteConfigs))

	notify:=remoteConfigs[0]

	test.Equal(t,"application",notify.NamespaceName)
	test.Equal(t,true,notify.NotificationId>0)
}

func TestErrorGetRemoteConfig(t *testing.T) {
	go runMockNotifyServer(errorResponse)
	defer closeMockNotifyServer()

	time.Sleep(1 * time.Second)

	var remoteConfigs []*apolloNotify
	var err error
	remoteConfigs, err = getRemoteConfig()

	test.NotNil(t, err)
	test.Nil(t, remoteConfigs)
	test.Equal(t, 0, len(remoteConfigs))
	t.Log("remoteConfigs:", remoteConfigs)
	t.Log("remoteConfigs size:", len(remoteConfigs))

	test.Equal(t,"Over Max Retry Still Error!",err.Error())
}

func TestUpdateAllNotifications(t *testing.T) {
	//clear
	allNotifications=&notificationsMap{
		notifications:make(map[string]int64,1),
	}
	notifyJson:=`[
  {
    "namespaceName": "application",
    "notificationId": 101
  }
]`
	notifies:=make([]*apolloNotify,0)

	err:=json.Unmarshal([]byte(notifyJson),&notifies)

	test.Nil(t,err)
	test.Equal(t,true,len(notifies)>0)

	updateAllNotifications(notifies)

	test.Equal(t,true,len(allNotifications.notifications)>0)
	test.Equal(t,int64(101),allNotifications.notifications["application"])
}


func TestUpdateAllNotificationsError(t *testing.T) {
	//clear
	allNotifications=&notificationsMap{
		notifications:make(map[string]int64,1),
	}

	notifyJson:=`ffffff`
	notifies:=make([]*apolloNotify,0)

	err:=json.Unmarshal([]byte(notifyJson),&notifies)

	test.NotNil(t,err)
	test.Equal(t,true,len(notifies)==0)

	updateAllNotifications(notifies)

	test.Equal(t,true,len(allNotifications.notifications)==0)
}

func TestToApolloConfigError(t *testing.T) {

	notified,err:=toApolloConfig([]byte("jaskldfjaskl"))
	test.Nil(t,notified)
	test.NotNil(t,err)
}
//
//func TestNotifyConfigComponent(t *testing.T) {
//	go func() {
//		for{
//			time.Sleep(5*time.Second)
//			fmt.Println(GetCurrentApolloConfig())
//		}
//	}()
//
//
//	c:=&NotifyConfigComponent{}
//	c.Start()
//
//}