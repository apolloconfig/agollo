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

package sync

import (
	"fmt"
	"github.com/zouyx/agollo/v4/component/remote"
	"github.com/zouyx/agollo/v4/component/remote/abs"
	"github.com/zouyx/agollo/v4/env/config"
	"github.com/zouyx/agollo/v4/protocol/http"
	"github.com/zouyx/agollo/v4/utils"
	"net/url"
)

var (
	remoteApollo apolloConfig
)

func init() {
	remoteApollo = apolloConfig{}
}

func GetInstance() remote.ApolloConfig {
	return &remoteApollo
}

type apolloConfig struct {
	abs.ApolloConfig
}

func (*apolloConfig) GetNotifyURLSuffix(notifications string, config config.AppConfig) string {
	return ""
}

func (*apolloConfig) GetSyncURI(config config.AppConfig, namespaceName string) string {
	return fmt.Sprintf("configfiles/%s/%s/%s?&ip=%s",
		url.QueryEscape(config.AppID),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		utils.GetInternal())
}

func (*apolloConfig) CallBack() http.CallBack {
	return http.CallBack{}
}

func (a *apolloConfig) Sync(appConfig *config.AppConfig) []*config.ApolloConfig {
	configs := make([]*config.ApolloConfig, 0, 8)
	config.SplitNamespaces(appConfig.NamespaceName, func(namespace string) {
		apolloConfig := a.SyncWithNamespace(namespace, appConfig)
		if apolloConfig == nil {
			return
		}

		configs = append(configs, apolloConfig)
	})
	return configs
}
