package agollo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

const responseStr = `[{"namespaceName":"application","notificationId":%d}]`
const responseTwoStr = `[{"namespaceName":"application","notificationId":%d},{"namespaceName":"abc1","notificationId":%d}]`

var normalNotifyCount = 1

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func runNormalResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		normalNotifyCount++
		if normalNotifyCount%3 == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(responseStr, normalNotifyCount)))
		} else {
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

func onlyNormalTwoResponse(rw http.ResponseWriter, req *http.Request) {
	result := fmt.Sprintf(responseTwoStr, 3,3)
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
