package extension

import (
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/utils/parse"
)

var formatParser = make(map[constant.ConfigFileFormat]parse.ContentParser, 0)

// AddFormatParser 设置 formatParser
func AddFormatParser(key constant.ConfigFileFormat, contentParser parse.ContentParser) {
	formatParser[key] = contentParser
}

// GetFormatParser 获取 formatParser
func GetFormatParser(key constant.ConfigFileFormat) parse.ContentParser {
	return formatParser[key]
}
