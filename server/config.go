package server

import (
	"net/http"
	"time"
)

const (
	// DefaultMaxHeaderBytes 是默认的 HTTP header 大小。
	DefaultMaxHeaderBytes = http.DefaultMaxHeaderBytes
)

// Config 是 HTTP server 的配置。
type Config struct {
	Addr string `config:"addr"` // 服务器监听的地址。

	ReadTimeout       time.Duration `config:"read_timeout"`        // ReadTimeout 设置读 HTTP 数据超时。
	ReadHeaderTimeout time.Duration `config:"read_header_timeout"` // ReadHeaderTimeout 设置读 HTTP header 超时。
	WriteTimeout      time.Duration `config:"write_timeout"`       // WriteTimeout 设置写超时。
	IdleTimeout       time.Duration `config:"idle_timeout"`        // IdleTimeout 设置空闲超时。
	MaxHeaderBytes    int           `config:"max_header_bytes"`    // MaxHeaderBytes 设置 HTTP header 最大大小，默认是 DefaultMaxHeaderBytes。

	Debug bool `config:"debug"` // Debug 表示是否处于调试状态。

	PingURI string `config:"ping_uri"` // PingURI 表示用作探针的 uri 地址，这个接口会在服务正常的时候返回 HTTP 200 OK。
}
