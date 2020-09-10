package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/altstory/go-log"
	"github.com/altstory/go-runner"
)

// Handler 代表一个处理函数。
//
// Handler 支持的函数签名格式：
//     - func(ctx context.Context, req *T) (res *U, err error)：最推荐的业务函数签名形式。
//                                                              其中 `T` 和 `U` 是请求和应答的结构类型。
//     - func(writer http.ResponseWriter, req *http.Request)：如果需要使用更底层的能力，例如传输文件，可以使用这种形式。
//                                                            这个签名跟 http.HandlerFunc 一致。
//     - http.Handler：也可以直接注册一个 http.Handler 实例，应用场景同上。
type Handler interface{}

var (
	typeOfHTTPHandlerFunc = reflect.TypeOf(http.HandlerFunc(nil))
	typeOfContext         = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeOfError           = reflect.TypeOf((*error)(nil)).Elem()
)

func parseHandlersForGin(handlers []Handler) ([]gin.HandlerFunc, error) {
	hfs := make([]gin.HandlerFunc, 0, len(handlers))

	for _, h := range handlers {
		hf, err := parseHandlerForGin(h)

		if err != nil {
			return nil, err
		}

		hfs = append(hfs, hf)
	}

	return hfs, nil
}

// parseHandlerForGin 将 handler 解析成 gin 需要的处理函数形式。
func parseHandlerForGin(handler Handler) (gin.HandlerFunc, error) {
	if h, ok := handler.(http.Handler); ok {
		return wrapHTTPHandler(h)
	}
	v := reflect.ValueOf(handler)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return nil, errors.New("go-http: route's handler must be a func")
	}

	if t.ConvertibleTo(typeOfHTTPHandlerFunc) {
		hf := v.Convert(typeOfHTTPHandlerFunc).Interface().(http.HandlerFunc)
		return wrapHTTPHandlerFunc(hf)
	}

	in := t.NumIn()
	out := t.NumOut()

	// 当前我们只支持下列形式的自定义 handler：
	//     - func(ctx context.Context, in T) (out V, err error)
	if in != 2 || out != 2 {
		return nil, errors.New("go-http: type of the handler is not supported")
	}

	tIn0 := t.In(0)
	tIn1 := t.In(1)

	if tIn0.Kind() != reflect.Interface {
		return nil, errors.New("go-http: type of the handler is not supported")
	}

	if !tIn0.Implements(typeOfContext) || !typeOfContext.Implements(tIn0) {
		return nil, errors.New("go-http: type of the handler is not supported")
	}

	if tIn1.Kind() == reflect.Ptr {
		tIn1 = tIn1.Elem()
	}

	if tIn1.Kind() != reflect.Struct {
		return nil, errors.New("go-http: type of the handler is not supported")
	}

	tOut0 := t.Out(0)
	tOut1 := t.Out(1)

	if tOut0.Kind() == reflect.Ptr {
		tOut0 = tOut0.Elem()
	}

	if tOut0.Kind() != reflect.Struct {
		return nil, errors.New("go-http: type of the handler is not supported")
	}

	if tOut1.Kind() != reflect.Interface {
		return nil, errors.New("go-http: type of the handler is not supported")
	}

	if !tOut1.Implements(typeOfError) || !typeOfError.Implements(tOut1) {
		return nil, errors.New("go-http: type of the handler is not supported")
	}

	return wrapBusinessHandler(v)
}

func wrapHTTPHandler(h http.Handler) (gin.HandlerFunc, error) {
	if h == nil {
		return nil, errors.New("go-http: handler must be valid")
	}

	return wrapGinHandlerFunc(gin.WrapH(h)), nil
}

func wrapHTTPHandlerFunc(hf http.HandlerFunc) (gin.HandlerFunc, error) {
	if hf == nil {
		return nil, errors.New("go-http: handler must be valid")
	}

	return wrapGinHandlerFunc(gin.WrapF(hf)), nil
}

