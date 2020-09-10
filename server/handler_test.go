package server

import (
	"context"
	"net/http"
	"testing"
)

type validBizSt1 struct{ Foo int }
type validBizSt2 struct{ Bar string }
type testContext interface {
	context.Context
}
type testError interface {
	error
}

func validBizFunc1(ctx context.Context, foo *validBizSt1) (bar *validBizSt2, err error) { return }
func validBizFunc2(ctx context.Context, foo validBizSt1) (bar *validBizSt2, err error)  { return }
func validBizFunc3(ctx context.Context, foo *validBizSt1) (bar validBizSt2, err error)  { return }
func validBizFunc4(ctx context.Context, foo validBizSt1) (bar validBizSt2, err error)   { return }
func validBizFunc5(ctx testContext, foo validBizSt1) (bar validBizSt2, err error)       { return }
func validBizFunc6(ctx testContext, foo validBizSt1) (bar validBizSt2, err testError)   { return }
func validBizFunc7(writer http.ResponseWriter, req *http.Request)                       {}

func TestParseValidBizFuncs(t *testing.T) {
	validFuncs := []Handler{validBizFunc1, validBizFunc2, validBizFunc3, validBizFunc4, validBizFunc5,
		validBizFunc6, validBizFunc7}
	hs, err := parseHandlersForGin(validFuncs)

	if err != nil {
		t.Fatalf("fail to parse handlers [err:%v]", err)
	}

	if len(hs) != len(validFuncs) {
		t.Fatalf("original handlers and parsed handlers mismatch. [expected:%v] [actual:%v]", len(validFuncs), len(hs))
	}
}

func invalidBizFunc1()                                                                       {}
func invalidBizFunc2(ctx context.Context)                                                    {}
func invalidBizFunc3(ctx context.Context, foo *validBizSt1) error                            { return nil }
func invalidBizFunc4(ctx context.Context, foo *validBizSt1) *validBizSt2                     { return nil }
func invalidBizFunc5(ctx context.Context, foo *validBizSt1) (b1, b2 *validBizSt2, err error) { return }
func invalidBizFunc6(ctx context.Context, f1, f2 *validBizSt1) (bar *validBizSt2, err error) { return }
func invalidBizFunc7(ctx context.Context, n int) (bar *validBizSt2, err error)               { return }
func invalidBizFunc8(ctx context.Context, foo *validBizSt1) (bar int, err error)             { return }
func invalidBizFunc9(a interface{}, b int) (c int, err error)                                { return }
func invalidBizFunc10(a string, b int) (c int, err error)                                    { return }
func invalidBizFunc11(ctx context.Context, foo *validBizSt1) (bar *validBizSt2, err int)     { return }
func invalidBizFunc12(ctx testContext, foo *validBizSt1) (bar *validBizSt2, err interface{}) { return }
func invalidBizFunc13(ctx context.Context, writer http.ResponseWriter, req *http.Request)    {}

func TestParseInvalidBizFuncs(t *testing.T) {
	invalidFuncs := []Handler{123, invalidBizFunc1, invalidBizFunc2, invalidBizFunc3, invalidBizFunc4, invalidBizFunc5,
		invalidBizFunc6, invalidBizFunc7, invalidBizFunc8, invalidBizFunc9, invalidBizFunc10,
		invalidBizFunc11, invalidBizFunc12, invalidBizFunc13}

	for _, f := range invalidFuncs {
		_, err := parseHandlerForGin(f)

		if err == nil {
			t.Fatalf("f should be invalid.")
		}
	}
}
