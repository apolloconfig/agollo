package agcache

import (
	. "github.com/tevid/gohamcrest"
	"testing"
)

type TestCacheFactory struct {
}

func (d *TestCacheFactory) Create() CacheInterface {
	return &DefaultCache{}
}

func TestUseCacheFactory(t *testing.T) {
	UseCacheFactory(&TestCacheFactory{})

	factory := GetCacheFactory()
	cacheFactory := factory.(*TestCacheFactory)
	Assert(t, cacheFactory, NotNilVal())
}
