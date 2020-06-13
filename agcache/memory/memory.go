package memory

import (
	"errors"
	"github.com/zouyx/agollo/v3/agcache"
	"github.com/zouyx/agollo/v3/extension"
	"sync"
	"sync/atomic"
)

func init() {
	extension.SetCacheFactory(&DefaultCacheFactory{})
}

//DefaultCache 默认缓存
type DefaultCache struct {
	defaultCache sync.Map
	count int64
}

//Set 获取缓存
func (d *DefaultCache) Set(key string, value interface{}, expireSeconds int) (err error) {
	d.defaultCache.Store(key, value)
	atomic.AddInt64(&d.count, int64(1))
	return nil
}

//EntryCount 获取实体数量
func (d *DefaultCache) EntryCount() (entryCount int64) {
	c := atomic.LoadInt64(&d.count)
	return c
}

//Get 获取缓存
func (d *DefaultCache) Get(key string) (value interface{}, err error) {
	v, ok := d.defaultCache.Load(key)
	if !ok {
		return nil, errors.New("load default cache fail")
	}
	return v, nil
}

//Range 遍历缓存
func (d *DefaultCache) Range(f func(key, value interface{}) bool) {
	d.defaultCache.Range(f)
}

//Del 删除缓存
func (d *DefaultCache) Del(key string) (affected bool) {
	d.defaultCache.Delete(key)
	atomic.AddInt64(&d.count, int64(-1))
	return true
}

//Clear 清除所有缓存
func (d *DefaultCache) Clear() {
	d.defaultCache = sync.Map{}
	atomic.StoreInt64(&d.count, int64(0))
}

//DefaultCacheFactory 构造默认缓存组件工厂类
type DefaultCacheFactory struct {
}

//Create 创建默认缓存组件
func (d *DefaultCacheFactory) Create() agcache.CacheInterface {
	return &DefaultCache{}
}
