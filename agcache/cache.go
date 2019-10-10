package agcache

import "github.com/coocood/freecache"

//CacheInterface 自定义缓存组件接口
type CacheInterface interface {
	Set(key, value []byte, expireSeconds int) (err error)

	EntryCount() (entryCount int64)

	Get(key []byte) (value []byte, err error)

	Del(key []byte) (affected bool)

	NewIterator() *freecache.Iterator

	TTL(key []byte) (timeLeft uint32, err error)

	Clear()
}

const (
	//50m
	apolloConfigCacheSize = 50 * 1024 * 1024

	//1 minute
	configCacheExpireTime = 120
)

//CacheFactory 缓存组件工厂接口
type CacheFactory interface {
	//Create 创建缓存组件
	Create() CacheInterface
}

//DefaultCacheFactory 构造默认缓存组件工厂类
type DefaultCacheFactory struct {

}

func (d *DefaultCacheFactory) Create()CacheInterface  {
	return freecache.NewCache(apolloConfigCacheSize)
}