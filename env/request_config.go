package env

import (
	"time"
)

type ConnectConfig struct {
	//设置到http.client中timeout字段
	Timeout time.Duration
	//连接接口的uri
	Uri string
}
