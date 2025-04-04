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

package extension

import (
	"testing"

	. "github.com/tevid/gohamcrest"
)

type TestAuth struct{}

func (a *TestAuth) HTTPHeaders(url string, appID string, secret string) map[string][]string {
	return nil
}

func TestSetHttpAuth(t *testing.T) {
	SetHTTPAuth(&TestAuth{})

	a := GetHTTPAuth()

	b := a.(*TestAuth)
	Assert(t, b, NotNilVal())
}
