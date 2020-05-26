package extension

import (
	"testing"

	. "github.com/tevid/gohamcrest"
)

type TestAuth struct{}

func (a *TestAuth) HttpHeaders(url string, appId string, secret string) map[string][]string {
	return nil
}

func TestSetHttpAuth(t *testing.T) {
	SetHttpAuth(&TestAuth{})

	a := GetHttpAuth()

	b := a.(*TestAuth)
	Assert(t, b, NotNilVal())
}
