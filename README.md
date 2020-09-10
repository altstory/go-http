# go-http：HTTP 协议封装 #

`go-http` 是 HTTP 相关的各种框架封装，当前只封装了 HTTP server 的实现。

## 使用方法 ##

### 启动 `Server` ###

`Server` 的核心设计是让业务代码与底层框架完全解耦，不依赖于任何具体的第三方库，为了做到这一点，`Server` 通过 Go 反射来自动解析业务代码签名，自动完成框架处理 HTTP 请求到业务代码。

业务代码使用 `Server` 时候，只需要用 `go-runner` 的启动机制来完成初始化。

```go
import (
    "github.com/altstory/go-http/server"
    "github.com/altstory/go-runner"
    "github.com/project/server/routes"
)

func main() {
    // 将所有业务 route 注册到 http server 里面，
    // 一旦注册了 routes，http server 就被自动启动起来了。
    server.AddRoutes(routes.Routes)

    // 启动服务。
    runner.Main()
}
```

所有跟 `Server` 相关的配置放在配置文件的 `[http.server]` 字段里，可配置的项详见 [config.go](server/config.go)。

```ini
[http.server]
addr = ":8080"
```

### 实现业务函数 ###

`Server` 支持两种形式的路由配置：

* 通过 `server.RouteList` 定义的路由表，一般用来定义一个目录中的所有 HTTP 接口；
* 通过 `server.RouteMap` 定义的路由表，一般用来定义树状路由结构。

这两种路由配置是等价的，都可以通过 `server.AddRoutes` 注册到默认服务里面去。

一般推荐所有的业务 URI 设计成 `module/method` 形式，例如 `user/login`，推荐的目录结构如下：

```bash
./routes
  |-- user/
  |   |-- login.go
  |   |-- routelist.go
  |
  |-- routes.go
```

其中，`./routes/user/login.go` 里面定义了一个接口函数 `Login`。

```go
func Login(ctx context.Context, req *LoginRequest) (resp *LoginResponse, err error) {
    // ...
}
```

文件 `./routes/user/routelist.go` 将 `Login` 放入到 `RouteList` 里面。

```go
var RouteList = server.RouteList{
    server.R("login", server.POST, Login),
}
```

文件 `./routes/routes.go` 将引用所有的 `routelist.go` 并且形成树状结构。

```go
var Routes = server.RouteMap{
    "user": user.RouteList,
}
```

最后，将这个 `Routes` 通过 `server.AddRoutes` 加入到路由表里面去。

### 配置探针接口 ###

在 k8s 环境下，我们需要通过调用一个 HTTP 接口的方法来探测当前服务是否假死，为了方便运维，框架里内置了这个能力，只需要配置 `PingURI` 即可实现此功能。

```ini
[http.server]
ping_uri = "/ping"
```

有了这个配置之后，访问这个服务的 `/ping` 接口就可以得到一个 HTTP 200 OK 的应答。
