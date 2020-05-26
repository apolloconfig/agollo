package extension

import "github.com/zouyx/agollo/v3/agcache"

var (
	gobalCacheFactory agcache.CacheFactory
)

//GetCacheFactory 获取CacheFactory
func GetCacheFactory() agcache.CacheFactory {
	return gobalCacheFactory
}

//SetCacheFactory 替换CacheFactory
func SetCacheFactory(cacheFactory agcache.CacheFactory) {
	gobalCacheFactory = cacheFactory
}
