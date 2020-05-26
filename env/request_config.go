package env

import (
	"time"
)

//ConnectConfig 网络请求配置
type ConnectConfig struct {
	//设置到http.client中timeout字段
	Timeout time.Duration
	//连接接口的uri
	URI string
	//是否重试
	IsRetry bool
	//appID
	AppId string
	//密钥
	Secret string
}
