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
	"github.com/apolloconfig/agollo/v4/agcache"
	"github.com/apolloconfig/agollo/v4/cluster"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/file"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/apolloconfig/agollo/v4/protocol/auth"
)

// SetSignature 设置自定义 http 授权控件
func SetSignature(auth auth.HTTPAuth) {
	if auth != nil {
		extension.SetHTTPAuth(auth)
	}
}

/**
 * SetBackupFileHandler 设置自定义备份文件处理组件。
 *
 * 此函数允许用户设置自定义的备份文件处理器，并指定其优先级。
 * 优先级越高的处理器，将会优先被读取。
 * 若优先级相同，会根据添加顺序决定读取顺序，使用链表实现，具有稳定性。
 *
 * 默认的文件缓存实现的优先级为 10，具有较好的可靠性，推荐优先使用。
 * 用户可以根据自己的需求设置不同的优先级来决定读取顺序。
 * 推荐将 ConfigMap 实现的优先级设置低于文件缓存的实现。
 *
 * 参数：
 * - file: 自定义的文件处理器。如果为 nil，则不会添加处理器。
 * - priority: 文件处理器的优先级，数值越大优先级越高。
 *
 * 示例：
 *  extension.SetBackupFileHandler(myFileHandler, 11)
 *  extension.SetBackupFileHandler(configMapHandler, 9)
 */
func SetBackupFileHandler(file file.FileHandler, priority int) {
	if file != nil {
		extension.AddFileHandler(file, priority)
	}
}

// SetConfigMapHandler 设置自定义configMap持久化组件
func SetConfigMapHandler(configMap file.FileHandler, priority int) {
	if configMap != nil {
		extension.AddFileHandler(configMap, priority)
	}
}

// SetLoadBalance 设置自定义负载均衡组件
func SetLoadBalance(loadBalance cluster.LoadBalance) {
	if loadBalance != nil {
		extension.SetLoadBalance(loadBalance)
	}
}

// SetLogger 设置自定义logger组件
func SetLogger(loggerInterface log.LoggerInterface) {
	if loggerInterface != nil {
		log.InitLogger(loggerInterface)
	}
}

// SetCache 设置自定义cache组件
func SetCache(cacheFactory agcache.CacheFactory) {
	if cacheFactory != nil {
		extension.SetCacheFactory(cacheFactory)
	}
}
