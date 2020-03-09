package agollo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/zouyx/agollo/v3/env/config"
)

const (
	configSecondResponseStr = `{
  "appId": "100004459",
  "cluster": "default",
  "namespaceName": "abc1",
  "configurations": {
    "key1-1":"value1-1",
    "key1-2":"value2-1"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	configResponseStr = `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`
	responseStr    = `[{"namespaceName":"application","notificationId":%d}]`
	responseTwoStr = `[{"namespaceName":"application","notificationId":%d},{"namespaceName":"abc1","notificationId":%d}]`
)

//run mock config server
func runMockConfigServer(handlerMap map[string]func(http.ResponseWriter, *http.Request),
	notifyHandler func(http.ResponseWriter, *http.Request),
	appConfig *config.AppConfig) *httptest.Server {
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

//Error response
//will hold 5s and keep response 404
func runErrorResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	return ts
}

func onlyNormalTwoResponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseTwoStr, 3, 3)
	fmt.Fprintf(rw, "%s", result)
}

func onlyNormalConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configResponseStr)
}

func onlyNormalSecondConfigResponse(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, configSecondResponseStr)
}

func onlyNormalResponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseStr, 3)
	fmt.Fprintf(rw, "%s", result)
}
