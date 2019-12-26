package storage

import (
	"net/http"
	"net/http/httptest"
)

const (
	configChangeResponseStr = `{
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
)

func runChangeConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configChangeResponseStr))
	}))

	return ts
}
