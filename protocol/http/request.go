package http

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/zouyx/agollo/v2/component/log"
	"github.com/zouyx/agollo/v2/env"
)

var (
	//for on error retry
	on_error_retry_interval = 1 * time.Second //1s

	connect_timeout = 1 * time.Second //1s

	//max retries connect apollo
	max_retries = 5
)

type CallBack struct {
	SuccessCallBack   func([]byte) (interface{}, error)
	NotModifyCallBack func() error
}

func request(requestUrl string, connectionConfig *env.ConnectConfig, callBack *CallBack) (interface{}, error) {
	client := &http.Client{}
	//如有设置自定义超时时间即使用
	if connectionConfig != nil && connectionConfig.Timeout != 0 {
		client.Timeout = connectionConfig.Timeout
	} else {
		client.Timeout = connect_timeout
	}

	retry := 0
	var responseBody []byte
	var err error
	var res *http.Response
	for {
		retry++

		if retry > max_retries {
			break
		}

		res, err = client.Get(requestUrl)

		if res == nil || err != nil {
			Logger.Error("Connect Apollo Server Fail,url:%s,Error:%s", requestUrl, err)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err != nil {
				Logger.Error("Connect Apollo Server Fail,url:%s,Error:", requestUrl, err)
				continue
			}

			if callBack != nil && callBack.SuccessCallBack != nil {
				return callBack.SuccessCallBack(responseBody)
			} else {
				return nil, nil
			}
		case http.StatusNotModified:
			Logger.Info("Config Not Modified:", err)
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			} else {
				return nil, nil
			}
		default:
			Logger.Error("Connect Apollo Server Fail,url:%s,Error:%s", requestUrl, err)
			if res != nil {
				Logger.Error("Connect Apollo Server Fail,url:%s,StatusCode:%s", requestUrl, res.StatusCode)
			}
			err = errors.New("Connect Apollo Server Fail!")
			// if error then sleep
			time.Sleep(on_error_retry_interval)
			continue
		}
	}

	Logger.Error("Over Max Retry Still Error,Error:", err)
	if err != nil {
		err = errors.New("Over Max Retry Still Error!")
	}
	return nil, err
}

func RequestRecovery(appConfig *env.AppConfig,
	connectConfig *env.ConnectConfig,
	callBack *CallBack) (interface{}, error) {
	format := "%s%s"
	var err error
	var response interface{}

	for {
		host := appConfig.SelectHost()
		if host == "" {
			return nil, err
		}

		requestUrl := fmt.Sprintf(format, host, connectConfig.Uri)
		response, err = request(requestUrl, connectConfig, callBack)
		if err == nil {
			return response, err
		}

		env.SetDownNode(host)
	}

	return nil, errors.New("Try all Nodes Still Error!")
}
