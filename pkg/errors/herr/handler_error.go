package herr

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
)

// ResponseError 统一错误响应
func ResponseError(c *gin.Context, statusCode int, appErr *errors.AppError, message string) {
	if message == "" {
		message = appErr.Message
	}

	c.JSON(statusCode, gin.H{
		"code":    appErr.Code,
		"message": message,
		"data":    nil,
	})
}

// ResponseSuccess 统一成功响应
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// IsErrNotFound 判断是否为 NotFound 错误
func IsErrNotFound(err error) bool {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.Code == errors.ErrNotFound ||
			appErr.Code == errors.ErrUserNotFound ||
			appErr.Code == errors.ErrTicketNotFound ||
			appErr.Code == errors.ErrKnowledgeNotFound ||
			appErr.Code == errors.ErrAgentNotFound
	}
	return false
}

// IsErrUnauthorized 判断是否为 Unauthorized 错误
func IsErrUnauthorized(err error) bool {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.Code == errors.ErrUnauthorized ||
			appErr.Code == errors.ErrInvalidToken ||
			appErr.Code == errors.ErrTokenExpired
	}
	return false
}

// IsErrForbidden 判断是否为 Forbidden 错误
func IsErrForbidden(err error) bool {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.Code == errors.ErrForbidden
	}
	return false
}

// 导出常用错误变量，方便 handler 层使用
var (
	// 通用错误
	ErrUnknown        = errors.ErrUnknownError
	ErrInvalidParam   = errors.ErrInvalidParameter
	ErrUnauthorized   = errors.ErrUnauthorizedAccess
	ErrForbidden      = errors.ErrForbiddenAccess
	ErrNotFound       = errors.ErrResourceNotFound
	ErrAlreadyExists  = errors.ErrResourceExists
	ErrInternalServer = errors.ErrInternalServerError

	// 用户相关错误
	ErrUserNotFound     = errors.ErrUserNotFoundErr
	ErrUserExists       = errors.ErrUserExistsErr
	ErrInvalidPassword  = errors.ErrInvalidPasswordErr
	ErrInvalidToken     = errors.ErrInvalidTokenErr
	ErrTokenExpired     = errors.ErrTokenExpiredErr

	// 工单相关错误
	ErrTicketNotFound    = errors.ErrTicketNotFoundErr
	ErrTicketExists      = errors.ErrTicketExistsErr
	ErrInvalidStatus     = errors.ErrInvalidStatusErr
	ErrInvalidTransition = errors.ErrInvalidTransitionErr

	// 知识库相关错误
	ErrKnowledgeNotFound = errors.ErrKnowledgeNotFoundErr
	ErrKnowledgeExists   = errors.ErrKnowledgeExistsErr

	// Agent 相关错误
	ErrAgentNotFound = errors.ErrAgentNotFoundErr
	ErrAgentExists   = errors.ErrAgentExistsErr

	// 渠道相关错误
	ErrChannelNotFound = errors.ErrChannelNotFoundErr
	ErrChannelConfig   = errors.ErrChannelConfigErr
)
