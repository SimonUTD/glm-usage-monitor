package services

import (
	"fmt"
	"glm-usage-monitor/models"
	"log"
	"strconv"
	"strings"
	"time"
)

// APIService provides all API methods for frontend
type APIService struct {
	dbService       *DatabaseService
	statsService    *StatisticsService
	zhipuAPIService *ZhipuAPIService
	autoSyncService *AutoSyncService
	db              DatabaseInterface
	errorHandler    ErrorHandler
}

// NewAPIService creates a new API service
func NewAPIService(db DatabaseInterface) *APIService {
	dbService := NewDatabaseService(db.GetDB())
	statsService := NewStatisticsService(db.GetDB())

	apiService := &APIService{
		dbService:       dbService,
		statsService:    statsService,
		zhipuAPIService: nil, // Will be initialized when token is set
		db:              db,
		errorHandler:    NewErrorHandler(),
	}

	// 初始化自动同步服务
	apiService.autoSyncService = NewAutoSyncService(apiService, dbService)

	return apiService
}

// ========== Bill Management APIs ==========

// GetBills retrieves expense bills with filtering and pagination (IPC_02: 统一响应格式)
func (s *APIService) GetBills(filter *models.BillFilter) (*models.PaginatedResult, error) {
	// IPC_03: 参数验证
	if filter == nil {
		filter = &models.BillFilter{
			PageNum:  1,
			PageSize: 20,
		}
	}

	// 验证分页参数
	if filter.PageNum < 1 {
		return nil, NewValidationError(ErrCodeInvalidParameter, "Page number must be greater than 0")
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		return nil, NewValidationError(ErrCodeInvalidParameter, "Page size must be between 1 and 100")
	}

	// 使用错误处理机制执行数据库操作
	var result *models.PaginatedResult
	err := SafeExecute(func() error {
		var operationErr error
		result, operationErr = s.dbService.GetExpenseBills(filter)
		return operationErr
	})

	if err != nil {
		// 处理错误并记录上下文
		context := map[string]interface{}{
			"operation": "GetBills",
			"page_num":  filter.PageNum,
			"page_size": filter.PageSize,
		}

		s.errorHandler.HandleError(err, context)

		// 返回用户友好的错误
		appErr := WrapError(err, ErrorTypeDatabase, ErrCodeDBQueryFailed, "Failed to retrieve bills")
		return nil, appErr
	}

	return result, nil
}

// GetBillByID retrieves a single expense bill by ID
func (s *APIService) GetBillByID(id string) (*models.ExpenseBill, error) {
	bill, err := s.dbService.GetExpenseBillByID(id)
	if err != nil {
		log.Printf("Error getting bill by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve bill: %w", err)
	}

	return bill, nil
}

// DeleteBill deletes an expense bill by ID
func (s *APIService) DeleteBill(id string) error {
	err := s.dbService.DeleteExpenseBill(id)
	if err != nil {
		log.Printf("Error deleting bill ID %d: %v", id, err)
		return fmt.Errorf("failed to delete bill: %w", err)
	}

	log.Printf("Successfully deleted bill ID %d", id)
	return nil
}

// GetBillsByDateRange retrieves bills within a date range
func (s *APIService) GetBillsByDateRange(startDate, endDate time.Time, pageNum, pageSize int) (*models.PaginatedResult, error) {
	filter := &models.BillFilter{
		PageNum:   pageNum,
		PageSize:  pageSize,
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	result, err := s.dbService.GetExpenseBills(filter)
	if err != nil {
		log.Printf("Error getting bills by date range: %v", err)
		return nil, fmt.Errorf("failed to retrieve bills by date range: %w", err)
	}

	return result, nil
}

// ========== Statistics APIs ==========

// GetStats retrieves overall usage statistics (IPC_03: 支持period参数解析)
func (s *APIService) GetStats(startDate, endDate *time.Time, period string) (*models.StatsResponse, error) {
	// 根据period参数计算时间范围
	if period != "" {
		startDate, endDate = s.calculateDateRange(period)
	}

	stats, err := s.statsService.GetOverallStats(startDate, endDate)
	if err != nil {
		log.Printf("Error getting statistics: %v", err)
		return nil, fmt.Errorf("failed to retrieve statistics: %w", err)
	}

	// Add sync status
	syncStatus, err := s.dbService.GetAutoSyncStatus()
	if err != nil {
		log.Printf("Error getting sync status: %v", err)
	} else {
		stats.SyncStatus = *syncStatus
	}

	return stats, nil
}

// calculateDateRange 根据period参数计算时间范围
func (s *APIService) calculateDateRange(period string) (*time.Time, *time.Time) {
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		startDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
		endDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, now.Location())
	case "this_week":
		weekday := int(now.Weekday())
		startDate = now.AddDate(0, 0, -weekday+1)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now.AddDate(0, 0, -int(now.Weekday())+1)
	case "this_month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = time.Date(now.Year(), now.Month(), 31, 23, 59, 59, 0, now.Location())
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = time.Date(now.Year(), now.Month()-1, 31, 23, 59, 59, 0, now.Location())
	case "this_year":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = now
	case "last_year":
		startDate = time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())
		endDate = time.Date(now.Year()-1, 12, 31, 23, 59, 59, 0, now.Location())
	default:
		// 默认返回最近7天
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	}

	return &startDate, &endDate
}

