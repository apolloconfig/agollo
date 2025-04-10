// Copyright 2025 Apollo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/server"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/utils"
)

var (
	// onErrorRetryInterval defines the waiting period between retry attempts
	// when a request fails (2 seconds)
	onErrorRetryInterval = 2 * time.Second

	// connectTimeout defines the default connection timeout (1 second)
	connectTimeout = 1 * time.Second

	// maxRetries defines the maximum number of retry attempts for Apollo server connections
	maxRetries = 5

	// defaultMaxConnsPerHost defines the maximum number of concurrent connections per host
	defaultMaxConnsPerHost = 512

	// defaultTimeoutBySecond defines the default timeout for HTTP connections
	defaultTimeoutBySecond = 1 * time.Second

	// defaultKeepAliveSecond defines the duration to keep connections alive
	defaultKeepAliveSecond = 60 * time.Second

	// once ensures thread-safe initialization of the HTTP transport
	once sync.Once

	// defaultTransport is the shared HTTP transport configuration
	defaultTransport *http.Transport
)

// getDefaultTransport returns a configured HTTP transport with connection pooling
// Parameters:
//   - insecureSkipVerify: Whether to skip SSL certificate verification
//
// Returns:
//   - *http.Transport: Configured transport instance
func getDefaultTransport(insecureSkipVerify bool) *http.Transport {
	once.Do(func() {
		defaultTransport = &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
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
	return defaultTransport
}

// CallBack defines the callback functions for handling HTTP responses
type CallBack struct {
	// SuccessCallBack handles successful responses (HTTP 200)
	SuccessCallBack func([]byte, CallBack) (interface{}, error)
	// NotModifyCallBack handles not modified responses (HTTP 304)
	NotModifyCallBack func() error
	// AppConfigFunc provides application configuration
	AppConfigFunc func() config.AppConfig
	// Namespace identifies the configuration namespace
	Namespace string
}

// Request performs an HTTP request to the Apollo server with retry mechanism
// Parameters:
//   - requestURL: Target URL for the request
//   - connectionConfig: Connection configuration including timeout and credentials
//   - callBack: Callback functions for handling different response scenarios
//
// Returns:
//   - interface{}: Response data processed by callback functions
//   - error: Any error that occurred during the request
func Request(requestURL string, connectionConfig *env.ConnectConfig, callBack *CallBack) (interface{}, error) {
	client := &http.Client{}
	// Use custom timeout if set
	if connectionConfig != nil && connectionConfig.Timeout != 0 {
		client.Timeout = connectionConfig.Timeout
	} else {
		client.Timeout = connectTimeout
	}
	var err error
	url, err := url2.Parse(requestURL)
	if err != nil {
		log.Errorf("request Apollo Server url: %q is invalid: %v", requestURL, err)
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
		var req *http.Request
		req, err = http.NewRequest("GET", requestURL, nil)
		if req == nil || err != nil {
			log.Errorf("Generate connect Apollo request Fail, url:%s, error:%v", requestURL, err)
			// if error then sleep
			return nil, errors.New("generate connect Apollo request fail")
		}

		// Add header options
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

		var res *http.Response
		res, err = client.Do(req)
		if res != nil {
			defer res.Body.Close()
		}

		if res == nil || err != nil {
			log.Errorf("Connect Apollo Server Fail, url:%s, error:%v", requestURL, err)
			// if error then sleep
			time.Sleep(onErrorRetryInterval)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			var responseBody []byte
			responseBody, err = ioutil.ReadAll(res.Body)
			if err != nil {
				log.Errorf("Connect Apollo Server Fail, url: %s , error: %v", requestURL, err)
				// if error then sleep
				time.Sleep(onErrorRetryInterval)
				continue
			}

			if callBack != nil && callBack.SuccessCallBack != nil {
				return callBack.SuccessCallBack(responseBody, *callBack)
			}
			return nil, nil
		case http.StatusNotModified:
			log.Debugf("Config Not Modified, error: %v", err)
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			}
			return nil, nil
		case http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusMethodNotAllowed:
			log.Errorf("Connect Apollo Server Fail, url:%s, StatusCode:%d", requestURL, res.StatusCode)
			return nil, errors.New(fmt.Sprintf("Connect Apollo Server Fail, StatusCode:%d", res.StatusCode))
		default:
			log.Errorf("Connect Apollo Server Fail, url:%s, StatusCode:%d", requestURL, res.StatusCode)
			// if error then sleep
			time.Sleep(onErrorRetryInterval)
			continue
		}
	}

	log.Errorf("Over Max Retry Still Error, error: %v", err)
	if retry > retries {
		err = errors.New("over Max Retry Still Error")
	}
	return nil, err
}

// RequestRecovery performs requests with automatic server failover
// Parameters:
//   - appConfig: Application configuration containing server information
//   - connectConfig: Connection configuration for the request
//   - callBack: Callback functions for handling responses
//
// Returns:
//   - interface{}: Processed response data
//   - error: Any error that occurred during the request
//
// This function implements a failover mechanism by:
// 1. Using load balancing to select a server
// 2. Attempting the request
// 3. Marking failed servers as down
// 4. Retrying with different servers until successful
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

		server.SetDownNode(appConfig.GetHost(), host)
	}
}

// loadBalance selects an Apollo server using the configured load balancing strategy
// Parameters:
//   - appConfig: Application configuration containing server information
//
// Returns:
//   - string: Selected server's homepage URL, empty if no server available
//
// This function:
// 1. Checks if direct connection is required
// 2. Uses load balancer to select a server if needed
// 3. Returns the appropriate server URL
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
