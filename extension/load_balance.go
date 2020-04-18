package extension

import "github.com/zouyx/agollo/v3/cluster"

var defaultLoadBalance cluster.LoadBalance

//SetLoadBalance 设置负载均衡器
func SetLoadBalance(loadBalance cluster.LoadBalance) {
	defaultLoadBalance = loadBalance
}

//GetLoadBalance 获取负载均衡器
func GetLoadBalance() cluster.LoadBalance {
	return defaultLoadBalance
}
