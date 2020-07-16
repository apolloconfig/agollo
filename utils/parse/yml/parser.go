package yml

import (
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"
	"github.com/zouyx/agollo/v3/utils/parse/conver"
)

func init() {
	extension.AddFormatParser(constant.YML, newYMLParser())
}

func newYMLParser() *Parser {
	parser := &Parser{}
	parser.vp = conver.NewConverter("yml")
	return parser
}

// Parser properties转换器
type Parser struct {
	vp *conver.Converter
}

// Parse 内存内容=>yml文件转换器
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	content, ok := configContent.(string)
	if !ok {
		return nil, nil
	}
	if utils.Empty == content {
		return nil, nil
	}
	return d.vp.ConvertToMap(content)
}
