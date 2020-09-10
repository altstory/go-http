package server

import (
	"context"
	"errors"

	"github.com/altstory/go-log"
	"github.com/altstory/go-runner"
)

// Hook 是一个 HTTP server 状态回调函数。
type Hook func(ctx context.Context, s *Server) error

var (
	startHooks    []Hook
	defaultServer *Server
)

func init() {
	runner.AddServer("http.server", func(ctx context.Context, config *Config) error {
		// 没有注册 hook 则直接跳过服务初始化。
		if len(startHooks) == 0 {
			return nil
		}

		if config == nil {
			return errors.New("go-http: missing http server config")
		}

		s := New(config)
		defaultServer = s

		for _, h := range startHooks {
			if err := h(ctx, s); err != nil {
				return err
			}
		}

		return s.Serve()
	})
}

// OnStart 注册一个回调，这个回调会在 HTTP server 初始化完成后且启动之前执行。
func OnStart(hook Hook) {
	if hook == nil {
		return
	}

	startHooks = append(startHooks, hook)
}

// AddRoutes 向默认 HTTP server 注册路由。
func AddRoutes(routes Routes) {
	OnStart(func(ctx context.Context, s *Server) error {
		if err := s.AddRoutes(routes); err != nil {
			log.Errorf(ctx, "err=%v||routes=%v||go-http: fail to add routes", err, routes)
			return err
		}

		return nil
	})
}

// Shutdown 关闭当前服务。
func Shutdown(ctx context.Context) error {
	if defaultServer == nil {
		return nil
	}

	return defaultServer.Shutdown(ctx)
}
