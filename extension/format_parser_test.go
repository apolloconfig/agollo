package extension

import (
	"github.com/zouyx/agollo/v3/constant"
	"testing"

	. "github.com/tevid/gohamcrest"
)

// TestParser 默认内容转换器
type TestParser struct {
}

// Parse 内存内容默认转换器
func (d *TestParser) Parse(s interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func TestAddFormatParser(t *testing.T) {
	AddFormatParser(constant.DEFAULT, &TestParser{})
	AddFormatParser(constant.Properties, &TestParser{})

	p := GetFormatParser(constant.DEFAULT)

	b := p.(*TestParser)
	Assert(t, b, NotNilVal())

	Assert(t, len(formatParser), Equal(2))
}
