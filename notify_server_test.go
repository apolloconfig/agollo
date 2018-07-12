package agollo

import (
	"net/http"
	"fmt"
	"time"
	"net/http/httptest"
)

const responseStr  =`[{"namespaceName":"application","notificationId":%d}]`

//run mock notify server
func runMockNotifyServer(handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/notifications/v2", handler)

	appConfig:=GetAppConfig(nil)
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}

	logger.Info("runMockNotifyServer:",appConfig.Ip)
	err:=http.ListenAndServe(fmt.Sprintf("%s",appConfig.Ip), nil)
	if err!=nil{
		logger.Error("runMockConfigServer err:",err)
	}
}

var normalNotifyCount=1

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func runNormalResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10*time.Second)
		normalNotifyCount++
		if normalNotifyCount%3==0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(responseStr, normalNotifyCount)))
		}else {
			time.Sleep(5 * time.Second)
			w.WriteHeader(http.StatusNotModified)
		}
	}))

	return ts
}

func onlyNormalResponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseStr, 3)
	fmt.Fprintf(rw, "%s", result)
}

//Error response
//will hold 5s and keep response 404
func runErrorResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	return ts
}