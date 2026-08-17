// Copyright 2026 Apollo Authors
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

package agollo

import "testing"

func TestParseYAMLPreservesKeyCase(t *testing.T) {
	values, err := parseYAML("myApp:\n  Timeout: 250ms\n", "yaml")
	if err != nil {
		t.Fatalf("parseYAML() error = %v", err)
	}
	if got := values["myApp.Timeout"]; got != "250ms" {
		t.Fatalf("exact-case value = %#v", got)
	}
	if _, exists := values["myapp.timeout"]; exists {
		t.Fatal("lower-cased YAML key was published")
	}
}
