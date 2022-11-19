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

package extension

import (
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/qshuai/agollo/v4/constant"
)

// TestParser 默认内容转换器
type TestParser struct {
}

// Parse 内存内容默认转换器
func (d *TestParser) Parse(s interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func TestAddFormatParser(t *testing.T) {
	AddFormatParser(constant.DEFAULT, &TestParser{})
	AddFormatParser(constant.Properties, &TestParser{})

	p := GetFormatParser(constant.DEFAULT)

	b := p.(*TestParser)
	Assert(t, b, NotNilVal())

	Assert(t, len(formatParser), Equal(2))
}
