package agollo

import (
	"net/http"
	"fmt"
	"time"
	"net/http/httptest"
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
	appConfig:=GetAppConfig(nil)
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
func runNormalConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalConfigCount++
		if normalConfigCount%3==0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(configResponseStr))
		}else {
			time.Sleep(500 * time.Microsecond)
			w.WriteHeader(http.StatusNotModified)
		}
	}))

	return ts
}

func runLongNotmodifiedConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Microsecond)
		w.WriteHeader(http.StatusNotModified)
	}))

	return ts
}

func runChangeConfigResponse()*httptest.Server{
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configChangeResponseStr))
	}))

	return ts
}

func onlyNormalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, configResponseStr)
}

func runNotModifyConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(800 * time.Microsecond)
		w.WriteHeader(http.StatusNotModified)
	}))

	return ts
}

//Error response
//will hold 5s and keep response 404
func runErrorConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Microsecond)
		w.WriteHeader(http.StatusNotFound)
	}))

	return ts
}