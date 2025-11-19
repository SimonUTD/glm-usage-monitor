package main

import (
	"fmt"
	"glm-usage-monitor/models"
	"glm-usage-monitor/services"
	"strconv"
	"strings"
	"time"
)

// ========== Bill Management API Bindings ==========

// GetBills retrieves expense bills with filtering and pagination
func (a *App) GetBills(filter interface{}) (*models.PaginatedResult, error) {
	// Convert the filter from interface{} to BillFilter
	var billFilter *models.BillFilter

	if filter != nil {
		// Try to convert to map[string]interface{}
		if filterMap, ok := filter.(map[string]interface{}); ok {
			billFilter = &models.BillFilter{
				PageNum:  1,
				PageSize: 20,
			}

			if pageNum, ok := filterMap["page_num"].(float64); ok {
				billFilter.PageNum = int(pageNum)
			}
			if pageSize, ok := filterMap["page_size"].(float64); ok {
				billFilter.PageSize = int(pageSize)
			}
			if modelName, ok := filterMap["model_name"].(string); ok && modelName != "" {
				billFilter.ModelName = &modelName
			}
		}
	} else {
		billFilter = &models.BillFilter{
			PageNum:  1,
			PageSize: 20,
		}
	}

	return a.apiService.GetBills(billFilter)
}

// GetBillByID retrieves a single expense bill by ID
func (a *App) GetBillByID(id string) (*models.ExpenseBill, error) {
	return a.apiService.GetBillByID(id)
}

// DeleteBill deletes an expense bill by ID
func (a *App) DeleteBill(id string) error {
	return a.apiService.DeleteBill(id)
}

// GetBillsByDateRange retrieves bills within a date range
func (a *App) GetBillsByDateRange(startDate, endDate time.Time, pageNum, pageSize int) (*models.PaginatedResult, error) {
	return a.apiService.GetBillsByDateRange(startDate, endDate, pageNum, pageSize)
}

// ========== Statistics API Bindings ==========

// GetStats retrieves overall usage statistics (IPC_03: 添加period参数)
func (a *App) GetStats(startDate, endDate *time.Time, period string) (*models.StatsResponse, error) {
	// 参数验证
	if period != "" {
		// 验证period参数的有效性
		validPeriods := []string{"today", "yesterday", "this_week", "last_week", "this_month", "last_month", "this_year", "last_year"}
		isValid := false
		for _, p := range validPeriods {
			if p == period {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, services.NewValidationError(services.ErrCodeInvalidParameter,
				fmt.Sprintf("invalid period: %s. Valid periods are: today, yesterday, this_week, last_week, this_month, last_month, this_year, last_year", period))
		}
	}

	result, err := a.apiService.GetStats(startDate, endDate, period)
	if err != nil {
		return nil, services.WrapError(err, services.ErrorTypeAPI, services.ErrCodeAPIInvalidResponse, "Failed to retrieve statistics")
	}

	return result, nil
}

// GetHourlyUsage retrieves hourly usage statistics
func (a *App) GetHourlyUsage(hours int) ([]models.HourlyUsageData, error) {
	return a.apiService.GetHourlyUsage(hours)
}

// GetModelDistribution retrieves usage distribution by model
func (a *App) GetModelDistribution(startDate, endDate *time.Time) ([]models.ModelDistributionData, error) {
	return a.apiService.GetModelDistribution(startDate, endDate)
}

// GetRecentUsage retrieves recent usage records
func (a *App) GetRecentUsage(limit int) ([]models.ExpenseBill, error) {
	return a.apiService.GetRecentUsage(limit)
}

// GetUsageTrend retrieves usage trend data
func (a *App) GetUsageTrend(days int) ([]models.HourlyUsageData, error) {
	return a.apiService.GetUsageTrend(days)
}

// ========== Token Management API Bindings ==========

// SaveToken saves an API token (IPC_02: 修复参数签名)
func (a *App) SaveToken(tokenName, tokenValue string) error {
	// 参数验证
	if tokenName == "" {
		return services.NewValidationError(services.ErrCodeInvalidParameter, "token name cannot be empty")
	}
	if tokenValue == "" {
		return services.NewValidationError(services.ErrCodeInvalidParameter, "token value cannot be empty")
	}

	err := a.apiService.SaveToken(tokenValue, tokenName, "", "")
	if err != nil {
		return services.WrapError(err, services.ErrorTypeDatabase, services.ErrCodeDBTransactionFailed, "failed to save token")
	}

	return nil
}

// GetToken retrieves the active API token
func (a *App) GetToken() (*models.APIToken, error) {
	token, err := a.apiService.GetToken()
	if err != nil {
		return nil, err
	}
	return token, nil
}

// GetAllTokens retrieves all API tokens
func (a *App) GetAllTokens() ([]models.APIToken, error) {
	return a.apiService.GetAllTokens()
}

// DeleteToken deletes an API token by ID
func (a *App) DeleteToken(id int) error {
	return a.apiService.DeleteToken(id)
}

// ValidateToken validates an API token
func (a *App) ValidateToken(token string) error {
	return a.apiService.ValidateToken(token)
}

// ValidateSavedToken validates the currently saved API token
func (a *App) ValidateSavedToken() (bool, error) {
	return a.apiService.ValidateSavedToken()
}

// ========== Sync Management API Bindings ==========

// GetSyncStatus retrieves current sync status
func (a *App) GetSyncStatus() (*models.SyncStatus, error) {
	// 直接返回 models.SyncStatus，避免类型转换问题
	return a.apiService.GetSyncStatus()
}

// GetSyncHistory retrieves sync history
func (a *App) GetSyncHistory(syncType string, pageNum, pageSize int) (*models.PaginatedResult, error) {
	return a.apiService.GetSyncHistory(syncType, pageNum, pageSize)
}

// SyncBills starts a sync operation for billing data (IPC_01: 修复参数签名)
func (a *App) SyncBills(billingMonth, syncType string) (map[string]interface{}, error) {
	// 参数验证
	if billingMonth == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Billing month is required",
		}, nil
	}

	if syncType == "" {
		syncType = "full"
	}

	// 验证billingMonth格式
	_, _, err := parseBillingMonth(billingMonth)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": services.GetErrorMessage(services.NewValidationError(services.ErrCodeInvalidParameter,
				fmt.Sprintf("Invalid billing month format: %v", err))),
		}, nil
	}

	// 调用服务层
	result, err := a.apiService.SyncBills(billingMonth, syncType, nil) // No progress callback for now
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": services.GetErrorMessage(err),
		}, nil
	}

	return map[string]interface{}{
		"success":     result.Success,
		"message":     "Sync started successfully",
		"syncedItems": result.SyncedItems,
		"totalItems":  result.TotalItems,
		"failedItems": result.FailedItems,
	}, nil
}

