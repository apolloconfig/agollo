/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http

import (
	"net/http"
	"net/http/httptest"
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

var IP1 = "localhost:7080"
var IP2 = "localhost:7081"

var servicesResponseStr = `[{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.102:apollo-configservice:8080",
"homepageUrl": "http://` + IP1 + `/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.125:apollo-configservice:8080",
"homepageUrl": "http://` + IP2 + `/"
}]`

var normalBackupConfigCount = 0

//Normal response
//First request will hold 5s and response http.StatusNotModified
//Second request will hold 5s and response http.StatusNotModified
//Second request will response [{"namespaceName":"application","notificationId":3}]
func runNormalBackupConfigResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalBackupConfigCount++
		if normalBackupConfigCount%3 == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(configResponseStr))
		} else {
			time.Sleep(500 * time.Microsecond)
			w.WriteHeader(http.StatusBadGateway)
		}
	}))

	return ts
}

func runNormalBackupConfigResponseWithHTTPS() *httptest.Server {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		normalBackupConfigCount++
		if normalBackupConfigCount%3 == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(configResponseStr))
		} else {
			time.Sleep(500 * time.Microsecond)
			w.WriteHeader(http.StatusBadGateway)
		}
	}))

	return ts
}

//wait long time then response
func runLongTimeResponse() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(configResponseStr))
	}))

	return ts
}
