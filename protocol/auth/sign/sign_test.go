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
	Assert(t, s, Equal("99c4bde465daec2a428c87eb6f18232abd2546ef"))
}

func TestUrl2PathWithQuery(t *testing.T) {

	pathWithQuery := url2PathWithQuery(rawURL)

	Assert(t, pathWithQuery, Equal("/a/b?key=1"))
}

func TestHttpHeaders(t *testing.T) {
	a := &AuthSignature{}
	headers := a.HttpHeaders(rawURL, appID, secret)
	t.Log(headers)

	Assert(t, headers, HasMapValue("Authorization"))
	Assert(t, headers, HasMapValue("Timestamp"))
}
