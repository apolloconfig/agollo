package load_balance

import (
	"github.com/zouyx/agollo/v2/env/config"
	"sync"
)

var defaultLoadBalance LoadBalance

//LoadBalance 负载均衡器
type LoadBalance interface {
	//Load 负载均衡，获取对应服务信息
	Load(servers *sync.Map) *config.ServerInfo
}

//SetLoadBalance 设置负载均衡器
func SetLoadBalance(loadBalance LoadBalance) {
	defaultLoadBalance = loadBalance
}

//GetLoadBalance 获取负载均衡器
func GetLoadBalance() LoadBalance {
	return defaultLoadBalance
}
