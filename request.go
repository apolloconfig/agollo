package agollo

import (
	"net/http"
	"io/ioutil"
	"time"
	"errors"
	"fmt"
)

type CallBack struct {
	SuccessCallBack func([]byte)(interface{},error)
	NotModifyCallBack func()error
}

type ConnectConfig struct {
	//设置到http.client中timeout字段
	Timeout time.Duration
	//连接接口的uri
	Uri string
}

func request(requestUrl string,connectionConfig *ConnectConfig,callBack *CallBack) (interface{},error){
	client := &http.Client{}
	//如有设置自定义超时时间即使用
	if connectionConfig!=nil&&connectionConfig.Timeout!=0{
		client.Timeout=connectionConfig.Timeout
	}else{
		client.Timeout=connect_timeout
	}

	retry:=0
	var responseBody []byte
	var err error
	var res *http.Response
	for{
		retry++

		if retry>max_retries{
			break
		}

		res,err=client.Get(requestUrl)

		if res==nil||err!=nil{
			logger.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err!=nil{
				logger.Error("Connect Apollo Server Fail,Error:",err)
				continue
			}

			if callBack!=nil&&callBack.SuccessCallBack!=nil {
				return callBack.SuccessCallBack(responseBody)
			}else{
				return nil,nil
			}
		case http.StatusNotModified:
			logger.Info("Config Not Modified:", err)
			if callBack!=nil&&callBack.NotModifyCallBack!=nil {
				return nil,callBack.NotModifyCallBack()
			}else{
				return nil,nil
			}
		default:
			logger.Error("Connect Apollo Server Fail,Error:",err)
			if res!=nil{
				logger.Error("Connect Apollo Server Fail,StatusCode:",res.StatusCode)
			}
			err=errors.New("Connect Apollo Server Fail!")
			// if error then sleep
			time.Sleep(on_error_retry_interval)
			continue
		}
	}

	logger.Error("Over Max Retry Still Error,Error:",err)
	if err!=nil{
		err=errors.New("Over Max Retry Still Error!")
	}
	return nil,err
}

func requestRecovery(appConfig *AppConfig,
	connectConfig *ConnectConfig,
	callBack *CallBack)(interface{},error) {
	format:="%s%s"
	var err error
	var response interface{}

	for {
		host:=appConfig.selectHost()
		if host==""{
			return nil,err
		}

		requestUrl:=fmt.Sprintf(format,host,connectConfig.Uri)
		response,err=request(requestUrl,connectConfig,callBack)
		if err==nil{
			return response,err
		}

		setDownNode(host)
	}

	return nil,errors.New("Try all Nodes Still Error!")
}