func wrapBusinessHandler(v reflect.Value) (gin.HandlerFunc, error) {
	in := v.Type().In(1)
	indirect := false

	if in.Kind() == reflect.Ptr {
		in = in.Elem()
		indirect = true
	}

	return wrapGinHandlerFunc(func(c *gin.Context) {
		ctx := c.Request.Context()
		vIn := reflect.New(in)

		if err := c.BindQuery(vIn.Interface()); err != nil {
			writeResponse(ctx, c, http.StatusBadRequest, newErrorMsg(ErrCodeBadRequest, fmt.Sprintf("go-http: fail to parse query with error: %v", err)).ToH(nil))
			return
		}

		if c.Request.Method != http.MethodGet && c.ContentType() == gin.MIMEJSON {
			if err := c.BindJSON(vIn.Interface()); err != nil {
				writeResponse(ctx, c, http.StatusBadRequest, newErrorMsg(ErrCodeBadRequest, fmt.Sprintf("go-http: invalid request content type or invalid JSON in body with error: %v", err)).ToH(nil))
				return
			}
		}

		if !indirect {
			vIn = vIn.Elem()
		}

		args := []reflect.Value{reflect.ValueOf(ctx), vIn}
		returns := v.Call(args)

		var data interface{}
		var err error

		if returns[0].IsValid() {
			data = returns[0].Interface()

			if returns[0].Kind() == reflect.Ptr && returns[0].IsNil() {
				data = nil
			}
		}

		if returns[1].IsValid() {
			err, _ = returns[1].Interface().(error)

			if err != nil {
				em, ok := err.(*errorMsg)

				if !ok {
					writeResponse(ctx, c, http.StatusInternalServerError, newErrorMsg(ErrCodeInvalidError, fmt.Sprintf("go-http: business returns an invalid error: %v", err)).ToH(nil))
					return
				}

				writeResponse(ctx, c, http.StatusOK, em.ToH(data))
				return
			}
		}

		writeResponse(ctx, c, http.StatusOK, newErrorMsg(ErrCodeOK, "").ToH(data))
	}), nil
}

type keyStartTimeType struct{}

var (
	keyStartTime keyStartTimeType
)

func wrapGinHandlerFunc(fn gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 往 ctx 里面放些东西。
		now := time.Now()
		traceid := now.UnixNano()
		ctx := context.Background()
		ctx = context.WithValue(ctx, keyStartTime, now)
		ctx = log.WithMoreInfo(ctx,
			log.Info{Key: "traceid", Value: traceid},
		)
		ctx = runner.WithStats(ctx, &runner.Stats{})

		defer func() {
			if r := recover(); r != nil {
				serverMetrics.Panic.Add(1)

				log.Errorf(ctx, "err=%v||url=%v||method=%v||go-http: caught a panic with call stack\n%v", r, c.Request.URL, c.Request.Method, string(debug.Stack()))
				writeResponse(ctx, c, http.StatusInternalServerError, newErrorMsg(ErrCodeServerPanic, fmt.Sprintf("go-http: caught a panic [err:%v]", r)).ToH(nil))
			}
		}()

		c.Request = c.Request.WithContext(ctx)
		log.Tracef(log.WithTag(ctx, "http.server.in"), "url=%v||method=%v||go-http: request starts",
			c.Request.URL.Path, c.Request.Method)
		fn(c)
	}
}

func writeResponse(ctx context.Context, c *gin.Context, status int, data gin.H) {
	c.JSON(status, data)

	start := ctx.Value(keyStartTime).(time.Time)
	proctime := time.Now().Sub(start)

	info := runner.StatsFromContext(ctx).Info()

	for i := range info {
		info[i].Key = "stats_" + info[i].Key
	}

	ctx = log.WithTag(ctx, "http.server.out")
	ctx = log.WithMoreInfo(ctx, info...)

	uri := c.Request.URL.Path
	code, hasErr := data["err"]
	proctimeMS := int64(proctime / time.Millisecond)

	httpMetrics.QPS.AddForTag(uri, 1)
	httpMetrics.Count.AddForTag(uri, 1)
	httpMetrics.ProcTime.AddForTag(uri, proctimeMS)
	httpMetrics.MaxProcTime.AddForTag(uri, proctimeMS)

	if c, ok := code.(int); !hasErr || !ok || c != 0 {
		httpMetrics.Failure.AddForTag(uri, 1)
	}

	log.Tracef(ctx, "url=%v||method=%v||code=%v||proctime=%.6f||go-http: request ends",
		uri, c.Request.Method, data["err"], proctime.Seconds())
}
