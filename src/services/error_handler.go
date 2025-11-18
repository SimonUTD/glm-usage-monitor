package services

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

// ErrorType 定义错误类型
type ErrorType string

const (
	ErrorTypeAPI        ErrorType = "API_ERROR"
	ErrorTypeDatabase   ErrorType = "DATABASE_ERROR"
	ErrorTypeNetwork    ErrorType = "NETWORK_ERROR"
	ErrorTypeValidation ErrorType = "VALIDATION_ERROR"
	ErrorTypeAuth       ErrorType = "AUTH_ERROR"
	ErrorTypeSync       ErrorType = "SYNC_ERROR"
	ErrorTypeInternal   ErrorType = "INTERNAL_ERROR"
	ErrorTypeNotFound   ErrorType = "NOT_FOUND_ERROR"
)

// AppError 应用程序错误结构
type AppError struct {
	Type       ErrorType              `json:"type"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Cause      error                  `json:"-"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s - %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// Unwrap 支持错误链
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError 创建新的应用程序错误
func NewAppError(errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause 添加原因错误
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithStackTrace 添加堆栈跟踪
func (e *AppError) WithStackTrace() *AppError {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	e.StackTrace = string(buf[:n])
	return e
}

// ToJSON 将错误转换为JSON格式
func (e *AppError) ToJSON() ([]byte, error) {
	// 创建副本避免序列化Cause字段
	copy := *e
	copy.Cause = nil
	return json.Marshal(copy)
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(err error, context map[string]interface{})
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	Logger *log.Logger
}

// NewErrorHandler 创建错误处理器
func NewErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{
		Logger: log.New(log.Writer(), "[ERROR] ", log.LstdFlags|log.Lshortfile),
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err error, context map[string]interface{}) {
	if err == nil {
		return
	}

	var appErr *AppError
	if IsAppError(err) {
		appErr = err.(*AppError)
	} else {
		appErr = NewAppError(ErrorTypeInternal, "INTERNAL_ERROR", "Internal server error").
			WithCause(err).
			WithDetails(err.Error()).
			WithStackTrace()
	}

	// 添加上下文信息
	for k, v := range context {
		appErr.WithContext(k, v)
	}

	// 记录错误日志
	h.logError(appErr)
}

// logError 记录错误日志
func (h *DefaultErrorHandler) logError(err *AppError) {
	logData := map[string]interface{}{
		"type":      err.Type,
		"code":      err.Code,
		"message":   err.Message,
		"timestamp": err.Timestamp,
		"context":   err.Context,
	}

	if err.Details != "" {
		logData["details"] = err.Details
	}

	if err.StackTrace != "" {
		logData["stack_trace"] = err.StackTrace
	}

	jsonLog, _ := json.Marshal(logData)
	h.Logger.Printf("%s", string(jsonLog))
}

// IsAppError 检查是否为应用程序错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// WrapError 包装普通错误为应用程序错误
func WrapError(err error, errorType ErrorType, code, message string) *AppError {
	if err == nil {
		return nil
	}

	if IsAppError(err) {
		return err.(*AppError)
	}

	return NewAppError(errorType, code, message).
		WithCause(err).
		WithDetails(err.Error())
}

// 预定义的错误创建函数
func NewAPIError(code, message string) *AppError {
	return NewAppError(ErrorTypeAPI, code, message)
}

func NewDatabaseError(code, message string) *AppError {
	return NewAppError(ErrorTypeDatabase, code, message)
}

func NewNetworkError(code, message string) *AppError {
	return NewAppError(ErrorTypeNetwork, code, message)
}

func NewValidationError(code, message string) *AppError {
	return NewAppError(ErrorTypeValidation, code, message)
}

func NewAuthError(code, message string) *AppError {
	return NewAppError(ErrorTypeAuth, code, message)
}

func NewSyncError(code, message string) *AppError {
	return NewAppError(ErrorTypeSync, code, message)
}

func NewNotFoundError(code, message string) *AppError {
	return NewAppError(ErrorTypeNotFound, code, message)
}

func NewInternalError(code, message string) *AppError {
	return NewAppError(ErrorTypeInternal, code, message)
}

// 错误代码常量
const (
	// API错误代码
	ErrCodeAPITimeout          = "API_TIMEOUT"
	ErrCodeAPIRateLimit        = "API_RATE_LIMIT"
	ErrCodeAPIInvalidResponse  = "API_INVALID_RESPONSE"
	ErrCodeAPIUnauthorized     = "API_UNAUTHORIZED"

	// 数据库错误代码
	ErrCodeDBConnectionFailed  = "DB_CONNECTION_FAILED"
	ErrCodeDBQueryFailed       = "DB_QUERY_FAILED"
	ErrCodeDBTransactionFailed = "DB_TRANSACTION_FAILED"
	ErrCodeDBRecordNotFound    = "DB_RECORD_NOT_FOUND"

	// 网络错误代码
	ErrCodeNetworkTimeout      = "NETWORK_TIMEOUT"
	ErrCodeNetworkUnavailable  = "NETWORK_UNAVAILABLE"

	// 验证错误代码
	ErrCodeValidationFailed    = "VALIDATION_FAILED"
	ErrCodeInvalidParameter    = "INVALID_PARAMETER"

	// 认证错误代码
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeTokenExpired        = "TOKEN_EXPIRED"
	ErrCodeInvalidToken        = "INVALID_TOKEN"

	// 同步错误代码
	ErrCodeSyncFailed          = "SYNC_FAILED"
	ErrCodeSyncTimeout         = "SYNC_TIMEOUT"
	ErrCodeSyncInterrupted     = "SYNC_INTERRUPTED"
	ErrCodeSyncAlreadyRunning  = "SYNC_ALREADY_RUNNING"
	ErrCodeSyncNoToken         = "SYNC_NO_TOKEN"

	// 内部错误代码
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
	Backoff    float64
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxRetries: 3,
	Delay:      time.Second,
	Backoff:    2.0,
}

// RetryWithBackoff 带退避策略的重试机制
func RetryWithBackUp(config RetryConfig, operation func() error) error {
	var lastErr error
	delay := config.Delay

	for i := 0; i <= config.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.Backoff)
		}

		if err := operation(); err != nil {
			lastErr = err

			// 如果是最后一次尝试，直接返回错误
			if i == config.MaxRetries {
				return WrapError(err, ErrorTypeInternal, ErrCodeInternalError,
					fmt.Sprintf("Operation failed after %d retries", config.MaxRetries+1))
			}

			// 记录重试日志
			log.Printf("Retry %d/%d failed: %v", i, config.MaxRetries, err)
			continue
		}

		return nil
	}

	return lastErr
}

