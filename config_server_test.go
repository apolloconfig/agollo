package agollo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

const configResponseStr = `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

const configSecondResponseStr = `{
  "appId": "100004459",
  "cluster": "default",
  "namespaceName": "abc1",
  "configurations": {
    "key1-1":"value1-1",
    "key1-2":"value2-1"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

const configChangeResponseStr = `{
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
func runMockConfigServer(handlerMap map[string]func(http.ResponseWriter, *http.Request),
	notifyHandler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	appConfig := GetAppConfig(nil)
	uriHandlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 0)
	for namespace, handler := range handlerMap {
		uri := fmt.Sprintf("/configs/%s/%s/%s", appConfig.AppId, appConfig.Cluster, namespace)
		uriHandlerMap[uri] = handler
	}
	uriHandlerMap["/notifications/v2"] = notifyHandler

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		for path, handler := range uriHandlerMap {
			if strings.HasPrefix(uri, path) {
				handler(w, r)
				break
			}
		}
	}))

	return ts
}

var normalConfigCount = 1

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func runNormalConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalConfigCount++
		if normalConfigCount%3 == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(configResponseStr))
		} else {
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

func runChangeConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configChangeResponseStr))
	}))

	return ts
}

func onlyNormalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configResponseStr)
}

func onlyNormalSecondConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configSecondResponseStr)
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
