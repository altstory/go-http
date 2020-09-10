package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type errorMsg struct {
	code int
	msg  string
	errs []error
}

func newErrorMsg(code int, msg string, errs ...error) *errorMsg {
	return &errorMsg{
		code: code,
		msg:  msg,
		errs: errs,
	}
}

func (em *errorMsg) Error() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "go-http: buniness error [code:%d] [msg:%s]", em.code, em.msg)

	for i, err := range em.errs {
		fmt.Fprintf(&sb, " [err-%d:%v]", i, err)
	}

	return sb.String()
}

func (em *errorMsg) ToH(data interface{}) gin.H {
	h := gin.H{
		"err": em.code,
		"now": time.Now().Format(time.RFC3339),
	}

	if em.code != 0 {
		h["msg"] = em.Error()
	}

	if data != nil {
		h["data"] = data
	}

	return h
}

// Error 构造一个带错误码的业务错误。
// 可以通过设置 errs，自动让框架在错误信息中附带各种系统错误的信息，方便调试。
func Error(code int, msg string, errs ...error) error {
	return newErrorMsg(code, msg, errs...)
}
