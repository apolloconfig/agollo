package utils

import (
	. "github.com/tevid/gohamcrest"
	"github.com/zouyx/agollo/v2/agcache"
	"testing"
)

var (
	testDefaultCache agcache.CacheInterface
	defaultParser    ContentParser
	propertiesParser ContentParser
)

func init() {
	factory := &agcache.DefaultCacheFactory{}
	testDefaultCache = factory.Create()

	defaultParser = &DefaultParser{}

	propertiesParser = &PropertiesParser{}

	testDefaultCache.Set("a", []byte("b"), 100)
	testDefaultCache.Set("c", []byte("d"), 100)
	testDefaultCache.Set("content", []byte("content"), 100)
}

func TestDefaultParser(t *testing.T) {
	s, err := defaultParser.Parse(testDefaultCache)
	Assert(t, err, NilVal())
	Assert(t, s, Equal("content"))
}

func TestPropertiesParser(t *testing.T) {
	s, err := propertiesParser.Parse(testDefaultCache)
	Assert(t, err, NilVal())
	Assert(t, s, Equal(`a=b
c=d
content=content
`))
}
