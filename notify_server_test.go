package agollo

import (
	"net/http"
	"fmt"
	"time"
)

const responseStr  =`[{"namespaceName":"application","notificationId":%d}]`

//run mock notify server
func runMockNotifyServer(handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/notifications/v2", handler)

	appConfig:=GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}

	logger.Info("runMockNotifyServer:",appConfig.Ip)
	err:=http.ListenAndServe(fmt.Sprintf("%s",appConfig.Ip), nil)
	if err!=nil{
		logger.Error("runMockConfigServer err:",err)
	}
}

func closeMockNotifyServer() {
	http.DefaultServeMux=&http.ServeMux{}
}

var normalNotifyCount=1

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func normalResponse(rw http.ResponseWriter, req *http.Request) {
	normalNotifyCount++
	var result string
	if normalNotifyCount%3==0 {
		result = fmt.Sprintf(responseStr, normalNotifyCount)
		fmt.Fprintf(rw, "%s", result)
	}else {
		time.Sleep(5 * time.Second)
		rw.WriteHeader(http.StatusNotModified)
	}
}

func onlyNormalResponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseStr, 3)
	fmt.Fprintf(rw, "%s", result)
}

//Error response
//will hold 5s and keep response 404
func errorResponse(rw http.ResponseWriter, req *http.Request) {
	time.Sleep(1 * time.Second)
	rw.WriteHeader(http.StatusNotFound)
}