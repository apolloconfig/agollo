package json

import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"
)

var vp = viper.New()

func init() {
	extension.AddFormatParser(constant.JSON, &Parser{})
	vp.SetConfigType("json")
}

// Parser properties转换器
type Parser struct {
}

// Parse 内存内容=>yml文件转换器
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	content, ok := configContent.(string)
	if !ok {
		return nil, nil
	}
	if utils.Empty == content{
		return nil, nil
	}

	buffer := bytes.NewBufferString(content)
	// 使用viper解析
	err := vp.ReadConfig(buffer)
	if err != nil {
		return nil, err
	}

	return convertToMap(vp), nil
}

func convertToMap(vp *viper.Viper) map[string]interface{}{
	if vp == nil {
		return nil
	}

	m := make(map[string]interface{})
	for _, key := range vp.AllKeys() {
		m[key] = vp.Get(key)
	}
	return m
}
