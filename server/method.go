package server

// Method 代表 HTTP 请求方法。
type Method string

func (m Method) String() string {
	return string(m)
}

// 各种 HTTP 请求方法。
const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	HEAD    Method = "HEAD"
	PATCH   Method = "PATCH"
	OPTIONS Method = "OPTIONS"

	ANY Method = "ANY" // 可以接受任意 HTTP 请求。
)
