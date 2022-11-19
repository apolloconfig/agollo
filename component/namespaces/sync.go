package namespaces

import (
	"encoding/json"
	"time"

	"github.com/qshuai/agollo/v4/component/log"
	"github.com/qshuai/agollo/v4/env"
	"github.com/qshuai/agollo/v4/env/config"
	"github.com/qshuai/agollo/v4/protocol/http"
)

type SyncNamespaceListComponent struct {
	appConfig func() config.AppConfig
	callBack  func(ns string) error
}

func (n *SyncNamespaceListComponent) Start() {
	interval := n.appConfig().SyncNamespaceInterval
	if interval < time.Second {
		interval = time.Second
	}

	t := time.NewTimer(interval)
	var err error
	for {
		select {
		case <-t.C:
			err = SyncNamespaceList(n.appConfig, n.callBack)
			if err != nil {
				log.Errorf("sync namespace err: %s", err)
			}

			t.Reset(interval)
		}
	}
}

func SyncNamespaceList(appConfigFn func() config.AppConfig, callBack func(ns string) error) error {
	if appConfigFn == nil || callBack == nil {
		return nil
	}

	appConfig := appConfigFn()
	c := &env.ConnectConfig{
		AppID:         appConfig.AppID,
		Secret:        appConfig.Secret,
		Authorization: appConfig.AuthorizationToken, // 认证
	}
	if appConfig.SyncServerTimeout > 0 {
		c.Timeout = time.Duration(appConfig.SyncServerTimeout) * time.Second
	}

	url, err := appConfig.GetNamespaceListURL()
	if err != nil {
		return err
	}
	ns, err := http.Request(url, c, &http.CallBack{
		SuccessCallBack:   syncNamespaceListSuccessCallBack,
		NotModifyCallBack: nil,
		AppConfigFunc:     nil,
	})
	if err != nil {
		return err
	}
	if ns == nil || len(ns.([]string)) == 0 {
		return nil
	}

	for _, namespace := range ns.([]string) {
		err = callBack(namespace)
		if err != nil {
			log.Errorf("add namespace callback err: %s", err)
		}
	}

	return nil
}

type namespaceResp struct {
	ID                         int64  `json:"id"`                         // 263
	AppID                      string `json:"appId"`                      // appid
	ClusterName                string `json:"clusterName"`                // default
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`        // apollo
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`      // 2022-09-02T17:27:36.000+0800
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`   // apollo
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"` // 2022-09-02T17:27:36.000+0800
	NamespaceName              string `json:"namespaceName"`              // application
}

func syncNamespaceListSuccessCallBack(responseBody []byte, cb http.CallBack) (interface{}, error) {
	log.Debugf("get all namespace info: ", string(responseBody))

	nsList := make([]*namespaceResp, 0)
	err := json.Unmarshal(responseBody, &nsList)
	if err != nil {
		return nil, err
	}
	if len(nsList) == 0 {
		log.Warn("namespace is empty")
		return nil, nil
	}

	ns := make([]string, 0, len(nsList))
	for _, entry := range nsList {
		ns = append(ns, entry.NamespaceName)
	}

	return ns, nil
}

func New(appConfig func() config.AppConfig, callBack func(namespace string) error) *SyncNamespaceListComponent {
	return &SyncNamespaceListComponent{
		appConfig: appConfig,
		callBack:  callBack,
	}
}
