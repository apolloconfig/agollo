// Copyright 2025 Apollo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package env

import (
	"time"
)

// ConnectConfig 网络请求配置
type ConnectConfig struct {
	//设置到http.client中timeout字段
	Timeout time.Duration
	//连接接口的uri
	URI string
	//是否重试
	IsRetry bool
	//appID
	AppID string
	//密钥
	Secret string
}
