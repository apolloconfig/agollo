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

package sign

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"time"
)

const (
	// httpHeaderAuthorization is the HTTP header key for authorization
	httpHeaderAuthorization = "Authorization"
	// httpHeaderTimestamp is the HTTP header key for request timestamp
	httpHeaderTimestamp = "Timestamp"

	// authorizationFormat defines the format for Apollo authorization header
	// Format: "Apollo {appID}:{signature}"
	authorizationFormat = "Apollo %s:%s"

	// delimiter used to separate components in the string to be signed
	delimiter = "\n"
	// question mark used in URL query string
	question = "?"
)

var (
	// h is the default hash function (SHA1) used for signature generation
	h = sha1.New
)

// SetHash updates the hash function used for signature generation
// Parameters:
//   - f: New hash function to be used
//
// Returns:
//   - func() hash.Hash: The previous hash function
//
// This function allows for custom hash implementations to be used
func SetHash(f func() hash.Hash) func() hash.Hash {
	o := h
	h = f
	return o
}

// AuthSignature implements Apollo's signature-based authentication
// It provides functionality to generate authentication headers for Apollo API requests
type AuthSignature struct {
}

// HTTPHeaders generates the authentication headers for Apollo API requests
// Parameters:
//   - url: The target API endpoint URL
//   - appID: Application identifier
//   - secret: Authentication secret key
//
// Returns:
//   - map[string][]string: HTTP headers containing authorization and timestamp
//
// This method:
// 1. Generates current timestamp
// 2. Extracts path and query from URL
// 3. Creates signature using timestamp, path, and secret
// 4. Formats headers according to Apollo's requirements
func (t *AuthSignature) HTTPHeaders(url string, appID string, secret string) map[string][]string {
	ms := time.Now().UnixNano() / int64(time.Millisecond)
	timestamp := strconv.FormatInt(ms, 10)
	pathWithQuery := url2PathWithQuery(url)

	stringToSign := timestamp + delimiter + pathWithQuery
	signature := signString(stringToSign, secret)
	headers := make(map[string][]string, 2)

	signatures := make([]string, 0, 1)
	signatures = append(signatures, fmt.Sprintf(authorizationFormat, appID, signature))
	headers[httpHeaderAuthorization] = signatures

	timestamps := make([]string, 0, 1)
	timestamps = append(timestamps, timestamp)
	headers[httpHeaderTimestamp] = timestamps
	return headers
}

// signString generates a signature for the given string using HMAC-SHA1
// Parameters:
//   - stringToSign: The string to be signed
//   - accessKeySecret: The secret key for signing
//
// Returns:
//   - string: Base64-encoded signature
func signString(stringToSign string, accessKeySecret string) string {
	key := []byte(accessKeySecret)
	mac := hmac.New(h, key)
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// url2PathWithQuery extracts the path and query string from a URL
// Parameters:
//   - rawURL: The complete URL to parse
//
// Returns:
//   - string: Combined path and query string, or empty string if parsing fails
//
// Format: /path?query=value
func url2PathWithQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	pathWithQuery := u.Path

	if len(u.RawQuery) > 0 {
		pathWithQuery += question + u.RawQuery
	}
	return pathWithQuery
}
