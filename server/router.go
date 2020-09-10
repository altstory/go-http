package server

import (
	"github.com/gin-gonic/gin"
)

// Router 代表一个路由器实现，Routes 可以向 Router 注册路由信息。
type Router interface {
	SubRouter(uri string, handlers ...Handler) (Router, error)
	Handle(method Method, uri string, handlers ...Handler) error
	HandleAny(uri string, handlers ...Handler) error
}

type ginRouter struct {
	router *gin.RouterGroup
}

func newGinRouter(router *gin.RouterGroup) *ginRouter {
	return &ginRouter{
		router: router,
	}
}

func (gr *ginRouter) SubRouter(uri string, handlers ...Handler) (Router, error) {
	hfs, err := parseHandlersForGin(handlers)

	if err != nil {
		return nil, err
	}

	router := gr.router.Group(uri, hfs...)
	return newGinRouter(router), nil
}

func (gr *ginRouter) Handle(method Method, uri string, handlers ...Handler) error {
	hfs, err := parseHandlersForGin(handlers)

	if err != nil {
		return err
	}

	gr.router.Handle(method.String(), uri, hfs...)
	return nil
}

func (gr *ginRouter) HandleAny(uri string, handlers ...Handler) error {
	hfs, err := parseHandlersForGin(handlers)

	if err != nil {
		return err
	}

	gr.router.Any(uri, hfs...)
	return nil
}
