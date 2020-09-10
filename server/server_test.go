package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huandu/go-assert"
)

var testServerCalled int

const (
	testUsername  = "huandu"
	testPassword  = "a-passport"
	testUID       = 906
	testUIDString = "906"
	testToken     = "a-token"
	testProjectID = 888
	testLimit     = 20
)

type testLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"passport"`
}

type testValidateRequest struct {
	UID   int64  `form:"uid" json:"-"`
	Token string `json:"token"`
}

type testProjectListRequest struct {
	UID       int64 `form:"uid" json:"-"`
	ProjectID int64 `json:"project_id"`
	Limit     int   `json:"limit"`
}

type testCommonResponse struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func newTestCommonResponse() *testCommonResponse {
	return &testCommonResponse{
		Foo: "foo",
		Bar: 1234,
	}
}

func testLogin(ctx context.Context, req testLoginRequest) (res *testCommonResponse, err error) {
	if req.Username != testUsername || req.Password != testPassword {
		return nil, Error(ErrCodeInvalidError, "failed")
	}

	testServerCalled++
	return newTestCommonResponse(), nil
}

func testValidate(ctx context.Context, req *testValidateRequest) (res testCommonResponse, err error) {
	if req.UID != testUID || req.Token != testToken {
		return res, errors.New("bad error")
	}

	testServerCalled++
	return *newTestCommonResponse(), nil
}

func testProjectList(ctx context.Context, req *testProjectListRequest) (res *testCommonResponse, err error) {
	if req.UID != testUID || req.ProjectID != testProjectID || req.Limit != testLimit {
		return nil, Error(ErrCodeInvalidError, "failed")
	}

	testServerCalled++
	return newTestCommonResponse(), nil
}

func TestServerRoutes(t *testing.T) {
	a := assert.New(t)
	testServerCalled = 0
	config := RouteMap{
		"/passport": RouteList{
			R("login", POST, testLogin),
			R("validate", ANY, testValidate),
		},
		"/project": RouteList{
			R("list", POST, testProjectList),
		},
	}

	pingURI := "/ping/it"
	server := New(&Config{
		PingURI: pingURI,
		Debug:   true,
	})
	server.AddRoutes(config)
	testServer := httptest.NewServer(server.Handler())
	defer testServer.Close()
	prefix := testServer.URL
	client := testServer.Client()

	const contentType = "application/json"
	called := 0

	// 下面的请求必须符合路由表中各个函数的具体实现，如果要修改，需要仔细阅读每个业务函数的具体实现。
	resp, err := client.Post(prefix+"/passport/login", contentType, toJSON(m{
		"username": testUsername,
		"passport": testPassword,
	}))
	called++
	a.NilError(err)

	validateResponse(a, resp)
	resp, err = client.Post(prefix+"/passport/validate?uid="+testUIDString, contentType, toJSON(m{
		"token": testToken,
	}))
	called++
	a.NilError(err)

	validateResponse(a, resp)
	resp, err = client.Post(prefix+"/project/list?uid="+testUIDString, contentType, toJSON(m{
		"project_id": testProjectID,
		"limit":      testLimit,
	}))
	called++
	a.NilError(err)
	validateResponse(a, resp)

	// 检查结果。
	a.Equal(called, testServerCalled)

	// 下面测试各种异常情况。

	// 非法 JSON。
	resp, err = client.Post(prefix+"/project/list?uid="+testUIDString, contentType, bytes.NewBuffer([]byte(`{"project_id"}`)))
	a.NilError(err)
	validateErrorResponse(a, resp)

	// 错误的在 JSON 里面传 UID。
	resp, err = client.Post(prefix+"/project/list", contentType, toJSON(m{
		"uid":        testUID, // 应该解析不出来。
		"project_id": testProjectID,
		"limit":      testLimit,
	}))
	a.NilError(err)
	validateFailedResponse(a, resp)

	// 忘记传 UID，并且处理函数返回了不符合规则的错误。
	resp, err = client.Post(prefix+"/passport/validate", contentType, toJSON(m{
		"token": testToken,
	}))
	a.NilError(err)
	validateFailedResponse(a, resp)

	// 尝试 ping。
	resp, err = client.Get(prefix + pingURI)
	a.NilError(err)
	a.Equal(resp.StatusCode, http.StatusOK)
}

type m map[string]interface{}

func toJSON(m m) io.Reader {
	data, _ := json.Marshal(m)
	return bytes.NewBuffer(data)
}

func validateResponse(a *assert.A, resp *http.Response) {
	expected := m{
		"err": 0.0,
		"data": map[string]interface{}{
			"foo": "foo",
			"bar": 1234.0,
		},
	}
	var actual m
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	a.NilError(err)

	json.Unmarshal(data, &actual)
	delete(actual, "now") // 不需要测试 now，这个是当前时间，会不断变化。
	a.Equal(expected, actual)
}

func validateFailedResponse(a *assert.A, resp *http.Response) {
	expected := m{
		"err": float64(ErrCodeInvalidError),
	}
	var actual m
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	a.NilError(err)

	json.Unmarshal(data, &actual)
	a.Equal(actual["err"], expected["err"])
}

func validateErrorResponse(a *assert.A, resp *http.Response) {
	expected := m{
		"err": float64(ErrCodeBadRequest),
	}
	expectedStatusCode := http.StatusBadRequest

	var actual m
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	a.NilError(err)
	a.Equal(resp.StatusCode, expectedStatusCode)

	json.Unmarshal(data, &actual)
	a.Equal(actual["err"], expected["err"])
}
