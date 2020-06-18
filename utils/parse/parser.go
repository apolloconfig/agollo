package parse

//ContentParser 内容转换
type ContentParser interface {
	Parse(configContent interface{}) (map[string]interface{}, error)
}
