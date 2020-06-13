package yml

import (
	"bytes"
	"encoding/json"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"
)

var vp = viper.New()

func init() {
	parser := &Parser{}
	extension.AddFormatParser(constant.YML, parser)
	extension.AddFormatParser(constant.YAML, parser)
	vp.SetConfigType("yml")
}

// Parser properties转换器
type Parser struct {
}

// Parse 内存内容=>yml文件转换器
func (d *Parser) Parse(configContent string) (map[string]string, error) {
	if utils.Empty == configContent {
		return nil, nil
	}

	buffer := bytes.NewBufferString(configContent)
	// 使用viper解析
	err := vp.ReadConfig(buffer)
	if err != nil {
		return nil, err
	}

	return convertToMap(vp), nil
}

func convertToMap(vp *viper.Viper) map[string]string {
	if vp == nil {
		return nil
	}

	m := make(map[string]string)
	for _, key := range vp.AllKeys() {
		v := vp.Get(key)
		s, err := cast.ToStringE(v)
		if err != nil {
			b, _ := json.Marshal(v)
			s = string(b)
		}
		m[key] = s
	}
	return m
}
