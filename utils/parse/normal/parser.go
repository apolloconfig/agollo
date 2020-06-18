package normal

import (
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
)

func init() {
	extension.AddFormatParser(constant.DEFAULT, &Parser{})
}

// Parser 默认内容转换器
type Parser struct {
}

// Parse 内存内容默认转换器
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return nil, nil
}
