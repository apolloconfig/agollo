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
