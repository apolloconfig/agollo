package auth

// HTTPAuth http 授权
type HTTPAuth interface {
	// HTTPHeaders 根据 @url 获取 http 授权请求头
	HTTPHeaders(url string, appID string, secret string) map[string][]string
}
