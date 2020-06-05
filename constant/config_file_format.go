package constant

//ConfigFileFormat 配置文件类型
type ConfigFileFormat string

const (
	//Properties Properties
	Properties ConfigFileFormat = "properties"
	//XML XML
	XML ConfigFileFormat = "xml"
	//JSON JSON
	JSON ConfigFileFormat = "json"
	//YML YML
	YML ConfigFileFormat = "yml"
	//YAML YAML
	YAML ConfigFileFormat = "yaml"
	// DEFAULT DEFAULT
	DEFAULT ConfigFileFormat = "default"
)
