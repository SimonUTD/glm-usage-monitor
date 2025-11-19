package middleware

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationResult contains validation results
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// ValidateBillingMonth validates billing month format (VALIDATION_01: 账单月份格式验证)
func ValidateBillingMonth(month string) error {
	if !regexp.MustCompile(`^\d{4}-\d{2}$`).MatchString(month) {
		return errors.New("invalid billing month format, expected YYYY-MM")
	}

	// 验证月份范围
	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		return errors.New("invalid billing month format")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 2020 || year > 2030 {
		return errors.New("invalid year range, expected 2020-2030")
	}

	monthNum, err := strconv.Atoi(parts[1])
	if err != nil || monthNum < 1 || monthNum > 12 {
		return errors.New("invalid month range, expected 1-12")
	}

	return nil
}

// ValidateTokenFormat validates token format (name:value)
func ValidateTokenFormat(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return errors.New("invalid token format, expected 'name:value'")
	}

	tokenName := strings.TrimSpace(parts[0])
	tokenValue := strings.TrimSpace(parts[1])

	if tokenName == "" {
		return errors.New("token name cannot be empty")
	}

	if tokenValue == "" {
		return errors.New("token value cannot be empty")
	}

	// 验证token长度
	if len(tokenValue) < 10 {
		return errors.New("token value too short, minimum 10 characters")
	}

	if len(tokenValue) > 500 {
		return errors.New("token value too long, maximum 500 characters")
	}

	return nil
}

// ValidateDateRange validates date range
func ValidateDateRange(startDate, endDate *time.Time) error {
	if startDate == nil || endDate == nil {
		return errors.New("start date and end date cannot be nil")
	}

	if startDate.After(*endDate) {
		return errors.New("start date cannot be after end date")
	}

	// 验证日期范围不能超过1年
	maxDuration := 365 * 24 * time.Hour // 1年
	if endDate.Sub(*startDate) > maxDuration {
		return errors.New("date range cannot exceed 1 year")
	}

	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(pageNum, pageSize int) error {
	if pageNum < 1 {
		return errors.New("page number must be greater than 0")
	}

	if pageSize < 1 || pageSize > 100 {
		return errors.New("page size must be between 1 and 100")
	}

	return nil
}

// ValidateModelName validates model name
func ValidateModelName(modelName string) error {
	if modelName == "" {
		return nil // 允许空字符串，表示不筛选
	}

	if len(modelName) > 100 {
		return errors.New("model name too long, maximum 100 characters")
	}

	// 验证只包含允许的字符
	if !regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`).MatchString(modelName) {
		return errors.New("model name contains invalid characters")
	}

	return nil
}

// ValidateChargeType validates charge type
func ValidateChargeType(chargeType string) error {
	if chargeType == "" {
		return nil // 允许空字符串，表示不筛选
	}

	validTypes := []string{"API调用", "模型调用", "数据存储", "其他"}
	for _, validType := range validTypes {
		if chargeType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid charge type: %s", chargeType)
}

// ValidateAPIUsage validates API usage parameters
func ValidateAPIUsage(used, limit int) error {
	if used < 0 {
		return errors.New("API usage cannot be negative")
	}

	if limit < 0 {
		return errors.New("API limit cannot be negative")
	}

	if used > limit {
		return fmt.Errorf("API usage (%d) cannot exceed limit (%d)", used, limit)
	}

	return nil
}

// ValidateCost validates cost amount
func ValidateCost(cost float64) error {
	if cost < 0 {
		return errors.New("cost cannot be negative")
	}

	// 验证成本不能超过合理范围
	maxCost := 10000.0 // 最大成本10000元
	if cost > maxCost {
		return fmt.Errorf("cost (%.2f) cannot exceed maximum (%.2f)", cost, maxCost)
	}

	return nil
}

// ValidateSyncConfig validates sync configuration
func ValidateSyncConfig(config map[string]interface{}) error {
	if config == nil {
		return errors.New("sync config cannot be nil")
	}

	// 验证必需字段
	if enabled, ok := config["enabled"]; !ok {
		return errors.New("enabled field is required")
	}

	if frequency, ok := config["frequency_seconds"]; !ok {
		return errors.New("frequency_seconds field is required")
	}

	// 验证enabled为布尔值
	if enabledValue := config["enabled"]; enabledValue != nil {
		if enabledBool, ok := enabledValue.(bool); !ok {
			return errors.New("enabled must be a boolean")
		} else if !enabledBool {
			// 如果禁用自动同步，其他字段可以为空
			return nil
		}
	}

	// 验证frequency为正整数
	if frequencyValue := config["frequency_seconds"]; frequencyValue != nil {
		if frequencyInt, ok := frequencyValue.(int); !ok {
			return errors.New("frequency_seconds must be an integer")
		} else if frequencyInt < 60 || frequencyInt > 86400 {
			return errors.New("frequency_seconds must be between 60 and 86400 (1 minute to 24 hours)")
		}
	}

	return nil
}

// ValidateID validates ID parameter
func ValidateID(id int) error {
	if id <= 0 {
		return errors.New("ID must be greater than 0")
	}

	return nil
}

// ValidateString validates string parameter
func ValidateString(value, fieldName string, required bool, maxLength int) error {
	if required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if len(value) > maxLength {
		return fmt.Errorf("%s too long, maximum %d characters", fieldName, maxLength)
	}

	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return nil // 允许空字符串
	}

	// 简单的邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// ValidatePhoneNumber validates phone number format
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return nil // 允许空字符串
	}

	// 简单的手机号格式验证（只包含数字和常见符号）
	phoneRegex := regexp.MustCompile(`^[0-9\-\+\(\)\s]+$`)
	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone number format")
	}

	if len(phone) < 8 || len(phone) > 20 {
		return errors.New("phone number length must be between 8 and 20 characters")
	}

	return nil
}
