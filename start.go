package agollo

import (
	"github.com/zouyx/agollo/v3/agcache"
	"github.com/zouyx/agollo/v3/component"
	"github.com/zouyx/agollo/v3/component/log"
	"github.com/zouyx/agollo/v3/component/notify"
	_ "github.com/zouyx/agollo/v3/component/serverlist"
	"github.com/zouyx/agollo/v3/env"
	"github.com/zouyx/agollo/v3/env/config"
	"github.com/zouyx/agollo/v3/env/filehandler"
	_ "github.com/zouyx/agollo/v3/env/filehandler/defaultfile"
	_ "github.com/zouyx/agollo/v3/loadbalance/roundrobin"
	"github.com/zouyx/agollo/v3/storage"
)

func init() {
}

//InitCustomConfig init config by custom
func InitCustomConfig(loadAppConfig func() (*config.AppConfig, error)) {
	env.InitConfig(loadAppConfig)
}

//start apollo
func Start() error {
	return startAgollo()
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

//SetFile define backup file (write and read) handler
func SetFileHandler(handler filehandler.FileHandler) {
	if handler != nil {
		filehandler.SetFileHandler(handler)
	}
}

func startAgollo() error {
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
