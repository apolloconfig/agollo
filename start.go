package agollo

import (
	"github.com/zouyx/agollo/v3/agcache"
	_ "github.com/zouyx/agollo/v3/cluster/roundrobin"
	"github.com/zouyx/agollo/v3/component"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/component/notify"
	"github.com/zouyx/agollo/v3/component/serverlist"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/env/file"
	_ "github.com/zouyx/agollo/v3/env/file/json"
	"github.com/zouyx/agollo/v3/extension"
	"github.com/zouyx/agollo/v3/storage"
)

var (
	initAppConfigFunc func() (*config.AppConfig, error)
)

//InitCustomConfig init config by custom
func InitCustomConfig(loadAppConfig func() (*config.AppConfig, error)) {
	initAppConfigFunc = loadAppConfig
}

//start apollo
func Start() error {
	return startAgollo()
}

//SetBackupFileHandler 设置自定义备份文件处理组件
func SetBackupFileHandler(file file.FileHandler) {
	if file != nil {
		extension.SetFileHandler(file)
	}
}

//SetLogger 设置自定义logger组件
func SetLogger(loggerInterface log.LoggerInterface) {
	if loggerInterface != nil {
		log.InitLogger(loggerInterface)
	}
}

//SetCache 设置自定义cache组件
func SetCache(cacheFactory agcache.CacheFactory) {
	if cacheFactory != nil {
		agcache.UseCacheFactory(cacheFactory)
		storage.InitConfigCache()
	}
}

func startAgollo() error {
	// 有了配置之后才能进行初始化
	if err := env.InitConfig(initAppConfigFunc); err != nil {
		return err
	}

	notify.InitAllNotifications(nil)
	serverlist.InitSyncServerIPList()

	//first sync
	if err := notify.SyncConfigs(); err != nil {
		return err
	}
	log.Debug("init notifySyncConfigServices finished")

	//start long poll sync config
	go component.StartRefreshConfig(&notify.ConfigComponent{})

	log.Info("agollo start finished ! ")

	return nil
}
