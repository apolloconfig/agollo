package memory

import (
	"testing"

	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v3/agcache"
)

var testDefaultCache agcache.CacheInterface

func init() {
	factory := &DefaultCacheFactory{}
	testDefaultCache = factory.Create()

	testDefaultCache.Set("a", []byte("b"), 100)
}

func TestDefaultCache_Set(t *testing.T) {
	err := testDefaultCache.Set("k", []byte("c"), 100)
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
	Assert(t, string(value), Equal("b"))
}

func TestDefaultCache_Clear(t *testing.T) {
	testDefaultCache.Clear()
	Assert(t, int64(0), Equal(testDefaultCache.EntryCount()))
}
