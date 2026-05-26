package errors

import (
	"errors"
	"net/http"
)

// AppError 应用错误类型
type AppError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Err     error       `json:"-"`
	Data    interface{} `json:"data,omitempty"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap 返回底层错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 创建应用错误
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 常见错误码常量
const (
	// 通用错误
	ErrUnknown        = 10000
	ErrInvalidParam   = 10001
	ErrUnauthorized   = 10002
	ErrForbidden      = 10003
	ErrNotFound       = 10004
	ErrAlreadyExists  = 10005
	ErrInternalServer = 10006

	// 用户相关错误
	ErrUserNotFound     = 20001
	ErrUserExists       = 20002
	ErrInvalidPassword  = 20003
	ErrInvalidToken     = 20004
	ErrTokenExpired     = 20005

	// 工单相关错误
	ErrTicketNotFound    = 30001
	ErrTicketExists      = 30002
	ErrInvalidStatus     = 30003
	ErrInvalidTransition = 30004

	// 知识库相关错误
	ErrKnowledgeNotFound = 40001
	ErrKnowledgeExists   = 40002

	// Agent 相关错误
	ErrAgentNotFound = 50001
	ErrAgentExists   = 50002

	// 渠道相关错误
	ErrChannelNotFound = 60001
	ErrChannelConfig   = 60002
)

// 预定义错误
var (
	// 通用错误
	ErrUnknownError        = NewAppError(ErrUnknown, "未知错误", nil)
	ErrInvalidParameter    = NewAppError(ErrInvalidParam, "参数无效", nil)
	ErrUnauthorizedAccess  = NewAppError(ErrUnauthorized, "未授权访问", nil)
	ErrForbiddenAccess     = NewAppError(ErrForbidden, "禁止访问", nil)
	ErrResourceNotFound    = NewAppError(ErrNotFound, "资源不存在", nil)
	ErrResourceExists      = NewAppError(ErrAlreadyExists, "资源已存在", nil)
	ErrInternalServerError = NewAppError(ErrInternalServer, "服务器内部错误", nil)

	// 用户相关错误
	ErrUserNotFoundErr     = NewAppError(ErrUserNotFound, "用户不存在", nil)
	ErrUserExistsErr       = NewAppError(ErrUserExists, "用户已存在", nil)
	ErrInvalidPasswordErr  = NewAppError(ErrInvalidPassword, "密码无效", nil)
	ErrInvalidTokenErr     = NewAppError(ErrInvalidToken, "令牌无效", nil)
	ErrTokenExpiredErr     = NewAppError(ErrTokenExpired, "令牌已过期", nil)

	// 工单相关错误
	ErrTicketNotFoundErr   = NewAppError(ErrTicketNotFound, "工单不存在", nil)
	ErrTicketExistsErr     = NewAppError(ErrTicketExists, "工单已存在", nil)
	ErrInvalidStatusErr    = NewAppError(ErrInvalidStatus, "状态无效", nil)
	ErrInvalidTransitionErr = NewAppError(ErrInvalidTransition, "状态转换无效", nil)

	// 知识库相关错误
	ErrKnowledgeNotFoundErr = NewAppError(ErrKnowledgeNotFound, "知识文章不存在", nil)
	ErrKnowledgeExistsErr   = NewAppError(ErrKnowledgeExists, "知识文章已存在", nil)

	// Agent 相关错误
	ErrAgentNotFoundErr    = NewAppError(ErrAgentNotFound, "Agent 不存在", nil)
	ErrAgentExistsErr      = NewAppError(ErrAgentExists, "Agent 已存在", nil)

	// 渠道相关错误
	ErrChannelNotFoundErr  = NewAppError(ErrChannelNotFound, "渠道不存在", nil)
	ErrChannelConfigErr    = NewAppError(ErrChannelConfig, "渠道配置错误", nil)
)

// HTTPStatusCode 获取 HTTP 状态码
func HTTPStatusCode(appErr *AppError) int {
	switch appErr.Code {
	case ErrInvalidParam:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrInvalidToken, ErrTokenExpired:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound, ErrUserNotFound, ErrTicketNotFound, ErrKnowledgeNotFound, ErrAgentNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrUserExists, ErrTicketExists, ErrKnowledgeExists, ErrAgentExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Is 判断错误是否为目标错误
func Is(err, target error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		var targetAppErr *AppError
		if errors.As(target, &targetAppErr) {
			return appErr.Code == targetAppErr.Code
		}
	}
	return errors.Is(err, target)
}

// Wrap 包装错误
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    ErrUnknown,
		Message: message,
		Err:     err,
	}
}

// WithData 添加错误数据
func (e *AppError) WithData(data interface{}) *AppError {
	e.Data = data
	return e
}
