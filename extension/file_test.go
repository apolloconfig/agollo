package extension

import (
	"testing"

	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/file"

	. "github.com/tevid/gohamcrest"
)

type TestFileHandler struct {
}

//WriteConfigFile 写入配置文件
func (r *TestFileHandler) WriteConfigFile(config *env.ApolloConfig, configPath string) error {
	return nil
}

//GetConfigFile 获得配置文件路径
func (r *TestFileHandler) GetConfigFile(configDir string, namespace string) string {
	return ""
}

func (r *TestFileHandler) LoadConfigFile(configDir string, namespace string) (*env.ApolloConfig, error) {
	return nil, nil
}

func TestSetFileHandler(t *testing.T) {
	SetFileHandler(&TestFileHandler{})

	fileHandler := GetFileHandler()

	b := fileHandler.(file.FileHandler)
	Assert(t, b, NotNilVal())
}