// SafeExecute 安全执行函数，捕获panic
func SafeExecute(operation func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewInternalError(ErrCodeInternalError, "Panic occurred in safe operation").
				WithDetails(fmt.Sprintf("Panic: %v", r)).
				WithStackTrace()
		}
	}()

	return operation()
}

// 错误恢复器
type ErrorRecovery struct {
	handler ErrorHandler
}

// NewErrorRecovery 创建错误恢复器
func NewErrorRecovery() *ErrorRecovery {
	return &ErrorRecovery{
		handler: NewErrorHandler(),
	}
}

// Recover 恢复错误
func (r *ErrorRecovery) Recover(err error, context map[string]interface{}) {
	r.handler.HandleError(err, context)
}

// GetErrorMessage 获取用户友好的错误消息
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}

	// 对于普通错误，返回简化的消息
	msg := err.Error()
	if len(msg) > 100 {
		msg = msg[:100] + "..."
	}
	return msg
}

// IsRetryable 判断错误是否可重试
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case ErrorTypeNetwork, ErrorTypeAPI:
			return strings.Contains(appErr.Code, "TIMEOUT") ||
				   strings.Contains(appErr.Code, "RATE_LIMIT")
		case ErrorTypeDatabase:
			return strings.Contains(appErr.Code, "CONNECTION")
		default:
			return false
		}
	}

	// 对于HTTP错误，检查状态码
	if strings.Contains(err.Error(), "timeout") ||
	   strings.Contains(err.Error(), "connection") {
		return true
	}

	return false
}