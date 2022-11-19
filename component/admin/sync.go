package admin

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	"github.com/qshuai/agollo/v4/component/log"
	"github.com/qshuai/agollo/v4/env"
	"github.com/qshuai/agollo/v4/env/config"
	"github.com/qshuai/agollo/v4/protocol/http"
)

const (
	syncInterval = time.Minute
)

var (
	// type: []string
	services atomic.Value
)

type SyncAdminServiceComponent struct {
	appConfig func() config.AppConfig
}

func (s *SyncAdminServiceComponent) Start() {
	srvs, err := SyncAdminService(s.appConfig)
	if err != nil {
		log.Errorf("sync admin service err: %s", err)
	} else if len(srvs) == 0 {
		log.Warn("admin service instance is empty")
	} else {
		services.Store(srvs)
	}

	ticker := time.NewTimer(syncInterval)
	for {
		select {
		case <-ticker.C:
			srvs, err = SyncAdminService(s.appConfig)
			if err != nil {
				log.Errorf("sync admin service err: %s", err)
			} else if len(srvs) == 0 {
				log.Warn("admin service instance is empty")
			} else {
				services.Store(srvs)
			}
		}

		ticker.Reset(syncInterval)
	}
}

func SyncAdminService(appConfigFn func() config.AppConfig) ([]string, error) {
	if appConfigFn == nil {
		return nil, nil
	}

	appConfig := appConfigFn()
	c := &env.ConnectConfig{
		AppID:  appConfig.AppID,
		Secret: appConfig.Secret,
	}
	if appConfig.SyncServerTimeout > 0 {
		c.Timeout = time.Duration(appConfig.SyncServerTimeout) * time.Second
	}

	srvList, err := http.Request(appConfig.GetAdminServiceURL(), c, &http.CallBack{
		SuccessCallBack:   syncAdminServiceSuccessCallBack,
		NotModifyCallBack: nil,
		AppConfigFunc:     nil,
		Namespace:         "",
	})
	if err != nil {
		return nil, err
	}
	if srvList == nil || len(srvList.([]string)) == 0 {
		return nil, nil
	}

	return srvList.([]string), nil
}

type adminServiceResp struct {
	AppName     string `json:"appName"`     // APOLLO-ADMINSERVICE
	HomepageURL string `json:"homepageUrl"` // http://10.10.1.101:8090/
	InstanceID  string `json:"instanceId"`  // apollo-adminservice-86d9jd464f-mh7j2:apollo-adminservice:8090
}

func syncAdminServiceSuccessCallBack(responseBody []byte, cb http.CallBack) (interface{}, error) {
	log.Debugf("admin service: %s", string(responseBody))

	services := make([]*adminServiceResp, 0)
	err := json.Unmarshal(responseBody, &services)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, nil
	}

	ret := make([]string, 0, len(services))
	for _, srv := range services {
		if strings.HasSuffix(srv.HomepageURL, "/") {
			ret = append(ret, srv.HomepageURL)
		} else {
			ret = append(ret, srv.HomepageURL+"/")
		}
	}

	return ret, nil
}

func New(appConfig func() config.AppConfig) *SyncAdminServiceComponent {
	return &SyncAdminServiceComponent{
		appConfig: appConfig,
	}
}
