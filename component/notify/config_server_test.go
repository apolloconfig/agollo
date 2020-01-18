package notify

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/zouyx/agollo/v2/env"
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

const configAbc1ResponseStr = `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "abc1",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

//run mock config server
func runMockConfigServer(handlerMap map[string]func(http.ResponseWriter, *http.Request),
	notifyHandler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	appConfig := env.GetPlainAppConfig()
	uriHandlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 0)
	for namespace, handler := range handlerMap {
		uri := fmt.Sprintf("/configs/%s/%s/%s", appConfig.AppID, appConfig.Cluster, namespace)
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

//Error response
//will hold 5s and keep response 404
func runErrorResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
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
