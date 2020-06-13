package properties

import (
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
)

func init() {
	extension.AddFormatParser(constant.Properties, &Parser{})
}

// Parser properties转换器
type Parser struct {
}

// Parse 内存内容=>properties文件转换器
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return nil, nil
}
