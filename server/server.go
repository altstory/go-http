package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/altstory/go-log"
)

// Server 代表一个 HTTP 服务。
type Server struct {
	server *http.Server
	engine *gin.Engine

	addr string
}

// New 创建一个新的 HTTP 服务。
func New(config *Config) *Server {
	if config.MaxHeaderBytes <= 0 {
		config.MaxHeaderBytes = DefaultMaxHeaderBytes
	}

	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	// 如果设置了 ping uri，注册这个 uri。
	pingURI := config.PingURI

	if pingURI != "" {
		if !strings.HasPrefix(pingURI, "/") {
			pingURI = "/" + pingURI
		}

		engine.GET(pingURI, func(c *gin.Context) {
			c.Writer.WriteString("OK")
		})
	}

	return &Server{
		server: &http.Server{
			Addr:    config.Addr,
			Handler: engine,

			ReadTimeout:       config.ReadTimeout,
			ReadHeaderTimeout: config.ReadHeaderTimeout,
			WriteTimeout:      config.WriteTimeout,
			IdleTimeout:       config.IdleTimeout,
			MaxHeaderBytes:    config.MaxHeaderBytes,
		},
		engine: engine,

		addr: config.Addr,
	}
}

// AddRoutes 将 routes 路由信息添加到路有里面去。
func (s *Server) AddRoutes(routes Routes) error {
	router := newGinRouter(&s.engine.RouterGroup)
	return routes.Register(router)
}

// MustAddRoutes 将 routes 路由信息添加到路有里面去，如果过程中发生任何错误，直接 panic。
// 由于一般来说 routes 格式错误都是程序 bug，所以这个函数可以简化业务代码，无需额外判断一个 error。
func (s *Server) MustAddRoutes(routes Routes) {
	if err := s.AddRoutes(routes); err != nil {
		log.Fatalf(context.Background(), "err=%v||go-http: fail to add routes", err)
	}
}

// Serve 开始提供 HTTP 服务。这个函数永远不会返回，直到 HTTP 服务终止。
func (s *Server) Serve() error {
	errs := make(chan error, 1)
	go func() {
		// 开始提供服务。
		log.Tracef(context.Background(), "addr=%v||http server is starting...", s.addr)

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errs <- err
		} else {
			errs <- nil
		}
	}()

	// 开始统计 goroutine 信息。
	exitTicker := make(chan bool, 1)
	metricsTicker := time.NewTicker(20 * time.Second)
	defer func() {
		metricsTicker.Stop()
		exitTicker <- true
	}()
	go func() {
		for {
			select {
			case <-metricsTicker.C:
				serverMetrics.Goroutine.Add(int64(runtime.NumGoroutine()))
			case <-exitTicker:
				return
			}
		}
	}()

	// 进行 graceful shutdown。
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(c)

	select {
	case err := <-errs:
		return err
	case <-c:
	}

	// 关闭服务器。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return <-errs
}

// Shutdown 关闭 HTTP 服务。
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Handler 返回一个 http.Handler 用于在外部启动服务。
func (s *Server) Handler() http.Handler {
	return s.engine
}
