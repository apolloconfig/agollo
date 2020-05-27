package extension

import (
	"github.com/zouyx/agollo/v3/protocol/auth"
)

var authSign auth.HTTPAuth

// SetHttpAuth 设置HttpAuth
func SetHTTPAuth(httpAuth auth.HTTPAuth) {
	authSign = httpAuth
}

// GetHttpAuth 获取HttpAuth
func GetHTTPAuth() auth.HTTPAuth {
	return authSign
}
