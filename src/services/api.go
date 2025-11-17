package services

import (
	"fmt"
	"glm-usage-monitor/models"
	"log"
	"time"
)

// APIService provides all API methods for the frontend
type APIService struct {
	dbService        *DatabaseService
	statsService     *StatisticsService
	zhipuAPIService  *ZhipuAPIService
	db               DatabaseInterface
}

// NewAPIService creates a new API service
func NewAPIService(db DatabaseInterface) *APIService {
	dbService := NewDatabaseService(db.GetDB())
	statsService := NewStatisticsService(db.GetDB())

	return &APIService{
		dbService:       dbService,
		statsService:    statsService,
		zhipuAPIService: nil, // Will be initialized when token is set
		db:              db,
	}
}

// ========== Bill Management APIs ==========

// GetBills retrieves expense bills with filtering and pagination
func (s *APIService) GetBills(filter *models.BillFilter) (*models.PaginatedResult, error) {
	if filter == nil {
		filter = &models.BillFilter{
			PageNum:  1,
			PageSize: 20,
		}
	}

	result, err := s.dbService.GetExpenseBills(filter)
	if err != nil {
		log.Printf("Error getting bills: %v", err)
		return nil, fmt.Errorf("failed to retrieve bills: %w", err)
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

// GetSyncStatus retrieves current sync status
func (s *APIService) GetSyncStatus() (*models.SyncStatus, error) {
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
	// Get all sync history first
	result, err := s.dbService.GetSyncHistory(1, 1000) // Get all records for filtering
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
	// Check if we have an API token
	if s.zhipuAPIService == nil {
		return &SyncResult{
			Success:      false,
			ErrorMessage: "No API token configured",
		}, fmt.Errorf("no API token configured")
	}

	// Check if there's already a running sync
	runningCount, err := s.dbService.GetRunningSyncCount()
	if err != nil {
		return nil, fmt.Errorf("failed to check running syncs: %w", err)
	}
	if runningCount > 0 {
		return &SyncResult{
			Success:      false,
			ErrorMessage: "Another sync operation is already in progress",
		}, fmt.Errorf("sync already in progress")
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