package agcache

import (
	"errors"
	"sync"
)

//DefaultCache 默认缓存
type DefaultCache struct {
	defaultCache sync.Map
}

//Set 获取缓存
func (d *DefaultCache)Set(key string, value []byte, expireSeconds int) (err error)  {
	d.defaultCache.Store(key,value)
	return nil
}

//EntryCount 获取实体数量
func (d *DefaultCache)EntryCount() (entryCount int64){
	count:=int64(0)
	d.defaultCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

//Get 获取缓存
func (d *DefaultCache)Get(key string) (value []byte, err error){
	v, ok := d.defaultCache.Load(key)
	if !ok{
		return nil,errors.New("load default cache fail")
	}
	return v.([]byte),nil
}

//Range 遍历缓存
func (d *DefaultCache)Range(f func(key, value interface{}) bool){
	d.defaultCache.Range(f)
}

//Del 删除缓存
func (d *DefaultCache)Del(key string) (affected bool) {
	d.defaultCache.Delete(key)
	return true
}

//Clear 清除所有缓存
func (d *DefaultCache)Clear() {
	d.defaultCache=sync.Map{}
}

//DefaultCacheFactory 构造默认缓存组件工厂类
type DefaultCacheFactory struct {

}

//Create 创建默认缓存组件
func (d *DefaultCacheFactory) Create()CacheInterface {
	return &DefaultCache{}
}

