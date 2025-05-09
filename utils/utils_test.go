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

package utils

import (
	"strings"
	"testing"

	. "github.com/tevid/gohamcrest"
)

func TestGetInternal(t *testing.T) {
	ip := GetInternal()

	t.Log("Internal ip:", ip)

	//只能在有网络下开启者配置,否则跑出错误
	Assert(t, ip, NotEqual(Empty))
	nums := strings.Split(ip, ".")

	Assert(t, true, Equal(len(nums) > 0))
}

func TestIsNotNil(t *testing.T) {
	flag := IsNotNil(nil)
	Assert(t, false, Equal(flag))

	flag = IsNotNil("")
	Assert(t, true, Equal(flag))
}
