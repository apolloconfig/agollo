package notify

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

const responseStr = `[{"namespaceName":"application","notificationId":%d}]`

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
