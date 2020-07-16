package conver
import (
	"bytes"
	"github.com/spf13/viper"
)

// Converter 格式转换器
type Converter struct {
	vp *viper.Viper
}

// NewConverter 根据传入的格式要求，生成相应格式转换器
func NewConverter(in string) *Converter {
	c := new(Converter)
	c.vp = viper.New()
	c.vp.SetConfigType(in)
	return c
}

// ConvertToMap 根据输入的格式字符串，转换成map类型
func (c *Converter) ConvertToMap(configContent string)(map[string]interface{}, error){
	buffer := bytes.NewBufferString(configContent)
	// 使用viper解析
	err := c.vp.ReadConfig(buffer)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for _, key := range c.vp.AllKeys() {
		m[key] = c.vp.Get(key)
	}
	return m, nil
}
