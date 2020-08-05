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

package sign

import (
	. "github.com/tevid/gohamcrest"
	"testing"
)

const (
	rawURL = "http://baidu.com/a/b?key=1"
	secret = "6ce3ff7e96a24335a9634fe9abca6d51"
	appID  = "testApplication_yang"
)

func TestSignString(t *testing.T) {
	s := signString(rawURL, secret)
	Assert(t, s, Equal("mcS95GXa7CpCjIfrbxgjKr0lRu8="))
}

func TestUrl2PathWithQuery(t *testing.T) {

	pathWithQuery := url2PathWithQuery(rawURL)

	Assert(t, pathWithQuery, Equal("/a/b?key=1"))
}

func TestHttpHeaders(t *testing.T) {
	a := &AuthSignature{}
	headers := a.HTTPHeaders(rawURL, appID, secret)

	Assert(t, headers, HasMapValue("Authorization"))
	Assert(t, headers, HasMapValue("Timestamp"))
}
