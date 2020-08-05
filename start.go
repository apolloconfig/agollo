/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package agollo

import (
	"container/list"
	"github.com/zouyx/agollo/v3/agcache"
	_ "github.com/zouyx/agollo/v3/agcache/memory"
	"github.com/zouyx/agollo/v3/cluster"
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
	"github.com/zouyx/agollo/v3/protocol/auth"
	_ "github.com/zouyx/agollo/v3/protocol/auth/sign"
	"github.com/zouyx/agollo/v3/storage"
	_ "github.com/zouyx/agollo/v3/utils/parse/normal"
	_ "github.com/zouyx/agollo/v3/utils/parse/properties"
	_ "github.com/zouyx/agollo/v3/utils/parse/yml"
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

//SetSignature 设置自定义 http 授权控件
func SetSignature(auth auth.HTTPAuth) {
	if auth != nil {
		extension.SetHTTPAuth(auth)
	}
}

//SetBackupFileHandler 设置自定义备份文件处理组件
func SetBackupFileHandler(file file.FileHandler) {
	if file != nil {
		extension.SetFileHandler(file)
	}
}

//SetLoadBalance 设置自定义负载均衡组件
func SetLoadBalance(loadBalance cluster.LoadBalance) {
	if loadBalance != nil {
		extension.SetLoadBalance(loadBalance)
	}
}

//SetLogger 设置自定义logger组件
func SetLogger(loggerInterface log.LoggerInterface) {
	if loggerInterface != nil {
		log.InitLogger(loggerInterface)
	}
}

//UseEventDispatch  添加为某些key分发event功能
func UseEventDispatch() {
	storage.UseEventDispatch()
}

//SetCache 设置自定义cache组件
func SetCache(cacheFactory agcache.CacheFactory) {
	if cacheFactory != nil {
		extension.SetCacheFactory(cacheFactory)
		storage.InitConfigCache()
	}
}

//AddChangeListener 增加变更监控
func AddChangeListener(listener storage.ChangeListener) {
	storage.AddChangeListener(listener)
}

//RemoveChangeListener 增加变更监控
func RemoveChangeListener(listener storage.ChangeListener) {
	storage.AddChangeListener(listener)
}

//GetChangeListeners 获取配置修改监听器列表
func GetChangeListeners() *list.List {
	return storage.GetChangeListeners()
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
