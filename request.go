package agollo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cihub/seelog"
)

func request(url string, successCallBack func([]byte) (interface{}, error)) (interface{}, error) {
	client := &http.Client{
		Timeout: connect_timeout,
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

		res, err = client.Get(url)

		if res == nil || err != nil {
			seelog.Error("Connect Apollo Server Fail,Error:", err)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err != nil {
				seelog.Error("Connect Apollo Server Fail,Error:", err)
				continue
			}
			return successCallBack(responseBody)

		case http.StatusNotModified:
			seelog.Warn("Config Not Modified:", err)
			// update data for expire
			touchApolloConfigCache()
			return nil, nil

		default:
			seelog.Error("Connect Apollo Server Fail,Error:", err)
			if res != nil {
				seelog.Error("Connect Apollo Server Fail,StatusCode:", res.StatusCode)
			}
			err = errors.New("Connect Apollo Server Fail!")
			// if error then sleep
			time.Sleep(on_error_retry_interval)
			continue
		}
	}

	seelog.Error("Over Max Retry Still Error,Error:", err)
	if err != nil {
		err = errors.New("Over Max Retry Still Error!")
	}
	return nil, err
}

func requestRecovery(appConfig *AppConfig,
	urlSuffix string,
	successCallBack func([]byte) (interface{}, error)) (interface{}, error) {
	format := "%s%s"
	var err error
	var response interface{}

	for {
		host := appConfig.selectHost()
		if host == "" {
			return nil, err
		}

		requestUrl := fmt.Sprintf(format, host, urlSuffix)
		response, err = request(requestUrl, successCallBack)
		if err == nil {
			return response, err
		}

		seelog.Error("setDownNode", host, requestUrl)
		setDownNode(host)
	}

	return nil, errors.New("Try all Nodes Still Error!")
}
