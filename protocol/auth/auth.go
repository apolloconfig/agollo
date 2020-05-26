package auth

// HttpAuth http 授权
type HttpAuth interface {
	// HttpHeaders 根据 @url 获取 http 授权请求头
	HttpHeaders(url string, appId string, secret string) map[string][]string
}
