package yaml

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestParser_ParseCamelCase 测试解析包含小驼峰法key的YAML
func TestParser_ParseCamelCase(t *testing.T) {
	// 初始化解析器
	parser := &Parser{}

	// 定义一个包含小驼峰法key的YAML字符串
	yamlContent := `
appConfig:
  appName: TestApp
  appVersion: 1.0
serverConfig:
  enableSsl: true
  maxConnections: 100
loggerConfig:
  logLevel: debug
  logOutput: /var/log/test.log
`

	// 调用解析方法
	result, err := parser.Parse(yamlContent)

	// 确保解析没有错误
	assert.NoError(t, err, "解析YAML时出错")

	// 验证 appConfig 的解析结果
	appConfig, ok := result["appConfig"].(map[interface{}]interface{})
	assert.True(t, ok, "appConfig 应该是一个 map")
	assert.Equal(t, "TestApp", appConfig["appName"], "appName 不匹配")
	assert.Equal(t, 1.0, appConfig["appVersion"], "appVersion 不匹配")

	// 验证 serverConfig 的解析结果
	serverConfig, ok := result["serverConfig"].(map[interface{}]interface{})
	assert.True(t, ok, "serverConfig 应该是一个 map")
	assert.Equal(t, true, serverConfig["enableSsl"], "enableSsl 不匹配")
	assert.Equal(t, 100, serverConfig["maxConnections"], "maxConnections 不匹配")

	// 验证 loggerConfig 的解析结果
	loggerConfig, ok := result["loggerConfig"].(map[interface{}]interface{})
	assert.True(t, ok, "loggerConfig 应该是一个 map")
	assert.Equal(t, "debug", loggerConfig["logLevel"], "logLevel 不匹配")
	assert.Equal(t, "/var/log/test.log", loggerConfig["logOutput"], "logOutput 不匹配")
}
