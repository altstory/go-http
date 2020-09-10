package server

const (
	// ErrCodeOK 代表业务正常。
	ErrCodeOK = 0

	// ErrCodeBadRequest 代表上游请求参数不合法。
	ErrCodeBadRequest = 1

	// ErrCodeInvalidError 代表业务返回了一个错误的 error 类型。
	// 业务应该始终使用 `Error()` 方法返回错误，而不能直接返回一个普通的 error。
	ErrCodeInvalidError = 2

	// ErrCodeServerPanic 代表业务代码崩溃，框架抓住这个错误并返回错误信息。
	ErrCodeServerPanic = 3
)
