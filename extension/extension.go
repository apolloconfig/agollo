package extension

import (
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/file"
	"github.com/zouyx/agollo/v3/env/file/json"
)

//InitFileHandler 根据配置文件初始化filehandler
func InitFileHandler() {
	if env.GetPlainAppConfig().GetWithRawBackup() {
		file.SetFileHandler(&json.RawHandler{})
		return
	}

	file.SetFileHandler(&json.JSONFileHandler{})
}