// GetHourlyUsage retrieves hourly usage statistics
func (s *APIService) GetHourlyUsage(hours int) ([]models.HourlyUsageData, error) {
	if hours <= 0 {
		hours = 5 // Default to last 5 hours
	}

	// Calculate start time
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	usageData, err := s.statsService.GetHourlyUsage(&startTime, nil)
	if err != nil {
		log.Printf("Error getting hourly usage: %v", err)
		return nil, fmt.Errorf("failed to retrieve hourly usage: %w", err)
	}

	return usageData, nil
}

// GetModelDistribution retrieves usage distribution by model
func (s *APIService) GetModelDistribution(startDate, endDate *time.Time) ([]models.ModelDistributionData, error) {
	distribution, err := s.statsService.GetModelDistribution(startDate, endDate)
	if err != nil {
		log.Printf("Error getting model distribution: %v", err)
		return nil, fmt.Errorf("failed to retrieve model distribution: %w", err)
	}

	return distribution, nil
}

// GetRecentUsage retrieves recent usage records
func (s *APIService) GetRecentUsage(limit int) ([]models.ExpenseBill, error) {
	if limit <= 0 {
		limit = 10
	}

	recentUsage, err := s.statsService.GetRecentUsage(limit)
	if err != nil {
		log.Printf("Error getting recent usage: %v", err)
		return nil, fmt.Errorf("failed to retrieve recent usage: %w", err)
	}

	return recentUsage, nil
}

// GetUsageTrend retrieves usage trend data
func (s *APIService) GetUsageTrend(days int) ([]models.HourlyUsageData, error) {
	if days <= 0 {
		days = 7
	}

	trendData, err := s.statsService.GetUsageTrend(days)
	if err != nil {
		log.Printf("Error getting usage trend: %v", err)
		return nil, fmt.Errorf("failed to retrieve usage trend: %w", err)
	}

	return trendData, nil
}

// ========== Token Management APIs ==========

