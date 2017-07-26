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

//run mock config server
func runMockConfigServer(handler func(http.ResponseWriter, *http.Request)) {
	appConfig:=GetAppConfig()
	uri:=fmt.Sprintf("/configs/%s/%s/%s",appConfig.AppId,appConfig.Cluster,appConfig.NamespaceName)
	http.HandleFunc(uri, handler)

	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}

	http.ListenAndServe(fmt.Sprintf("%s",appConfig.Ip), nil)
}

func closeMockConfigServer() {
	http.DefaultServeMux=&http.ServeMux{}
}


//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func normalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	i++
	if i%3==0 {
		fmt.Fprintf(rw, configResponseStr)
	}else {
		time.Sleep(500 * time.Microsecond)
		rw.WriteHeader(http.StatusNotModified)
	}
}

//Error response
//will hold 5s and keep response 404
func errorConfigResponse(rw http.ResponseWriter, req *http.Request) {
	time.Sleep(500 * time.Microsecond)
	rw.WriteHeader(http.StatusNotFound)
}