package http

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/loadbalance"
	"github.com/zouyx/agollo/v3/utils"

	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/env"
)

var (
	//for on error retry
	onErrorRetryInterval = 2 * time.Second //2s

	connectTimeout = 1 * time.Second //1s

	//max retries connect apollo
	maxRetries = 5
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

	retry := 0
	var responseBody []byte
	var err error
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

		res, err = client.Get(requestURL)

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
			log.Info("Config Not Modified:", err)
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			}
			return nil, nil
		default:
			log.Error("Connect Apollo Server Fail,url:%s,StatusCode:%s", requestURL, res.StatusCode)
			err = errors.New("connect Apollo Server Fail")
			// if error then sleep
			time.Sleep(onErrorRetryInterval)
			continue
		}
	}

	log.Error("Over Max Retry Still Error,Error:", err)
	if err != nil {
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
	serverInfo := loadbalance.GetLoadBalance().Load(env.GetServers())
	if serverInfo == nil {
		return utils.Empty
	}

	return serverInfo.HomepageURL
}
