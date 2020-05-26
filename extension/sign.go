package extension

import (
	"github.com/zouyx/agollo/v3/protocol/auth"
)

var authSign auth.HttpAuth

// SetHttpAuth 设置HttpAuth
func SetHttpAuth(httpAuth auth.HttpAuth) {
	authSign = httpAuth
}

// GetHttpAuth 获取HttpAuth
func GetHttpAuth() auth.HttpAuth {
	return authSign
}
