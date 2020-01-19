package agcache

import (
	"testing"

	. "github.com/tevid/gohamcrest"
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
