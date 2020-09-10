package server

// Routes 是一个抽象的路由配置表。
type Routes interface {
	Register(router Router) error
}

// Route 是单条路由配置。
type Route struct {
	URI      string
	Method   Method
	Handlers []Handler
}

// R 生成一条路由记录。
func R(uri string, method Method, handlers ...Handler) *Route {
	return &Route{
		URI:      uri,
		Method:   method,
		Handlers: handlers,
	}
}

// RouteMap 是路由配置表。
type RouteMap map[string]Routes

// Register 将 rm 的路由配置注册到 router 里面去。
func (rm RouteMap) Register(router Router) error {
	for uri, routes := range rm {
		sub, err := router.SubRouter(uri)

		if err != nil {
			return err
		}

		routes.Register(sub)
	}

	return nil
}

// RouteList 是一个 Route 列表。
type RouteList []*Route

// Register 将 rl 的路由配置注册到 router 里面去。
func (rl RouteList) Register(router Router) error {
	var err error

	for _, r := range rl {
		switch r.Method {
		case ANY:
			err = router.HandleAny(r.URI, r.Handlers...)
		default:
			err = router.Handle(r.Method, r.URI, r.Handlers...)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
