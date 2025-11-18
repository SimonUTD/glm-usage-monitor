package services

import (
	"fmt"
	"glm-usage-monitor/models"
	"log"
	"strconv"
	"strings"
	"time"
)

// APIService provides all API methods for the frontend
type APIService struct {
	dbService        *DatabaseService
	statsService     *StatisticsService
	zhipuAPIService  *ZhipuAPIService
	autoSyncService  *AutoSyncService
	db               DatabaseInterface
	errorHandler     ErrorHandler
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

// GetBills retrieves expense bills with filtering and pagination
func (s *APIService) GetBills(filter *models.BillFilter) (*models.PaginatedResult, error) {
	// 参数验证
	if filter == nil {
		filter = &models.BillFilter{
			PageNum:  1,
			PageSize: 20,
		}
	}

	var result *models.PaginatedResult

	// 使用错误处理机制执行数据库操作
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
func (s *APIService) GetBillByID(id int) (*models.ExpenseBill, error) {
	bill, err := s.dbService.GetExpenseBillByID(id)
	if err != nil {
		log.Printf("Error getting bill by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve bill: %w", err)
	}

	return bill, nil
}

// DeleteBill deletes an expense bill by ID
func (s *APIService) DeleteBill(id int) error {
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

// GetStats retrieves overall usage statistics
func (s *APIService) GetStats(startDate, endDate *time.Time) (*models.StatsResponse, error) {
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

// SaveToken saves an API token
func (s *APIService) SaveToken(tokenName, tokenValue string) error {
	token := &models.APIToken{
		TokenName:  tokenName,
		TokenValue: tokenValue,
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

// GetToken retrieves the active API token
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

	// Reset Zhipu API service if the active token was deleted
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

// ValidateSavedToken validates the currently saved API token
func (s *APIService) ValidateSavedToken() (bool, error) {
	token, err := s.GetToken()
	if err != nil {
		return false, fmt.Errorf("failed to get saved token: %w", err)
	}

	if token == nil || token.TokenValue == "" {
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

// StartSync 启动异步同步任务
func (s *APIService) StartSync(billingMonth string) (*StartSyncResponse, error) {
	// 1. 检查是否有正在运行的同步
	runningCount, err := s.dbService.GetRunningSyncCount()
	if err != nil {
		return nil, fmt.Errorf("failed to check running syncs: %w", err)
	}
	if runningCount > 0 {
		return &StartSyncResponse{
			Success: false,
			Message: "已有同步任务正在运行，请稍后再试",
		}, nil
	}

	// 2. 解析账单月份
	year, month, err := parseBillingMonth(billingMonth)
	if err != nil {
		return nil, fmt.Errorf("invalid billing month format: %w", err)
	}

	// 3. 检查API Token
	if s.zhipuAPIService == nil {
		return &StartSyncResponse{
			Success: false,
			Message: "API Token 未配置",
		}, nil
	}

	// 4. 创建同步历史记录
	syncHistory := &models.SyncHistory{
		SyncType:     "full",
		BillingMonth: billingMonth,
		StartTime:    time.Now(),
		Status:       "running",
		PageSynced:   0,
		TotalPages:   0,
	}

	err = s.dbService.CreateSyncHistory(syncHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync history: %w", err)
	}

	// 5. 启动异步同步goroutine
	go s.performAsyncSync(syncHistory.ID, year, month, billingMonth)

	// 6. 立即返回任务信息
	return &StartSyncResponse{
		Success: true,
		SyncID:  syncHistory.ID,
		Message: "同步任务已启动",
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

// GetSyncStatus 获取同步状态
func (s *APIService) GetSyncStatus() (*SyncStatusResponse, error) {
	// 获取最新的同步历史
	latestSync, err := s.dbService.GetLatestSyncHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sync history: %w", err)
	}

	if latestSync == nil {
		return &SyncStatusResponse{
			Syncing:  false,
			Progress: 0,
			Message:  "暂无同步记录",
			Status:   "idle",
		}, nil
	}

	// 计算进度
	var progress float64 = 0
	if latestSync.Status == "running" && latestSync.TotalPages > 0 {
		progress = float64(latestSync.PageSynced) / float64(latestSync.TotalPages) * 100
	} else if latestSync.Status == "completed" {
		progress = 100
	}

	// 生成状态消息
	message := s.getSyncStatusMessage(latestSync)

	return &SyncStatusResponse{
		Syncing:      latestSync.Status == "running",
		Progress:     progress,
		CurrentPage:  latestSync.PageSynced,
		TotalPages:   latestSync.TotalPages,
		SyncedCount:  latestSync.RecordsSynced,
		FailedCount:  latestSync.FailedCount,
		TotalCount:   latestSync.TotalRecords,
		Message:      message,
		LastSyncTime: latestSync.StartTime,
		Status:       latestSync.Status,
	}, nil
}

// getSyncStatusMessage 生成状态消息
func (s *APIService) getSyncStatusMessage(sync *models.SyncHistory) string {
	switch sync.Status {
	case "running":
		if sync.TotalPages > 0 {
			return fmt.Sprintf("正在同步第 %d/%d 页...", sync.PageSynced, sync.TotalPages)
		}
		return "正在准备同步..."
	case "completed":
		return fmt.Sprintf("同步完成: 成功%d条, 失败%d条", sync.RecordsSynced, sync.FailedCount)
	case "failed":
		if sync.ErrorMessage != nil && *sync.ErrorMessage != "" {
			return fmt.Sprintf("同步失败: %s", *sync.ErrorMessage)
		}
		return "同步失败"
	default:
		return "未知状态"
	}
}

// GetSyncStatus retrieves current sync status (保持兼容性)
func (s *APIService) GetSyncStatusLegacy() (*models.SyncStatus, error) {
	status, err := s.dbService.GetAutoSyncStatus()
	if err != nil {
		log.Printf("Error getting sync status: %v", err)
		return nil, fmt.Errorf("failed to retrieve sync status: %w", err)
	}

	return status, nil
}

// SyncHistoryResponse represents the format expected by frontend
type SyncHistoryResponse struct {
	SyncTime     string `json:"sync_time"`
	BillingMonth string `json:"billing_month"`
	Status       string `json:"status"`
	SyncedCount  int    `json:"synced_count"`
	FailedCount  int    `json:"failed_count"`
	TotalCount   int    `json:"total_count"`
	Message      string `json:"message"`
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

// SyncBills starts a sync operation for billing data
func (s *APIService) SyncBills(year, month int, progressCallback func(*SyncProgress)) (*SyncResult, error) {
	// 验证输入参数
	if year < 2020 || year > 2030 || month < 1 || month > 12 {
		return &SyncResult{
			Success:      false,
			ErrorMessage: "Invalid year or month parameter",
		}, NewValidationError(ErrCodeInvalidParameter, "Invalid year or month parameter")
	}

	// 检查API令牌是否配置
	if s.zhipuAPIService == nil {
		err := NewAuthError(ErrCodeSyncNoToken, "No API token configured")
		context := map[string]interface{}{
			"operation": "SyncBills",
			"year":      year,
			"month":     month,
		}
		s.errorHandler.HandleError(err, context)

		return &SyncResult{
			Success:      false,
			ErrorMessage: GetErrorMessage(err),
		}, err
	}

	// 检查是否已有同步任务在运行
	var runningCount int
	err := SafeExecute(func() error {
		var operationErr error
		runningCount, operationErr = s.dbService.GetRunningSyncCount()
		return operationErr
	})

	if err != nil {
		context := map[string]interface{}{
			"operation": "CheckRunningSync",
			"year":      year,
			"month":     month,
		}
		s.errorHandler.HandleError(err, context)

		dbErr := WrapError(err, ErrorTypeDatabase, ErrCodeDBQueryFailed, "Failed to check running syncs")
		return nil, dbErr
	}

	if runningCount > 0 {
		err := NewSyncError(ErrCodeSyncAlreadyRunning, "Another sync operation is already in progress")
		context := map[string]interface{}{
			"operation":    "SyncBills",
			"runningCount": runningCount,
		}
		s.errorHandler.HandleError(err, context)

		return &SyncResult{
			Success:      false,
			ErrorMessage: GetErrorMessage(err),
		}, err
	}

	// Create sync history record
	syncHistory := &models.SyncHistory{
		SyncType:  "manual",
		StartTime: time.Now(),
		Status:    "running",
	}

	err = s.dbService.CreateSyncHistory(syncHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync history: %w", err)
	}

	// Perform the sync
	result, err := s.zhipuAPIService.SyncFullMonth(year, month, progressCallback)
	if err != nil {
		// Update sync history with failure
		errorMsg := err.Error()
		syncHistory.EndTime = &[]time.Time{time.Now()}[0]
		syncHistory.Status = "failed"
		syncHistory.ErrorMessage = &errorMsg
		s.dbService.UpdateSyncHistory(syncHistory.ID, syncHistory)

		return nil, fmt.Errorf("sync failed: %w", err)
	}

	// Save the synced bills to database if sync was successful
	if result.Success && len(result.ProcessedBills) > 0 {
		// Convert slice of structs to slice of pointers
		billPointers := make([]*models.ExpenseBill, len(result.ProcessedBills))
		for i := range result.ProcessedBills {
			billPointers[i] = &result.ProcessedBills[i]
		}
		err = s.dbService.BatchCreateExpenseBills(billPointers)
		if err != nil {
			log.Printf("Failed to save synced bills: %v", err)
			result.ErrorMessage = fmt.Sprintf("Sync completed but failed to save bills: %v", err)
			result.Success = false
		}
	}

	// Update sync history with completion
	endTime := time.Now()
	syncHistory.EndTime = &endTime
	syncHistory.Status = "completed"
	if result.Success {
		syncHistory.RecordsSynced = result.SyncedItems
		syncHistory.TotalRecords = result.TotalItems
	}

	err = s.dbService.UpdateSyncHistory(syncHistory.ID, syncHistory)
	if err != nil {
		log.Printf("Failed to update sync history: %v", err)
	}

	log.Printf("Sync completed: %d records processed, %d synced, %d failed",
		result.TotalItems, result.SyncedItems, result.FailedItems)

	return result, nil
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
		if result.Success && len(result.ProcessedBills) > 0 {
			// Convert slice of structs to slice of pointers
			billPointers := make([]*models.ExpenseBill, len(result.ProcessedBills))
			for i := range result.ProcessedBills {
				billPointers[i] = &result.ProcessedBills[i]
			}
			err := s.dbService.BatchCreateExpenseBills(billPointers)
			if err != nil {
				log.Printf("Failed to save bills for month: %v", err)
				result.ErrorMessage = fmt.Sprintf("Sync completed but failed to save bills: %v", err)
				result.Success = false
			}
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
		"path":     s.db.GetPath(),
		"type":     "SQLite3",
		"version":  "3.x",
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

// CheckAPIConnectivity checks if the API is accessible
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
		"current_usage": apiUsageCount,
		"daily_usage":   todayStats.TotalRecords,
		"daily_limit":   dailyLimit,
		"monthly_limit": monthlyLimit,
		"daily_percentage": float64(todayStats.TotalRecords) / float64(dailyLimit) * 100,
		"monthly_percentage": float64(apiUsageCount) / float64(monthlyLimit) * 100,
		"tier": "free",
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
		"current_usage": totalTokenUsage,
		"daily_usage":   todayTokenUsage,
		"monthly_limit": float64(maxTokens),
		"daily_percentage": todayTokenUsage / float64(maxTokens/30) * 100, // Approximate daily limit
		"monthly_percentage": totalTokenUsage / float64(maxTokens) * 100,
		"tier": "free",
		"tokens_formatted": formatTokenCount(totalTokenUsage),
		"limit_formatted": formatTokenCount(float64(maxTokens)),
	}

	return result, nil
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
	dailyCostLimit := 50.0   // 50 yuan per day
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
		"current_usage": monthlyCost,
		"daily_usage":   todayCost,
		"daily_limit":   dailyCostLimit,
		"monthly_limit": monthlyCostLimit,
		"daily_percentage": todayCost / dailyCostLimit * 100,
		"monthly_percentage": monthlyCost / monthlyCostLimit * 100,
		"currency": "CNY",
		"tier": "free",
		"formatted_daily": fmt.Sprintf("¥%.2f", todayCost),
		"formatted_monthly": fmt.Sprintf("¥%.2f", monthlyCost),
		"formatted_daily_limit": fmt.Sprintf("¥%.2f", dailyCostLimit),
		"formatted_monthly_limit": fmt.Sprintf("¥%.2f", monthlyCostLimit),
	}

	return result, nil
}

// ForceResetSyncStatus forces reset of sync status (for error recovery)
func (s *APIService) ForceResetSyncStatus() error {
	err := s.dbService.ResetRunningSyncs()
	if err != nil {
		log.Printf("Error resetting sync status: %v", err)
		return fmt.Errorf("failed to reset sync status: %w", err)
	}

	log.Printf("Successfully reset all running sync statuses")
	return nil
}

// ========== 异步同步核心方法 ==========

// performAsyncSync 异步执行同步任务
func (s *APIService) performAsyncSync(syncHistoryID int, year, month int, billingMonth string) {
	// 错误恢复
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Sync panic recovered: %v", r)
			s.updateSyncStatus(syncHistoryID, "failed", fmt.Sprintf("同步异常: %v", r))
		}
	}()

	// 初始化计数器
	var totalSynced = 0
	var totalFailed = 0
	var currentPage = 1
	var totalPages = 1
	var totalRecords = 0
	const pageSize = 20

	log.Printf("Starting async sync for %s (ID: %d)", billingMonth, syncHistoryID)

	// 分页循环获取数据
	for currentPage <= totalPages {
		log.Printf("Syncing page %d/%d", currentPage, totalPages)

		// 获取当前页数据
		billingResp, err := s.zhipuAPIService.GetExpenseBillsPage(year, month, currentPage, pageSize)
		if err != nil {
			errorMsg := fmt.Sprintf("获取第%d页数据失败: %v", currentPage, err)
			log.Printf("Sync error: %s", errorMsg)
			s.updateSyncStatus(syncHistoryID, "failed", errorMsg)
			return
		}

		// 第一次请求时更新总页数和总记录数
		if currentPage == 1 {
			totalPages = billingResp.Data.TotalPages
			totalRecords = billingResp.Data.Total
			log.Printf("Total pages: %d, Total records: %d", totalPages, totalRecords)
		}

		// 转换 BillItem 到 ExpenseBill
		var items []models.ExpenseBill
		for _, billItem := range billingResp.Data.BillList {
			// Convert to map for transformation
			billMap, err := s.zhipuAPIService.BillItemToMap(&billItem)
			if err != nil {
				log.Printf("Failed to convert bill item to map: %v", err)
				totalFailed++
				continue
			}

			// Transform to expense bill
			expenseBill, err := models.TransformExpenseBill(billMap)
			if err != nil {
				log.Printf("Failed to transform expense bill: %v", err)
				totalFailed++
				continue
			}

			// Validate the bill
			if err := models.ValidateExpenseBill(expenseBill); err != nil {
				log.Printf("Invalid expense bill: %v", err)
				totalFailed++
				continue
			}

			items = append(items, *expenseBill)
		}

		// 保存当前页数据
		savedCount, failedCount := s.saveBatchData(syncHistoryID, items)
		totalSynced += savedCount
		totalFailed += failedCount

		log.Printf("Page %d: saved %d, failed %d", currentPage, savedCount, failedCount)

		// 更新同步进度
		err = s.updateSyncProgress(syncHistoryID, currentPage, totalPages, totalSynced, totalFailed, totalRecords)
		if err != nil {
			log.Printf("Failed to update sync progress: %v", err)
		}

		// 检查是否完成
		if currentPage >= totalPages {
			break
		}

		currentPage++

		// 避免API限制
		time.Sleep(100 * time.Millisecond)
	}

	// 完成同步
	successMsg := fmt.Sprintf("同步完成: 成功%d条, 失败%d条", totalSynced, totalFailed)
	log.Printf("Sync completed: %s", successMsg)
	s.updateSyncStatus(syncHistoryID, "completed", successMsg)
}

// saveBatchData 保存批量数据（使用事务）
func (s *APIService) saveBatchData(syncHistoryID int, items []models.ExpenseBill) (savedCount, failedCount int) {
	if len(items) == 0 {
		return 0, 0
	}

	// 开始事务
	tx, err := s.dbService.BeginTx()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return 0, len(items)
	}

	// 确保事务被正确处理
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back due to error: %v", err)
		} else {
			err = tx.Commit()
			if err != nil {
				log.Printf("Failed to commit transaction: %v", err)
				savedCount = 0
				failedCount = len(items)
			}
		}
	}()

	// 批量插入数据
	for i := range items {
		err = s.dbService.CreateOrUpdateExpenseBillInTx(tx, &items[i])
		if err != nil {
			log.Printf("Failed to save bill %s: %v", items[i].BillingNo, err)
			failedCount++
		} else {
			savedCount++
		}
	}

	return savedCount, failedCount
}

// updateSyncProgress 更新同步进度
func (s *APIService) updateSyncProgress(syncHistoryID, pageSynced, totalPages, recordsSynced, failedCount, totalRecords int) error {
	db := s.dbService.GetDB()

	updateSQL := `
		UPDATE sync_history
		SET page_synced = ?,
		    total_pages = ?,
		    records_synced = ?,
		    total_records = ?,
		    failed_count = ?
		WHERE id = ?
	`

	_, err := db.Exec(updateSQL, pageSynced, totalPages, recordsSynced, totalRecords, failedCount, syncHistoryID)
	return err
}

// updateSyncStatus 更新同步状态
func (s *APIService) updateSyncStatus(syncHistoryID int, status, message string) error {
	db := s.dbService.GetDB()

	var updateSQL string
	var args []interface{}

	if status == "completed" || status == "failed" {
		updateSQL = `
			UPDATE sync_history
			SET status = ?,
			    error_message = ?,
			    end_time = datetime('now')
			WHERE id = ?
		`
		args = []interface{}{status, message, syncHistoryID}
	} else {
		updateSQL = `
			UPDATE sync_history
			SET status = ?,
			    error_message = ?
			WHERE id = ?
		`
		args = []interface{}{status, message, syncHistoryID}
	}

	_, err := db.Exec(updateSQL, args...)
	if err != nil {
		log.Printf("Failed to update sync status: %v", err)
	}
	return err
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
			config.LastSyncTime = lastSyncTime.(*time.Time).Format("2006-01-02 15:04:05")
		}
		if nextSyncTime, ok := status["next_sync_time"]; ok && nextSyncTime != nil {
			config.NextSyncTime = nextSyncTime.(string)
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
func (s *APIService) TriggerAutoSync() error {
	err := s.autoSyncService.TriggerNow()
	if err != nil {
		log.Printf("Error triggering auto sync: %v", err)
		return fmt.Errorf("failed to trigger auto sync: %w", err)
	}

	log.Println("Auto sync triggered successfully")
	return nil
}

// StopAutoSync 停止自动同步
func (s *APIService) StopAutoSync() error {
	err := s.autoSyncService.Stop()
	if err != nil {
		log.Printf("Error stopping auto sync: %v", err)
		return fmt.Errorf("failed to stop auto sync: %w", err)
	}

	log.Println("Auto sync stopped successfully")
	return nil
}

// GetAutoSyncStatus 获取自动同步状态
func (s *APIService) GetAutoSyncStatus() (map[string]interface{}, error) {
	status, err := s.autoSyncService.GetStatus()
	if err != nil {
		log.Printf("Error getting auto sync status: %v", err)
		return nil, fmt.Errorf("failed to get auto sync status: %w", err)
	}

	return status, nil
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