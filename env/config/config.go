package config

import (
	"strings"
	"time"
)

//File 读写配置文件
type File interface {
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)

	Write(content interface{}, configPath string) error
}

//AppConfig 配置文件
type AppConfig struct {
	AppID            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	IP               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	IsBackupConfig   bool   `default:"true" json:"isBackupConfig"`
	BackupConfigPath string `json:"backupConfigPath"`
	Secret           string `json:"secret"`
}

//ServerInfo 服务器信息
type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceID  string `json:"instanceId"`
	HomepageURL string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

//GetIsBackupConfig whether backup config after fetch config from apollo
//false : no
//true : yes (default)
func (a *AppConfig) GetIsBackupConfig() bool {
	return a.IsBackupConfig
}

//GetBackupConfigPath GetBackupConfigPath
func (a *AppConfig) GetBackupConfigPath() string {
	return a.BackupConfigPath
}

//GetHost GetHost
func (a *AppConfig) GetHost() string {
	if strings.HasPrefix(a.IP, "http") {
		if !strings.HasSuffix(a.IP, "/") {
			return a.IP + "/"
		}
		return a.IP
	}
	return "http://" + a.IP + "/"
}

//SetNextTryConnTime if this connect is fail will set this time
func (a *AppConfig) SetNextTryConnTime(nextTryConnectPeriod int64) {
	a.NextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
}

//IsConnectDirectly is connect by ip directly
//false : no
//true : yes
func (a *AppConfig) IsConnectDirectly() bool {
	if a.NextTryConnTime >= 0 && a.NextTryConnTime > time.Now().Unix() {
		return true
	}

	return false
}
