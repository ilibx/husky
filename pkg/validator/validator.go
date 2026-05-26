package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validator 参数验证器
type Validator struct {
	validate *validator.Validate
}

// NewValidator 创建验证器实例
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// Validate 验证结构体
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// ValidateField 验证单个字段
func (v *Validator) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// RegisterValidation 注册自定义验证规则
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// CommonTags 常用验证标签常量
const (
	TagRequired        = "required"
	TagEmail           = "email"
	TagMin             = "min=%d"
	TagMax             = "max=%d"
	TagMinLen          = "min_len=%d"
	TagMaxLen          = "max_len=%d"
	TagLen             = "len=%d"
	TagOneOf           = "oneof=%s"
	TagURL             = "url"
	TagHTTPURL         = "http_url"
	TagBase64          = "base64"
	TagHexadecimal     = "hexadecimal"
	TagNumeric         = "numeric"
	TagAlpha           = "alpha"
	TagAlphanumeric    = "alphanumeric"
	TagUUID            = "uuid"
	TagUUID3           = "uuid3"
	TagUUID4           = "uuid4"
	TagUUID5           = "uuid5"
	TagIP              = "ip"
	TagIPv4            = "ipv4"
	TagIPv6            = "ipv6"
	TagMAC             = "mac"
	TagDatetime        = "datetime=2006-01-02T15:04:05Z07:00"
	TagDate            = "2006-01-02"
	TagTime            = "15:04:05"
	TagBoolean         = "boolean"
	TagContains        = "contains=%s"
	TagStartsWith      = "startswith=%s"
	TagEndsWith        = "endswith=%s"
	TagNe              = "ne=%s"
	TagEq              = "eq=%s"
	TagGt              = "gt=%d"
	TagGte             = "gte=%d"
	TagLt              = "lt=%d"
	TagLte             = "lte=%d"
)

// ValidationErrors 验证错误响应
type ValidationErrors struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
}

// FormatValidationErrors 格式化验证错误
func FormatValidationErrors(err error) []ValidationErrors {
	var errors []ValidationErrors

	if err == nil {
		return errors
	}

	if _, ok := err.(validator.ValidationErrors); ok {
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationErrors{
				Field:   e.Field(),
				Message: getErrorMessage(e),
				Tag:     e.Tag(),
				Value:   e.Param(),
			})
		}
	}

	return errors
}

// getErrorMessage 获取错误消息
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "该字段为必填项"
	case "email":
		return "请输入有效的邮箱地址"
	case "min":
		return "最小值为 " + e.Param()
	case "max":
		return "最大值为 " + e.Param()
	case "len":
		return "长度必须为 " + e.Param()
	case "oneof":
		return "必须是以下值之一：" + e.Param()
	case "url":
		return "请输入有效的 URL"
	case "uuid":
		return "请输入有效的 UUID"
	default:
		return "验证失败"
	}
}
