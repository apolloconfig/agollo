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

package agollo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/zouyx/agollo/v4/env/config"
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

//run mock config Files server
func runMockConfigFilesServer(handlerMap map[string]func(http.ResponseWriter, *http.Request),
	notifyHandler func(http.ResponseWriter, *http.Request),
	appConfig *config.AppConfig) *httptest.Server {
	uriHandlerMap := make(map[string]func(http.ResponseWriter, *http.Request), 0)
	for namespace, handler := range handlerMap {
		uri := fmt.Sprintf("/configfiles/%s/%s/%s", appConfig.AppID, appConfig.Cluster, namespace)
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