// parseBillingMonth 解析账单月份字符串 "2024-01" -> (2024, 1)
func parseBillingMonth(billingMonth string) (int, int, error) {
	parts := strings.Split(billingMonth, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format, expected YYYY-MM")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid year: %w", err)
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid month: %w", err)
	}

	if month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("month must be between 1 and 12")
	}

	return year, month, nil
}

// SyncRecentMonths syncs billing data for recent months
func (a *App) SyncRecentMonths(months int) (map[string]interface{}, error) {
	results, err := a.apiService.SyncRecentMonths(months, nil) // No progress callback for now
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, err
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Synced %d months", len(results)),
		"results": results,
	}, nil
}

// ========== Configuration Management API Bindings ==========

// GetConfig retrieves a configuration value
func (a *App) GetConfig(key string) (string, error) {
	return a.apiService.GetConfig(key)
}

// SetConfig saves a configuration value
func (a *App) SetConfig(key, value, description string) error {
	return a.apiService.SetConfig(key, value, description)
}

// GetAllConfigs retrieves all configuration values
func (a *App) GetAllConfigs() ([]models.AutoSyncConfig, error) {
	return a.apiService.GetAllConfigs()
}

// ========== Utility API Bindings ==========

// GetDatabaseInfo retrieves database information
func (a *App) GetDatabaseInfo() (map[string]interface{}, error) {
	return a.apiService.GetDatabaseInfo()
}

// CheckAPIConnectivity checks if the API is accessible
func (a *App) CheckAPIConnectivity() (map[string]interface{}, error) {
	return a.apiService.CheckAPIConnectivity()
}

// SaveSyncHistory saves sync history record (IPC_04: 完善saveSyncHistory方法)
func (a *App) SaveSyncHistory(syncType, billingMonth, status string, totalRecords, recordsSynced int, errorMessage *string) (map[string]interface{}, error) {
	// 参数验证
	if syncType == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Sync type cannot be empty",
		}, nil
	}
	if status == "" {
		return map[string]interface{}{
			"success": false,
			"message": "Status cannot be empty",
		}, nil
	}

	// Create sync history record
	history := &models.SyncHistory{
		SyncType:      syncType,
		StartTime:     time.Now(),
		Status:        status,
		TotalRecords:  totalRecords,
		RecordsSynced: recordsSynced,
		ErrorMessage:  errorMessage,
	}

	// Save to database via API service
	err := a.apiService.SaveSyncHistory(history)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to save sync history: " + err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"message": "Sync history saved successfully",
		"sync_id": history.ID,
	}, nil
}

// APIResponse provides unified response format for all IPC methods
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

// WrapResponse creates a standard API response
func WrapResponse(success bool, message string, data interface{}) *APIResponse {
	response := &APIResponse{
		Success: success,
		Message: message,
	}

	if data != nil {
		response.Data = data
	}

	return response
}

// WrapErrorResponse creates a standard error response
func WrapErrorResponse(message string, err error) *APIResponse {
	response := &APIResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		errMsg := err.Error()
		response.Error = &errMsg
	}

	return response
}