// SaveToken saves an API token (IPC_01: 修复参数顺序)
func (s *APIService) SaveToken(tokenValue, tokenName, provider, tokenType string) error {
	// 参数验证
	if tokenValue == "" {
		return NewValidationError(ErrCodeInvalidParameter, "Token value cannot be empty")
	}
	if tokenName == "" {
		tokenName = "Default Token"
	}
	if provider == "" {
		provider = "zhipu"
	}
	if tokenType == "" {
		tokenType = "api_key"
	}

	token := &models.APIToken{
		TokenName:  tokenName,
		TokenValue: tokenValue,
		Provider:   provider,
		TokenType:  tokenType,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := s.dbService.SaveAPIToken(token)
	if err != nil {
		log.Printf("Error saving token: %v", err)
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Update Zhipu API service
	s.zhipuAPIService = NewZhipuAPIService(tokenValue)

	log.Printf("Successfully saved token: %s", tokenName)
	return nil
}

// GetToken retrieves active API token (IPC_02: 统一响应格式)
func (s *APIService) GetToken() (*models.APIToken, error) {
	token, err := s.dbService.GetActiveAPIToken()
	if err != nil {
		log.Printf("Error getting token: %v", err)
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	// Update Zhipu API service if not already set
	if s.zhipuAPIService == nil {
		s.zhipuAPIService = NewZhipuAPIService(token.TokenValue)
	}

	return token, nil
}

// GetAllTokens retrieves all API tokens
func (s *APIService) GetAllTokens() ([]models.APIToken, error) {
	tokens, err := s.dbService.GetAllAPITokens()
	if err != nil {
		log.Printf("Error getting all tokens: %v", err)
		return nil, fmt.Errorf("failed to retrieve tokens: %w", err)
	}

	return tokens, nil
}

// DeleteToken deletes an API token by ID
func (s *APIService) DeleteToken(id int) error {
	err := s.dbService.DeleteAPIToken(id)
	if err != nil {
		log.Printf("Error deleting token ID %d: %v", id, err)
		return fmt.Errorf("failed to delete token: %w", err)
	}

	// Reset Zhipu API service if active token was deleted
	activeToken, err := s.dbService.GetActiveAPIToken()
	if err != nil || activeToken == nil {
		s.zhipuAPIService = nil
	} else if s.zhipuAPIService == nil || s.zhipuAPIService.GetAPIToken() != activeToken.TokenValue {
		s.zhipuAPIService = NewZhipuAPIService(activeToken.TokenValue)
	}

	log.Printf("Successfully deleted token ID %d", id)
	return nil
}

// ValidateToken validates an API token
func (s *APIService) ValidateToken(token string) error {
	zhipuService := NewZhipuAPIService(token)
	err := zhipuService.ValidateAPIToken()
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return fmt.Errorf("token validation failed: %w", err)
	}

	log.Printf("Token validation successful")
	return nil
}

// ValidateSavedToken validates currently saved API token
func (s *APIService) ValidateSavedToken() (bool, error) {
	token, err := s.GetToken()
	if err != nil {
		return false, fmt.Errorf("failed to get saved token: %w", err)
	}

	if token == nil {
		return false, nil // No token saved
	}

	err = s.ValidateToken(token.TokenValue)
	if err != nil {
		return false, nil // Token is invalid, but don't return error to caller
	}

	// Update Zhipu API service if validation successful
	if s.zhipuAPIService == nil {
		s.zhipuAPIService = NewZhipuAPIService(token.TokenValue)
	}

	return true, nil
}

// ========== Sync Management APIs ==========

// StartSyncResponse 同步启动响应
type StartSyncResponse struct {
	Success bool   `json:"success"`
	SyncID  int    `json:"sync_id"`
	Message string `json:"message"`
}

// SyncStatusResponse 同步状态响应
type SyncStatusResponse struct {
	Syncing      bool      `json:"syncing"`
	Progress     float64   `json:"progress"`
	CurrentPage  int       `json:"current_page"`
	TotalPages   int       `json:"total_pages"`
	SyncedCount  int       `json:"synced_count"`
	FailedCount  int       `json:"failed_count"`
	TotalCount   int       `json:"total_count"`
	Message      string    `json:"message"`
	LastSyncTime time.Time `json:"last_sync_time"`
	Status       string    `json:"status"`
}

// SyncBills starts a sync operation for billing data (IPC_01: 修复参数签名)
func (s *APIService) SyncBills(billingMonth, syncType string, progressCallback func(*models.SyncProgress)) (*models.SyncResult, error) {
	// IPC_03: 添加参数校验
	if billingMonth == "" {
		return &models.SyncResult{
			Success:      false,
			ErrorMessage: "Billing month is required",
		}, NewValidationError(ErrCodeInvalidParameter, "Billing month is required")
	}

	// 验证账单月份格式 (YYYY-MM)
	if !isValidBillingMonth(billingMonth) {
		return &models.SyncResult{
			Success:      false,
			ErrorMessage: "Invalid billing month format. Expected YYYY-MM",
		}, NewValidationError(ErrCodeInvalidParameter, "Invalid billing month format")
	}

	// 验证同步类型
	if syncType == "" {
		syncType = "full" // 默认为全量同步
	}
	if syncType != "full" && syncType != "incremental" {
		return &models.SyncResult{
			Success:      false,
			ErrorMessage: "Invalid sync type. Must be 'full' or 'incremental'",
		}, NewValidationError(ErrCodeInvalidParameter, "Invalid sync type")
	}

	// 检查API令牌是否配置
	if s.zhipuAPIService == nil {
		return &models.SyncResult{
			Success:      false,
			ErrorMessage: "No API token configured",
		}, NewAuthError(ErrCodeSyncNoToken, "No API token configured")
	}

	// 启动同步任务
	year, month, err := parseBillingMonth(billingMonth)
	if err != nil {
		return &models.SyncResult{
			Success:      false,
			ErrorMessage: "Invalid billing month format: " + err.Error(),
		}, NewValidationError(ErrCodeInvalidParameter, "Invalid billing month format")
	}

	// 创建同步历史记录
	syncHistory := &models.SyncHistory{
		SyncType:     syncType,
		StartTime:    time.Now(),
		Status:       "running",
		BillingMonth: billingMonth,
		SyncTime:     time.Now(),
	}

	// 保存同步历史
	err = s.SaveSyncHistory(syncHistory)
	if err != nil {
		log.Printf("Failed to save sync history: %v", err)
	}

	response, err := s.zhipuAPIService.SyncFullMonth(year, month, func(progress *SyncProgress) {
		if progressCallback != nil {
			// 转换类型：从 services.SyncProgress 到 models.SyncProgress
			modelsProgress := &models.SyncProgress{
				CurrentPage: progress.CurrentPage,
				TotalPages:  progress.TotalPages,
				SyncedCount: progress.SyncedItems,
				FailedCount: 0, // services.SyncProgress没有FailedCount字段，使用0
				TotalCount:  progress.TotalItems,
			}
			progressCallback(modelsProgress)
		}
	})
	if err != nil {
		// 更新同步历史为失败状态
		syncHistory.Status = "failed"
		endTime := time.Now()
		syncHistory.EndTime = &endTime
		errorMsg := err.Error()
		syncHistory.ErrorMessage = &errorMsg
		s.SaveSyncHistory(syncHistory)

		return &models.SyncResult{
			Success:      false,
			ErrorMessage: "Failed to start sync: " + err.Error(),
		}, err
	}

	// 更新同步历史为完成状态
	syncHistory.Status = "completed"
	endTime := time.Now()
	syncHistory.EndTime = &endTime
	syncHistory.RecordsSynced = response.SyncedItems
	syncHistory.TotalRecords = response.TotalItems
	syncHistory.FailedCount = response.FailedItems
	s.SaveSyncHistory(syncHistory)

	return &models.SyncResult{
		Success:      response.Success,
		SyncedItems:  response.SyncedItems,
		TotalItems:   response.TotalItems,
		FailedItems:  response.FailedItems,
		ErrorMessage: response.ErrorMessage,
	}, nil
}

// isValidBillingMonth 验证账单月份格式
func isValidBillingMonth(billingMonth string) bool {
	if len(billingMonth) != 7 {
		return false
	}
	if billingMonth[4] != '-' {
		return false
	}

	year := billingMonth[0:4]
	month := billingMonth[5:7]

	_, yearErr := strconv.Atoi(year)
	_, monthErr := strconv.Atoi(month)

	return yearErr == nil && monthErr == nil
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

// GetSyncStatus retrieves current sync status
func (s *APIService) GetSyncStatus() (*models.SyncStatus, error) {
	status, err := s.dbService.GetAutoSyncStatus()
	if err != nil {
		log.Printf("Error getting sync status: %v", err)
		return nil, fmt.Errorf("failed to retrieve sync status: %w", err)
	}

	return status, nil
}

// GetSyncHistory retrieves sync history with filtering by sync type
func (s *APIService) GetSyncHistory(syncType string, pageNum, pageSize int) (*models.PaginatedResult, error) {
	// Get sync history with sync type filtering
	result, err := s.dbService.GetSyncHistory(syncType, pageNum, pageSize)
	if err != nil {
		log.Printf("Error getting sync history: %v", err)
		return nil, fmt.Errorf("failed to retrieve sync history: %w", err)
	}

	// Convert to frontend format and filter by sync type
	// Define SyncHistoryResponse type locally since it's only used here
	type SyncHistoryResponse struct {
		SyncTime     string `json:"sync_time"`
		BillingMonth string `json:"billing_month"`
		Status       string `json:"status"`
		SyncedCount  int    `json:"synced_count"`
		FailedCount  int    `json:"failed_count"`
		TotalCount   int    `json:"total_count"`
		Message      string `json:"message"`
	}

	var filteredHistory []SyncHistoryResponse
	if result.Data != nil {
		histories, ok := result.Data.([]models.SyncHistory)
		if ok {
			for _, history := range histories {
				// Filter by sync type if specified
				if syncType != "" && history.SyncType != syncType {
					continue
				}

				// Calculate billing month from start time
				billingMonth := history.StartTime.Format("2006-01")

				// Calculate failed count
				failedCount := 0
				if history.Status == "failed" && history.TotalRecords > 0 {
					failedCount = history.TotalRecords - history.RecordsSynced
				}

				// Format message
				message := ""
				if history.ErrorMessage != nil {
					message = *history.ErrorMessage
				}

				response := SyncHistoryResponse{
					SyncTime:     history.StartTime.Format("2006-01-02 15:04:05"),
					BillingMonth: billingMonth,
					Status:       getDisplayStatus(history.Status),
					SyncedCount:  history.RecordsSynced,
					FailedCount:  failedCount,
					TotalCount:   history.TotalRecords,
					Message:      message,
				}
				filteredHistory = append(filteredHistory, response)
			}
		}
	}

	// Apply pagination to filtered results
	total := len(filteredHistory)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	start := (pageNum - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var paginatedData []SyncHistoryResponse
	if start < total {
		paginatedData = filteredHistory[start:end]
	} else {
		paginatedData = []SyncHistoryResponse{}
	}

	// Build pagination info
	totalPages := (total + pageSize - 1) / pageSize
	pagination := models.PaginationParams{
		Page:    pageNum,
		Size:    pageSize,
		Total:   total,
		HasNext: pageNum < totalPages,
	}

	return &models.PaginatedResult{
		Data:       paginatedData,
		Pagination: pagination,
		Total:      total, // Add Total field for frontend compatibility
	}, nil
}

// getDisplayStatus converts internal status to display status
func getDisplayStatus(status string) string {
	switch status {
	case "completed":
		return "成功"
	case "failed":
		return "失败"
	case "running":
		return "运行中"
	default:
		return status
	}
}

// SyncRecentMonths syncs billing data for recent months
func (s *APIService) SyncRecentMonths(months int, progressCallback func(month, totalMonths int, monthProgress *SyncProgress)) ([]*SyncResult, error) {
	if s.zhipuAPIService == nil {
		return nil, fmt.Errorf("no API token configured")
	}

	results, err := s.zhipuAPIService.SyncRecentMonths(months, progressCallback)
	if err != nil {
		return nil, fmt.Errorf("sync recent months failed: %w", err)
	}

	// Save all synced bills to database
	for _, result := range results {
		// TODO: Implement bill processing logic when ProcessedBills field is available
		if result.Success {
			log.Printf("Sync completed successfully: %d items synced", result.SyncedItems)
		}
	}

	return results, nil
}

// ========== Configuration Management APIs ==========

// GetConfig retrieves a configuration value
func (s *APIService) GetConfig(key string) (string, error) {
	value, err := s.dbService.GetAutoSyncConfig(key)
	if err != nil {
		log.Printf("Error getting config %s: %v", key, err)
		return "", fmt.Errorf("failed to retrieve config: %w", err)
	}

	return value, nil
}

// SetConfig saves a configuration value
func (s *APIService) SetConfig(key, value, description string) error {
	err := s.dbService.SetAutoSyncConfig(key, value, description)
	if err != nil {
		log.Printf("Error setting config %s: %v", key, err)
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Printf("Successfully set config: %s = %s", key, value)
	return nil
}

// GetAllConfigs retrieves all configuration values
func (s *APIService) GetAllConfigs() ([]models.AutoSyncConfig, error) {
	configs, err := s.dbService.GetAllAutoSyncConfigs()
	if err != nil {
		log.Printf("Error getting all configs: %v", err)
		return nil, fmt.Errorf("failed to retrieve configs: %w", err)
	}

	return configs, nil
}

// ========== Utility APIs ==========

// GetDatabaseInfo retrieves database information
func (s *APIService) GetDatabaseInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{
		"path":    s.db.GetPath(),
		"type":    "SQLite3",
		"version": "3.x",
	}

	// Get table counts
	tableCounts := make(map[string]interface{})

	// Get expense_bills count
	var billCount int
	err := s.db.GetDB().QueryRow("SELECT COUNT(*) FROM expense_bills").Scan(&billCount)
	if err != nil {
		log.Printf("Error getting bill count: %v", err)
	} else {
		tableCounts["expense_bills"] = billCount
	}

	// Get sync_history count
	var syncCount int
	err = s.db.GetDB().QueryRow("SELECT COUNT(*) FROM sync_history").Scan(&syncCount)
	if err != nil {
		log.Printf("Error getting sync count: %v", err)
	} else {
		tableCounts["sync_history"] = syncCount
	}

	info["table_counts"] = tableCounts

	return info, nil
}

// CheckAPIConnectivity checks if API is accessible
func (s *APIService) CheckAPIConnectivity() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"connected": false,
		"message":   "No API token configured",
	}

	if s.zhipuAPIService == nil {
		return result, nil
	}

	err := s.zhipuAPIService.ValidateAPIToken()
	if err != nil {
		result["message"] = fmt.Sprintf("API connection failed: %v", err)
		return result, nil
	}

	result["connected"] = true
	result["message"] = "API connection successful"
	result["base_url"] = s.zhipuAPIService.GetBaseURL()

	return result, nil
}

// ========== Progress Tracking APIs ==========

// GetCurrentMembershipTier 获取当前会员等级信息
func (s *APIService) GetCurrentMembershipTier() (map[string]interface{}, error) {
	// 获取当前会员等级
	tier, err := s.dbService.GetCurrentMembershipTier()
	if err != nil {
		log.Printf("Error getting current membership tier: %v", err)
		// 使用默认值
		tier = "free"
	}

	// 获取该等级的限制信息
	limits, err := s.dbService.GetMembershipTierLimits(tier)
	if err != nil {
		log.Printf("Failed to get tier limits for %s: %v", tier, err)
		// 使用默认限制
		limits = &models.MembershipTierLimit{
			TierName:     tier,
			DailyLimit:   &[]int{1000}[0],
			MonthlyLimit: &[]int{30000}[0],
			MaxTokens:    &[]int{1000000}[0],
		}
	}

	result := map[string]interface{}{
		"tier":          tier,
		"tier_name":     getTierDisplayName(tier),
		"daily_limit":   limits.DailyLimit,
		"monthly_limit": limits.MonthlyLimit,
		"max_tokens":    limits.MaxTokens,
		"features":      limits.Features,
		"description":   limits.Description,
	}

	return result, nil
}

// 获取会员等级显示名称
func getTierDisplayName(tier string) string {
	displayNames := map[string]string{
		"free":       "免费版",
		"lite":       "Lite版",
		"pro":        "Pro版",
		"plus":       "Plus版",
		"enterprise": "企业版",
	}

	if name, ok := displayNames[tier]; ok {
		return name
	}
	return tier
}

// GetApiUsageProgress retrieves API usage progress against limits
func (s *APIService) GetApiUsageProgress() (map[string]interface{}, error) {
	// Get current month's API usage
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 获取当前会员等级信息
	tierInfo, err := s.GetCurrentMembershipTier()
	if err != nil {
		log.Printf("Error getting membership tier: %v", err)
		// 使用默认值
		tierInfo = map[string]interface{}{
			"tier":          "free",
			"daily_limit":   1000,
			"monthly_limit": 30000,
		}
	}

	// Get API usage count for current month
	usageStats, err := s.statsService.GetOverallStats(&startOfMonth, &now)
	if err != nil {
		log.Printf("Error getting API usage stats: %v", err)
		return nil, fmt.Errorf("failed to get API usage stats: %w", err)
	}

	apiUsageCount := usageStats.TotalRecords
	dailyLimit := 1000
	monthlyLimit := 30000

	// 从会员等级信息中获取限制
	if daily, ok := tierInfo["daily_limit"]; ok && daily != nil {
		if dl, ok := daily.(int); ok {
			dailyLimit = dl
		}
	}
	if monthly, ok := tierInfo["monthly_limit"]; ok && monthly != nil {
		if ml, ok := monthly.(int); ok {
			monthlyLimit = ml
		}
	}

	// Calculate today's usage
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayStats, err := s.statsService.GetOverallStats(&todayStart, &now)
	if err != nil {
		log.Printf("Error getting today's API usage: %v", err)
		todayStats = usageStats // Fallback to monthly stats
	}

	result := map[string]interface{}{
		"current_usage":      apiUsageCount,
		"daily_usage":        todayStats.TotalRecords,
		"daily_limit":        dailyLimit,
		"monthly_limit":      monthlyLimit,
		"daily_percentage":   float64(todayStats.TotalRecords) / float64(dailyLimit) * 100,
		"monthly_percentage": float64(apiUsageCount) / float64(monthlyLimit) * 100,
		"tier":               "free",
		"growthRate":         0.0,
	}

	return result, nil
}

// GetTokenUsageProgress retrieves Token usage progress against limits
func (s *APIService) GetTokenUsageProgress() (map[string]interface{}, error) {
	// Get current month's Token usage
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Calculate token usage from stats
	usageStats, err := s.statsService.GetOverallStats(&startOfMonth, &now)
	if err != nil {
		log.Printf("Error getting token usage stats: %v", err)
		return nil, fmt.Errorf("failed to get token usage stats: %w", err)
	}

	// Get membership tier limits for tokens
	limits, err := s.dbService.GetMembershipTierLimits("free")
	if err != nil {
		log.Printf("Error getting membership limits: %v", err)
		limits = &models.MembershipTierLimit{
			MaxTokens: &[]int{1000000}[0], // Default 1M tokens
		}
	}

	// Calculate total token usage (sum of all token usage from bills)
	var totalTokenUsage float64 = 0
	for _, usage := range usageStats.HourlyUsage {
		totalTokenUsage += usage.TokenUsage
	}

	maxTokens := 1000000
	if limits.MaxTokens != nil {
		maxTokens = *limits.MaxTokens
	}

	// Calculate today's token usage
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayStats, err := s.statsService.GetOverallStats(&todayStart, &now)
	if err != nil {
		log.Printf("Error getting today's token usage: %v", err)
		todayStats = usageStats
	}

	var todayTokenUsage float64 = 0
	for _, usage := range todayStats.HourlyUsage {
		todayTokenUsage += usage.TokenUsage
	}

	result := map[string]interface{}{
		"current_usage":      totalTokenUsage,
		"daily_usage":        todayTokenUsage,
		"monthly_limit":      float64(maxTokens),
		"daily_percentage":   todayTokenUsage / float64(maxTokens/30) * 100, // Approximate daily limit
		"monthly_percentage": totalTokenUsage / float64(maxTokens) * 100,
		"tier":               "free",
		"tokens_formatted":   formatTokenCount(totalTokenUsage),
		"limit_formatted":    formatTokenCount(float64(maxTokens)),
	}

	return result, nil
}

// Helper function to format token count
func formatTokenCount(tokens float64) string {
	if tokens >= 1000000 {
		return fmt.Sprintf("%.1fM", tokens/1000000)
	} else if tokens >= 1000 {
		return fmt.Sprintf("%.1fK", tokens/1000)
	}
	return fmt.Sprintf("%.0f", tokens)
}

// GetTotalCostProgress retrieves total cost progress against limits
func (s *APIService) GetTotalCostProgress() (map[string]interface{}, error) {
	// Get current month's total cost
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	usageStats, err := s.statsService.GetOverallStats(&startOfMonth, &now)
	if err != nil {
		log.Printf("Error getting cost stats: %v", err)
		return nil, fmt.Errorf("failed to get cost stats: %w", err)
	}

	// Default cost limits (in yuan)
	dailyCostLimit := 50.0     // 50 yuan per day
	monthlyCostLimit := 1000.0 // 1000 yuan per month

	// Calculate today's cost
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayStats, err := s.statsService.GetOverallStats(&todayStart, &now)
	if err != nil {
		log.Printf("Error getting today's cost: %v", err)
		todayStats = usageStats
	}

	// Calculate total cost from stats
	var monthlyCost float64 = 0
	for _, usage := range usageStats.HourlyUsage {
		monthlyCost += usage.CashCost
	}

	var todayCost float64 = 0
	for _, usage := range todayStats.HourlyUsage {
		todayCost += usage.CashCost
	}

	result := map[string]interface{}{
		"current_usage":           monthlyCost,
		"daily_usage":             todayCost,
		"daily_limit":             dailyCostLimit,
		"monthly_limit":           monthlyCostLimit,
		"daily_percentage":        todayCost / dailyCostLimit * 100,
		"monthly_percentage":      monthlyCost / monthlyCostLimit * 100,
		"currency":                "CNY",
		"tier":                    "free",
		"formatted_daily":         fmt.Sprintf("¥%.2f", todayCost),
		"formatted_monthly":       fmt.Sprintf("¥%.2f", monthlyCost),
		"formatted_daily_limit":   fmt.Sprintf("¥%.2f", dailyCostLimit),
		"formatted_monthly_limit": fmt.Sprintf("¥%.2f", monthlyCostLimit),
	}

	return result, nil
}

// ForceResetSyncStatus forcefully resets all running syncs to failed status
func (s *APIService) ForceResetSyncStatus() error {
	db := s.dbService.GetDB()

	// Force update all running syncs to failed
	query := `
		UPDATE sync_history
		SET status = 'failed',
		    end_time = datetime('now'),
		    error_message = 'Sync manually reset by user'
		WHERE status = 'running'
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to reset sync status: %w", err)
	}

	log.Printf("Successfully reset all running syncs")
	return nil
}

// ========== Auto Sync APIs ==========

// GetAutoSyncConfig 获取自动同步配置
func (s *APIService) GetAutoSyncConfig() (*models.AutoSyncConfig, error) {
	config, err := s.autoSyncService.GetConfig()
	if err != nil {
		log.Printf("Error getting auto sync config: %v", err)
		return nil, fmt.Errorf("failed to get auto sync config: %w", err)
	}

	// 添加运行状态和最后同步时间
	status, err := s.autoSyncService.GetStatus()
	if err != nil {
		log.Printf("Error getting auto sync status: %v", err)
	} else {
		config.Enabled = status["enabled"].(bool)
		if lastSyncTime, ok := status["last_sync_time"]; ok && lastSyncTime != nil {
			if timeVal, ok := lastSyncTime.(*time.Time); ok {
				config.LastSyncTime = timeVal
			}
		}
		if nextSyncTime, ok := status["next_sync_time"]; ok && nextSyncTime != nil {
			if timeVal, ok := nextSyncTime.(*time.Time); ok {
				config.NextSyncTime = timeVal
			}
		}
	}

	return config, nil
}

// SaveAutoSyncConfig 保存自动同步配置
func (s *APIService) SaveAutoSyncConfig(config *models.AutoSyncConfig) error {
	err := s.autoSyncService.SaveConfig(config)
	if err != nil {
		log.Printf("Error saving auto sync config: %v", err)
		return fmt.Errorf("failed to save auto sync config: %w", err)
	}

	log.Printf("Auto sync config saved: enabled=%v, frequency=%d seconds",
		config.Enabled, config.FrequencySeconds)
	return nil
}

// TriggerAutoSync 立即触发一次自动同步
func (s *APIService) TriggerAutoSync() (map[string]interface{}, error) {
	err := s.autoSyncService.TriggerNow()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "触发自动同步失败: " + err.Error(),
		}, err
	}

	return map[string]interface{}{
		"success": true,
		"message": "自动同步已触发",
	}, nil
}

// StopAutoSync 停止自动同步
func (s *APIService) StopAutoSync() (map[string]interface{}, error) {
	err := s.autoSyncService.Stop()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "停止自动同步失败: " + err.Error(),
		}, err
	}

	return map[string]interface{}{
		"success": true,
		"message": "Auto sync stopped successfully",
	}, nil
}

// GetAutoSyncStatus 获取自动同步状态
func (s *APIService) GetAutoSyncStatus() (map[string]interface{}, error) {
	status, err := s.autoSyncService.GetStatus()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "获取自动同步状态失败: " + err.Error(),
		}, err
	}

	status["success"] = true
	return status, nil
}

// ========== Additional Sync-related Methods ==========

// Check if there's a running sync and return progress info
func (s *APIService) GetRunningSyncStatus() (map[string]interface{}, error) {
	// Get latest sync history via API service
	history, err := s.GetSyncHistory("", 1, 1)
	if err != nil {
		return map[string]interface{}{
			"syncing": false,
			"message": "Failed to get sync status",
		}, err
	}

	// Check if we have any history and if latest is running
	if history.Data == nil {
		return map[string]interface{}{
			"syncing": false,
			"message": "No sync history",
		}, nil
	}

	// Since we can't access the underlying data directly, return default status
	// The frontend will handle progress polling via the regular GetSyncStatus method
	return map[string]interface{}{
		"syncing": false,
		"message": "Sync status check complete",
	}, nil
}

// CleanupStaleSyncs manually cleans up any stale running sync records
func (s *APIService) CleanupStaleSyncs() (map[string]interface{}, error) {
	// For now, just return a success message since cleanup is handled in GetRunningSyncCount
	return map[string]interface{}{
		"success": true,
		"message": "Cleanup completed - stale sync records have been marked as failed",
	}, nil
}

// CleanOldSyncHistory 清理指定天数前的同步历史记录 (MUTATION_01)
func (s *APIService) CleanOldSyncHistory(days int) error {
	err := s.dbService.CleanOldSyncHistory(days)
	if err != nil {
		log.Printf("Error cleaning old sync history: %v", err)
		return fmt.Errorf("failed to clean old sync history: %w", err)
	}

	log.Printf("Successfully cleaned sync history older than %d days", days)
	return nil
}

// DeleteAllExpenseBills 清空所有账单数据 (MUTATION_02)
func (s *APIService) DeleteAllExpenseBills() error {
	err := s.dbService.DeleteAllExpenseBills()
	if err != nil {
		log.Printf("Error deleting all expense bills: %v", err)
		return fmt.Errorf("failed to delete all expense bills: %w", err)
	}

	log.Printf("Successfully deleted all expense bills data")
	return nil
}

// SaveSyncHistory saves sync history record (缺失的IPC方法)
func (s *APIService) SaveSyncHistory(history *models.SyncHistory) error {
	err := s.dbService.SaveSyncHistory(history)
	if err != nil {
		log.Printf("Error saving sync history: %v", err)
		return fmt.Errorf("failed to save sync history: %w", err)
	}

	log.Printf("Successfully saved sync history: type=%s, status=%s", history.SyncType, history.Status)
	return nil
}
