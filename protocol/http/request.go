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
	"time"

	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"
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
)

//CallBack 请求回调函数
type CallBack struct {
	SuccessCallBack   func([]byte) (interface{}, error)
	NotModifyCallBack func() error
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
	tp := &http.Transport{
		MaxIdleConns:        defaultMaxConnsPerHost,
		MaxIdleConnsPerHost: defaultMaxConnsPerHost,
		DialContext: (&net.Dialer{
			KeepAlive: defaultKeepAliveSecond,
			Timeout:   defaultTimeoutBySecond,
		}).DialContext,
	}
	var err error
	url, err := url2.Parse(requestURL)
	if err != nil {
		log.Error("request Apollo Server url:%s, is invalid %s", requestURL, err)
		return nil, err
	}
	if strings.HasPrefix(url.Scheme, "https") {
		tp.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	client.Transport = tp
	retry := 0
	var responseBody []byte
	var res *http.Response
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
			log.Error("Generate connect Apollo request Fail,url:%s,Error:%s", requestURL, err)
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
		}

		res, err = client.Do(req)

		if res == nil || err != nil {
			log.Error("Connect Apollo Server Fail,url:%s,Error:%s", requestURL, err)
			// if error then sleep
			time.Sleep(onErrorRetryInterval)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err != nil {
				log.Error("Connect Apollo Server Fail,url:%s,Error:", requestURL, err)
				// if error then sleep
				time.Sleep(onErrorRetryInterval)
				continue
			}

			if callBack != nil && callBack.SuccessCallBack != nil {
				return callBack.SuccessCallBack(responseBody)
			}
			return nil, nil
		case http.StatusNotModified:
			log.Debug("Config Not Modified:", err)
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			}
			return nil, nil
		default:
			log.Error("Connect Apollo Server Fail,url:%s,StatusCode:%s", requestURL, res.StatusCode)
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
func RequestRecovery(appConfig *config.AppConfig,
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
			return response, err
		}

		env.SetDownNode(host)
	}
}

func loadBalance(appConfig *config.AppConfig) string {
	if !appConfig.IsConnectDirectly() {
		return appConfig.GetHost()
	}
	serverInfo := extension.GetLoadBalance().Load(env.GetServers())
	if serverInfo == nil {
		return utils.Empty
	}

	return serverInfo.HomepageURL
}
