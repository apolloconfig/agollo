package cluster

import (
	"sync"

	"github.com/zouyx/agollo/v3/env/config"
)

//LoadBalance 负载均衡器
type LoadBalance interface {
	//Load 负载均衡，获取对应服务信息
	Load(servers *sync.Map) *config.ServerInfo
}
