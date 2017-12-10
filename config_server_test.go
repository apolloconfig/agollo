package agollo

import (
	"net/http"
	"fmt"
	"time"
)

const configResponseStr  =`{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

const configChangeResponseStr  =`{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2",
    "string":"string"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

//run mock config server
func runMockConfigServer(handler func(http.ResponseWriter, *http.Request)) {
	appConfig:=GetAppConfig()
	uri:=fmt.Sprintf("/configs/%s/%s/%s",appConfig.AppId,appConfig.Cluster,appConfig.NamespaceName)
	http.HandleFunc(uri, handler)

	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}

	err:=http.ListenAndServe(fmt.Sprintf("%s",appConfig.Ip), nil)
	if err!=nil{
		logger.Error("runMockConfigServer err:",err)
	}
}

func closeMockConfigServer() {
	http.DefaultServeMux=&http.ServeMux{}
}

var normalConfigCount=1

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func normalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	normalConfigCount++
	if normalConfigCount%3==0 {
		fmt.Fprintf(rw, configResponseStr)
	}else {
		time.Sleep(500 * time.Microsecond)
		rw.WriteHeader(http.StatusNotModified)
	}
}

func changeConfigResponse(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, configChangeResponseStr)
}

func onlyNormalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, configResponseStr)
}

func notModifyConfigResponse(rw http.ResponseWriter, req *http.Request) {
	time.Sleep(800 * time.Microsecond)
	rw.WriteHeader(http.StatusNotModified)
}

//Error response
//will hold 5s and keep response 404
func errorConfigResponse(rw http.ResponseWriter, req *http.Request) {
	time.Sleep(500 * time.Microsecond)
	rw.WriteHeader(http.StatusNotFound)
}