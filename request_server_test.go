package agollo

import (
			"net/http"
	"time"
	"net/http/httptest"
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

//Normal response
func runNormalServicesResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(servicesResponseStr))
	}))

	return ts
}

var normalBackupConfigCount=0

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func runNormalBackupConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalBackupConfigCount++
		if normalBackupConfigCount%3==0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(configResponseStr))
		}else {
			time.Sleep(500 * time.Microsecond)
			w.WriteHeader(http.StatusBadGateway)
		}
	}))

	return ts
}

//wait long time then response
func runLongTimeResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10*time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configResponseStr))
	}))

	return ts
}