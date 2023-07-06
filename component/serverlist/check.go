package serverlist

import (
	"fmt"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/perror"
	"github.com/apolloconfig/agollo/v4/protocol/http"
	"strconv"
	"time"
)

// CheckSecretOK 检查秘钥是否正确
func CheckSecretOK(appConfigFunc func() config.AppConfig) (err error) {
	if appConfigFunc == nil {
		return fmt.Errorf("没有找到Apollo配置，请确认！")
	}

	appConfig := appConfigFunc()
	c := &env.ConnectConfig{
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	if appConfigFunc().SyncServerTimeout > 0 {
		duration, err := time.ParseDuration(strconv.Itoa(appConfigFunc().SyncServerTimeout) + "s")
		if err != nil {
			return err
		}
		c.Timeout = duration
	}
	if _, err = http.Request(appConfig.GetCheckSecretURL(), c, nil); err != nil {
		switch err {
		case perror.ErrOverMaxRetryStill:
			return fmt.Errorf("检查Apollo秘钥正确性失败. err: %v", err)
		case perror.ErrUnauthorized:
			return fmt.Errorf("Apollo-Secret不正确. AppID: %s, Cluster: %s", appConfig.AppID, appConfig.Cluster)
		default:
			return
		}
	}
	return
}
