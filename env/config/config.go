package config

import (
	"strings"
	"time"
)

//ConfigFile 读写配置文件
type ConfigFile interface {
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)

	Write(content interface{}, configPath string) error
}

//AppConfig 配置文件
type AppConfig struct {
	AppId            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	Ip               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	IsBackupConfig   bool   `default:"true" json:"isBackupConfig"`
	BackupConfigPath string `json:"backupConfigPath"`
}

//ServerInfo 服务器信息
type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceId  string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

//getIsBackupConfig whether backup config after fetch config from apollo
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
	if strings.HasPrefix(a.Ip, "http") {
		if !strings.HasSuffix(a.Ip, "/") {
			return a.Ip + "/"
		}
		return a.Ip
	}
	return "http://" + a.Ip + "/"
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
