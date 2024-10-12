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

package yaml

import (
	"github.com/xuxiaofan1101/agollo/v4/utils"
	"gopkg.in/yaml.v2"
)

// Parser properties转换器
type Parser struct {
}

// Parse 内存内容 => yml文件转换器
func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	content, ok := configContent.(string)
	if !ok {
		return nil, nil
	}
	if utils.Empty == content {
		return nil, nil
	}

	var result map[string]interface{}

	// 使用 yaml.v2 进行解析
	err := yaml.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
