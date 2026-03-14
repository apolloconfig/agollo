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

package yaml

import (
	"testing"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/utils"
	"github.com/apolloconfig/agollo/v4/utils/parse"
)

var (
	yamlParser parse.ContentParser = &Parser{}
)

func TestYAMLParser(t *testing.T) {
	s, err := yamlParser.Parse(`
a:
    a1: a1
b:
    b1: b1
c:
    c1: c1
d:
    d1: d1
e:  
    e1: e1`)
	Assert(t, err, NilVal())

	Assert(t, s["a.a1"], Equal("a1"))

	Assert(t, s["b.b1"], Equal("b1"))

	Assert(t, s["c.c1"], Equal("c1"))

}

func TestYAMLParserOnException(t *testing.T) {
	s, err := yamlParser.Parse(utils.Empty)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())
	s, err = yamlParser.Parse(0)
	Assert(t, err, NilVal())
	Assert(t, s, NilVal())

	m := convertToMap(nil)
	Assert(t, m, NilVal())
}

// TestYAMLParserArray 测试 YAML 数组解析
func TestYAMLParserArray(t *testing.T) {
	s, err := yamlParser.Parse(`
items:
  - test111
  - test222
numbers:
  - 1
  - 2
  - 3
nested:
  items:
    - a
    - b
    - c
`)
	Assert(t, err, NilVal())

	// 验证字符串数组被解析为 []interface{}
	items, ok := s["items"].([]interface{})
	Assert(t, ok, Equal(true))
	Assert(t, len(items), Equal(2))
	Assert(t, items[0], Equal("test111"))
	Assert(t, items[1], Equal("test222"))

	// 验证数字数组被解析为 []interface{}
	numbers, ok := s["numbers"].([]interface{})
	Assert(t, ok, Equal(true))
	Assert(t, len(numbers), Equal(3))

	// 验证嵌套数组
	nestedItems, ok := s["nested.items"].([]interface{})
	Assert(t, ok, Equal(true))
	Assert(t, len(nestedItems), Equal(3))
	Assert(t, nestedItems[0], Equal("a"))
	Assert(t, nestedItems[1], Equal("b"))
	Assert(t, nestedItems[2], Equal("c"))
}
