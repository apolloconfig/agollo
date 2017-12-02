package agollo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var IP1="localhost:7080"
var IP2="localhost:7081"

var servicesResponseStr = `[{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.102:apollo-configservice:8080",
"homepageUrl": "http://`+IP1+`/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.125:apollo-configservice:8080",
"homepageUrl": "http://`+IP2+`/"
}]`

//run mock config server
func runMockServicesServer(handler func(http.ResponseWriter, *http.Request)) {
	tmpServerInfo := make([]*serverInfo, 0)

	json.Unmarshal([]byte(servicesResponseStr), &tmpServerInfo)

	uri := fmt.Sprintf("/services/config")
	http.HandleFunc(uri, handler)

	http.ListenAndServe(fmt.Sprintf("%s", appConfig.Ip), nil)
}

func closeMockServicesServer() {
	http.DefaultServeMux = &http.ServeMux{}
}

//Normal response
func normalServicesResponse(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, servicesResponseStr)
}


//run mock config Real And Backup server
func runMockConfigBackupServer(handler func(http.ResponseWriter, *http.Request)) {
	mockConfigServerListen(IP1,handler)
	//mockConfigServerListen(IP2,handler)
}

func mockConfigServerListen(ip string,handler func(http.ResponseWriter, *http.Request)) {
	appConfig:=GetAppConfig()
	uri:=fmt.Sprintf("/configs/%s/%s/%s",appConfig.AppId,appConfig.Cluster,appConfig.NamespaceName)
	http.HandleFunc(uri, handler)

	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}

	logger.Info("mockConfigServerListen:",appConfig.Ip)
	err:=http.ListenAndServe(fmt.Sprintf("%s",ip), nil)
	if err!=nil{
		logger.Error("runMockConfigServer err:",err)
	}
}

func closeAllMockServicesServer() {
	http.DefaultServeMux = &http.ServeMux{}
}

var normalBackupConfigCount=0

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func normalBackupConfigResponse(rw http.ResponseWriter, req *http.Request) {
	normalBackupConfigCount++
	if normalBackupConfigCount%3==0 {
		fmt.Fprintf(rw, configResponseStr)
	}else {
		time.Sleep(500 * time.Microsecond)
		rw.WriteHeader(http.StatusBadGateway)
	}
}

//wait long time then response
func longTimeResponse(rw http.ResponseWriter, req *http.Request) {
	time.Sleep(10*time.Second);
	fmt.Fprintf(rw, configResponseStr)
}
//
//func errorBackupConfigResponse(rw http.ResponseWriter, req *http.Request) {
//	rw.WriteHeader(http.StatusBadGateway)
//}