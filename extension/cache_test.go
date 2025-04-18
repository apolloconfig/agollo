// Copyright 2025 Apollo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package extension

import (
	"errors"
	"sync"
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/agcache"
)

type TestCacheFactory struct {
}

func (d *TestCacheFactory) Create() agcache.CacheInterface {
	return &defaultCache{}
}

// DefaultCache 默认缓存
type defaultCache struct {
	defaultCache sync.Map
}

// Set 获取缓存
func (d *defaultCache) Set(key string, value interface{}, expireSeconds int) (err error) {
	d.defaultCache.Store(key, value)
	return nil
}

// EntryCount 获取实体数量
func (d *defaultCache) EntryCount() (entryCount int64) {
	count := int64(0)
	d.defaultCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Get 获取缓存
func (d *defaultCache) Get(key string) (value interface{}, err error) {
	v, ok := d.defaultCache.Load(key)
	if !ok {
		return nil, errors.New("load default cache fail")
	}
	return v.([]byte), nil
}

// Range 遍历缓存
func (d *defaultCache) Range(f func(key, value interface{}) bool) {
	d.defaultCache.Range(f)
}

// Del 删除缓存
func (d *defaultCache) Del(key string) (affected bool) {
	d.defaultCache.Delete(key)
	return true
}

// Clear 清除所有缓存
func (d *defaultCache) Clear() {
	d.defaultCache = sync.Map{}
}

// DefaultCacheFactory 构造默认缓存组件工厂类
type DefaultCacheFactory struct {
}

// Create 创建默认缓存组件
func (d *DefaultCacheFactory) Create() agcache.CacheInterface {
	return &defaultCache{}
}

func TestUseCacheFactory(t *testing.T) {
	SetCacheFactory(&TestCacheFactory{})

	factory := GetCacheFactory()
	cacheFactory := factory.(*TestCacheFactory)
	Assert(t, cacheFactory, NotNilVal())
}
