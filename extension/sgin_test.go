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
