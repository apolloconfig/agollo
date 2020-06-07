package parse

//ContentParser 内容转换
type ContentParser interface {
	Parse(configContent string) (map[string]string, error)
}
