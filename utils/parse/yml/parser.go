package yml

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"github.com/zouyx/agollo/v3/agcache"
	"github.com/zouyx/agollo/v3/constant"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/utils"
)

const (
	defaultContentKey = "content"
	propertiesFormat  = "%s=%s\n"
)

var vp = viper.New()

func init() {
	extension.AddFormatParser(constant.YML, &Parser{})
	vp.SetConfigType(string(constant.YML))
}

// Parser properties转换器
type Parser struct {
}

// Parse 内存内容=>properties文件转换器
func (d *Parser) Parse(cache agcache.CacheInterface) (string, error) {
	if cache == nil {
		return utils.Empty, nil
	}

	value, err := cache.Get(defaultContentKey)
	if err != nil {
		return utils.Empty, err
	}
	buffer := bytes.NewBuffer(value)
	// 使用viper解析
	err = vp.ReadConfig(buffer)
	if err != nil {
		return utils.Empty, err
	}

	return convertToProperties(vp), nil
}

func convertToProperties(vp *viper.Viper) string {
	properties := utils.Empty
	if vp == nil {
		return properties
	}
	for _, key := range vp.AllKeys() {
		properties += fmt.Sprintf(propertiesFormat, key, vp.GetString(key))
	}
	return properties
}
