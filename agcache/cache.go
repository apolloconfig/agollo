package agcache

import "github.com/coocood/freecache"

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

type CacheFactory interface {
	Create() CacheInterface
}

type DefaultCacheFactory struct {

}

func (this *DefaultCacheFactory) Create()CacheInterface  {
	return freecache.NewCache(apolloConfigCacheSize)
}