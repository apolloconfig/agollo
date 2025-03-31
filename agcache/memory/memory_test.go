/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package memory

import (
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/agcache"
)

var testDefaultCache agcache.CacheInterface

func init() {
	factory := &DefaultCacheFactory{}
	testDefaultCache = factory.Create()

	_ = testDefaultCache.Set("a", "b", 100)
}

func TestDefaultCache_Set(t *testing.T) {
	err := testDefaultCache.Set("k", "c", 100)
	Assert(t, err, NilVal())
	Assert(t, int64(2), Equal(testDefaultCache.EntryCount()))
}

func TestDefaultCache_Range(t *testing.T) {
	var count int
	testDefaultCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	Assert(t, 2, Equal(count))
}

func TestDefaultCache_Del(t *testing.T) {
	b := testDefaultCache.Del("k")
	Assert(t, true, Equal(b))
	Assert(t, int64(1), Equal(testDefaultCache.EntryCount()))
}

func TestDefaultCache_Get(t *testing.T) {
	value, err := testDefaultCache.Get("a")
	Assert(t, err, NilVal())
	Assert(t, value, Equal("b"))
}

func TestDefaultCache_Clear(t *testing.T) {
	testDefaultCache.Clear()
	Assert(t, int64(0), Equal(testDefaultCache.EntryCount()))
}
