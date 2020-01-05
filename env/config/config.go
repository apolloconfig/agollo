package config

import (
	"strings"
	"sync"
	"time"
)

type ConfigFile interface {
	Load(fileName string, unmarshal func([]byte) (interface{}, error)) (interface{}, error)

	Write(content interface{}, configPath string) error
}

type AppConfig struct {
	AppId            string `json:"appId"`
	Cluster          string `json:"cluster"`
	NamespaceName    string `json:"namespaceName"`
	Ip               string `json:"ip"`
	NextTryConnTime  int64  `json:"-"`
	IsBackupConfig   bool   `default:"true" json:"isBackupConfig"`
	BackupConfigPath string `json:"backupConfigPath"`
}

type ServerInfo struct {
	AppName     string `json:"appName"`
	InstanceId  string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`
	IsDown      bool   `json:"-"`
}

//getIsBackupConfig whether backup config after fetch config from apollo
//false : no
//true : yes (default)
func (this *AppConfig) GetIsBackupConfig() bool {
	return this.IsBackupConfig
}

func (this *AppConfig) GetBackupConfigPath() string {
	return this.BackupConfigPath
}

func (this *AppConfig) GetHost() string {
	if strings.HasPrefix(this.Ip, "http") {
		if !strings.HasSuffix(this.Ip, "/") {
			return this.Ip + "/"
		}
		return this.Ip
	}
	return "http://" + this.Ip + "/"
}

//if this connect is fail will set this time
func (this *AppConfig) SetNextTryConnTime(nextTryConnectPeriod int64) {
	this.NextTryConnTime = time.Now().Unix() + nextTryConnectPeriod
}

//is connect by ip directly
//false : no
//true : yes
func (this *AppConfig) isConnectDirectly() bool {
	if this.NextTryConnTime >= 0 && this.NextTryConnTime > time.Now().Unix() {
		return true
	}

	return false
}

func (this *AppConfig) SelectHost(servers *sync.Map) string {
	if !this.isConnectDirectly() {
		return this.GetHost()
	}

	host := ""

	servers.Range(func(k, v interface{}) bool {
		server := v.(*ServerInfo)
		// if some node has down then select next node
		if server.IsDown {
			return true
		}
		host = k.(string)
		return false
	})

	return host
}
