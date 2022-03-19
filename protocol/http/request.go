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
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	url2 "net/url"
	"strings"
	"sync"
	"time"

	"github.com/apolloconfig/agollo/v4/env/server"

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/utils"
)

var (
	//for on error retry
	onErrorRetryInterval = 2 * time.Second //2s

	connectTimeout = 1 * time.Second //1s

	//max retries connect apollo
	maxRetries = 5

	//defaultMaxConnsPerHost defines the maximum number of concurrent connections
	defaultMaxConnsPerHost = 512
	//defaultTimeoutBySecond defines the default timeout for http connections
	defaultTimeoutBySecond = 1 * time.Second
	//defaultKeepAliveSecond defines the connection time
	defaultKeepAliveSecond = 60 * time.Second
	// once for single http.Transport
	once sync.Once
	// defaultTransport http.Transport
	defaultTransport *http.Transport
)

func getDefaultTransport(insecureSkipVerify bool) *http.Transport {
	if defaultTransport == nil {
		once.Do(func() {
			defaultTransport = &http.Transport{
				MaxIdleConns:        defaultMaxConnsPerHost,
				MaxIdleConnsPerHost: defaultMaxConnsPerHost,
				DialContext: (&net.Dialer{
					KeepAlive: defaultKeepAliveSecond,
					Timeout:   defaultTimeoutBySecond,
				}).DialContext,
			}
			if insecureSkipVerify {
				defaultTransport.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: insecureSkipVerify,
				}
			}
		})
	}

	return defaultTransport
}

//CallBack 请求回调函数
type CallBack struct {
	SuccessCallBack   func([]byte, CallBack) (interface{}, error)
	NotModifyCallBack func() error
	AppConfigFunc     func() config.AppConfig
	Namespace         string
}

//Request 建立网络请求
func Request(requestURL string, connectionConfig *env.ConnectConfig, callBack *CallBack) (interface{}, error) {
	client := &http.Client{}
	//如有设置自定义超时时间即使用
	if connectionConfig != nil && connectionConfig.Timeout != 0 {
		client.Timeout = connectionConfig.Timeout
	} else {
		client.Timeout = connectTimeout
	}
	var err error
	url, err := url2.Parse(requestURL)
	if err != nil {
		log.Error("request Apollo Server url:%s, is invalid %s", requestURL, err)
		return nil, err
	}
	var insecureSkipVerify bool
	if strings.HasPrefix(url.Scheme, "https") {
		insecureSkipVerify = true
	}
	client.Transport = getDefaultTransport(insecureSkipVerify)
	retry := 0
	var retries = maxRetries
	if connectionConfig != nil && !connectionConfig.IsRetry {
		retries = 1
	}
	for {

		retry++

		if retry > retries {
			break
		}
		req, err := http.NewRequest("GET", requestURL, nil)
		if req == nil || err != nil {
			log.Errorf("Generate connect Apollo request Fail,url:%s,Error:%s", requestURL, err)
			// if error then sleep
			return nil, errors.New("generate connect Apollo request fail")
		}

		//增加header选项
		httpAuth := extension.GetHTTPAuth()
		if httpAuth != nil {
			headers := httpAuth.HTTPHeaders(requestURL, connectionConfig.AppID, connectionConfig.Secret)
			if len(headers) > 0 {
				req.Header = headers
			}
			host := req.Header.Get("Host")
			if len(host) > 0 {
				req.Host = host
			}
		}

		res, err := client.Do(req)
		if res != nil {
			defer res.Body.Close()
		}

		if res == nil || err != nil {
			log.Errorf("Connect Apollo Server Fail,url:%s,Error:%s", requestURL, err)
			// if error then sleep
			time.Sleep(onErrorRetryInterval)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Errorf("Connect Apollo Server Fail,url : %s ,Error: %s ", requestURL, err)
				// if error then sleep
				time.Sleep(onErrorRetryInterval)
				continue
			}

			if callBack != nil && callBack.SuccessCallBack != nil {
				return callBack.SuccessCallBack(responseBody, *callBack)
			}
			return nil, nil
		case http.StatusNotModified:
			log.Debug("Config Not Modified:", err)
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			}
			return nil, nil
		default:
			log.Errorf("Connect Apollo Server Fail,url:%s,StatusCode:%d", requestURL, res.StatusCode)
			// if error then sleep
			time.Sleep(onErrorRetryInterval)
			continue
		}
	}

	log.Error("Over Max Retry Still Error,Error:", err)
	if retry > retries {
		err = errors.New("over Max Retry Still Error")
	}
	return nil, err
}

//RequestRecovery 可以恢复的请求
func RequestRecovery(appConfig config.AppConfig,
	connectConfig *env.ConnectConfig,
	callBack *CallBack) (interface{}, error) {
	format := "%s%s"
	var err error
	var response interface{}

	for {
		host := loadBalance(appConfig)
		if host == "" {
			return nil, err
		}

		requestURL := fmt.Sprintf(format, host, connectConfig.URI)
		response, err = Request(requestURL, connectConfig, callBack)
		if err == nil {
			return response, nil
		}

		if host == appConfig.GetHost() {
			return response, err
		}

		server.SetDownNode(host, appConfig.GetHost())
	}
}

func loadBalance(appConfig config.AppConfig) string {
	if !server.IsConnectDirectly(appConfig.GetHost()) {
		return appConfig.GetHost()
	}
	serverInfo := extension.GetLoadBalance().Load(server.GetServers(appConfig.GetHost()))
	if serverInfo == nil {
		return utils.Empty
	}

	return serverInfo.HomepageURL
}
