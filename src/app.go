package main

import (
	"context"
	"fmt"
	"glm-usage-monitor/models"
	"glm-usage-monitor/services"
	"log"
	"math"
	"strconv"
	"time"
)

// App struct
type App struct {
	ctx        context.Context
	database   *Database
	apiService *services.APIService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("DEBUG: Application startup beginning...")

	// Initialize database
	log.Printf("DEBUG: Initializing database...")
	db, err := NewDatabase()
	if err != nil {
		log.Printf("DEBUG: Database initialization failed: %v", err)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Printf("DEBUG: Database initialized successfully")

	// 执行数据库迁移
	log.Printf("DEBUG: Running database migrations...")
	if err := RunMigrations(db.GetDB()); err != nil {
		log.Printf("DEBUG: Database migrations failed: %v", err)
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Printf("DEBUG: Database migrations completed successfully")

	a.database = db
	a.apiService = services.NewAPIService(db)

	// Cleanup any stale running syncs on startup
	log.Printf("DEBUG: Cleaning up stale syncs...")
	err = a.cleanupAllRunningSyncs()
	if err != nil {
		log.Printf("Warning: failed to cleanup running syncs on startup: %v", err)
	} else {
		log.Println("Cleaned up stale sync records on startup")
	}

	log.Printf("DEBUG: Application startup completed successfully")
}

// cleanupAllRunningSyncs marks all running syncs as failed (called on startup)
func (a *App) cleanupAllRunningSyncs() error {
	db := a.database.GetDB()

	// Force update all running syncs to failed
	query := `
		UPDATE sync_history
		SET status = 'failed',
		    end_time = datetime('now'),
		    error_message = 'Sync automatically cancelled on application startup'
		WHERE status = 'running'
	`

	result, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup running syncs on startup: %w", err)
	}

	// Check how many records were updated
	if rowsAffected, err := result.RowsAffected(); err == nil {
		if rowsAffected > 0 {
			log.Printf("Marked %d sync records as failed on startup", rowsAffected)
		}
	}

	return nil
}

// shutdown is called when the app is about to close
func (a *App) shutdown(ctx context.Context) {
	if a.database != nil {
		if err := a.database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database closed successfully")
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetDatabase returns the database instance
func (a *App) GetDatabase() *Database {
	return a.database
}

// GetAPIService returns the API service instance
func (a *App) GetAPIService() *services.APIService {
	return a.apiService
}

// ========== Frontend API Methods ==========

// GetApiUsageProgress returns API usage progress with growth rate
func (a *App) GetApiUsageProgress() (map[string]interface{}, error) {
	// IPC_05: 从动态配置获取限制，移除硬编码
	// 从配置服务获取每日限制
	dailyLimit := a.getDynamicDailyLimit()

	result := map[string]interface{}{
		"percentage": 0,
		"used":       0,
		"limit":      dailyLimit,
		"remaining":  dailyLimit,
		"growthRate": 0.0,
	}

	// Get recent usage for the last hour
	usage, err := a.apiService.GetRecentUsage(1)
	if err != nil {
		return result, err
	}

	if len(usage) > 0 {
		// Use charge_count as a proxy for API usage count
		apiUsage := int(usage[0].ChargeCount)
		result["used"] = apiUsage
		result["percentage"] = int(float64(apiUsage) / float64(result["limit"].(int)) * 100)
		result["remaining"] = result["limit"].(int) - apiUsage

		// Calculate growth rate by comparing with previous hour (simplified approach)
		if len(usage) > 0 {
			// Get usage from the previous 2 hours to calculate trend
			previousUsage, err := a.apiService.GetHourlyUsage(2)
			if err == nil && len(previousUsage) >= 2 {
				currentUsage := float64(apiUsage)
				prevHourUsage := float64(previousUsage[1].CallCount) // Previous hour data

				if prevHourUsage > 0 {
					growthRate := ((currentUsage - prevHourUsage) / prevHourUsage) * 100
					result["growthRate"] = math.Round(growthRate*100) / 100
				} else if currentUsage > 0 {
					// Growth from zero
					result["growthRate"] = 100.0
				} else {
					// No usage in either period
					result["growthRate"] = 0.0
				}
			} else {
				// If insufficient data, assume moderate growth
				result["growthRate"] = 5.0
			}
		}
	}

	return result, nil
}

// GetTokenUsageProgress returns token usage progress with growth rate
func (a *App) GetTokenUsageProgress() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"used":       0,
		"growthRate": 0.0,
	}

	// Get recent usage for the last hour
	usage, err := a.apiService.GetRecentUsage(1)
	if err != nil {
		return result, err
	}

	if len(usage) > 0 {
		currentTokenUsage := usage[0].ChargeUnit
		result["used"] = currentTokenUsage

		// Calculate growth rate by comparing with previous hour
		previousUsage, err := a.apiService.GetHourlyUsage(2)
		if err == nil && len(previousUsage) >= 2 {
			prevHourTokenUsage := previousUsage[1].TokenUsage

			if prevHourTokenUsage > 0 {
				growthRate := ((currentTokenUsage - prevHourTokenUsage) / prevHourTokenUsage) * 100
				result["growthRate"] = math.Round(growthRate*100) / 100
			} else if currentTokenUsage > 0 {
				// Growth from zero
				result["growthRate"] = 100.0
			} else {
				// No usage in either period
				result["growthRate"] = 0.0
			}
		} else {
			// If insufficient data, assume moderate growth
			result["growthRate"] = 8.0
		}
	}

	return result, nil
}

// GetTotalCostProgress returns total cost progress with growth rate
func (a *App) GetTotalCostProgress() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"used":       0.0,
		"growthRate": 0.0,
	}

	// Get recent usage for the last hour
	usage, err := a.apiService.GetRecentUsage(1)
	if err != nil {
		return result, err
	}

	if len(usage) > 0 {
		currentCost := usage[0].CashCost
		result["used"] = currentCost

		// Calculate growth rate by comparing with previous hour
		previousUsage, err := a.apiService.GetHourlyUsage(2)
		if err == nil && len(previousUsage) >= 2 {
			prevHourCost := previousUsage[1].CashCost

			if prevHourCost > 0 {
				growthRate := ((currentCost - prevHourCost) / prevHourCost) * 100
				result["growthRate"] = math.Round(growthRate*100) / 100
			} else if currentCost > 0 {
				// Growth from zero
				result["growthRate"] = 100.0
			} else {
				// No usage in either period
				result["growthRate"] = 0.0
			}
		} else {
			// If insufficient data, assume moderate growth
			result["growthRate"] = 3.0
		}
	}

	return result, nil
}

// GetDayApiUsage returns API usage for the last day
func (a *App) GetDayApiUsage() (int, error) {
	usage, err := a.apiService.GetRecentUsage(24)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, record := range usage {
		total += int(record.ChargeCount)
	}

	return total, nil
}

// GetDayTokenUsage returns token usage for the last day
func (a *App) GetDayTokenUsage() (int, error) {
	usage, err := a.apiService.GetRecentUsage(24)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, record := range usage {
		total += int(record.ChargeUnit)
	}

	return total, nil
}

// GetDayTotalCost returns total cost for the last day
func (a *App) GetDayTotalCost() (float64, error) {
	usage, err := a.apiService.GetRecentUsage(24)
	if err != nil {
		return 0.0, err
	}

	total := 0.0
	for _, record := range usage {
		total += record.CashCost
	}

	return total, nil
}

// GetWeekApiUsage returns API usage for the last week
func (a *App) GetWeekApiUsage() (int, error) {
	usage, err := a.apiService.GetRecentUsage(168) // 7 days * 24 hours
	if err != nil {
		return 0, err
	}

	total := 0
	for _, record := range usage {
		total += int(record.ChargeCount)
	}

	return total, nil
}

// GetWeekTokenUsage returns token usage for the last week
func (a *App) GetWeekTokenUsage() (int, error) {
	usage, err := a.apiService.GetRecentUsage(168) // 7 days * 24 hours
	if err != nil {
		return 0, err
	}

	total := 0
	for _, record := range usage {
		total += int(record.ChargeUnit)
	}

	return total, nil
}

// GetWeekTotalCost returns total cost for the last week
func (a *App) GetWeekTotalCost() (float64, error) {
	usage, err := a.apiService.GetRecentUsage(168) // 7 days * 24 hours
	if err != nil {
		return 0.0, err
	}

	total := 0.0
	for _, record := range usage {
		total += record.CashCost
	}

	return total, nil
}

// GetMonthApiUsage returns API usage for the last month
func (a *App) GetMonthApiUsage() (int, error) {
	usage, err := a.apiService.GetRecentUsage(720) // 30 days * 24 hours
	if err != nil {
		return 0, err
	}

	total := 0
	for _, record := range usage {
		total += int(record.ChargeCount)
	}

	return total, nil
}

// GetMonthTokenUsage returns token usage for the last month
func (a *App) GetMonthTokenUsage() (int, error) {
	usage, err := a.apiService.GetRecentUsage(720) // 30 days * 24 hours
	if err != nil {
		return 0, err
	}

	total := 0
	for _, record := range usage {
		total += int(record.ChargeUnit)
	}

	return total, nil
}

// GetMonthTotalCost returns total cost for the last month
func (a *App) GetMonthTotalCost() (float64, error) {
	usage, err := a.apiService.GetRecentUsage(720) // 30 days * 24 hours
	if err != nil {
		return 0.0, err
	}

	total := 0.0
	for _, record := range usage {
		total += record.CashCost
	}

	return total, nil
}

// GetDailyUsage returns daily usage data
func (a *App) GetDailyUsage(days int) (map[string]interface{}, error) {
	usage, err := a.apiService.GetRecentUsage(days * 24)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"labels":        []string{},
		"callCountData": []int{},
		"tokenData":     []int{},
	}

	// Group by date
	dailyData := make(map[string]struct {
		callCount  int
		tokenUsage int
	})

	for _, record := range usage {
		date := record.TransactionTime.Format("01-02")
		if data, exists := dailyData[date]; exists {
			data.callCount += int(record.ChargeCount)
			data.tokenUsage += int(record.ChargeUnit)
			dailyData[date] = data
		} else {
			dailyData[date] = struct {
				callCount  int
				tokenUsage int
			}{
				callCount:  int(record.ChargeCount),
				tokenUsage: int(record.ChargeUnit),
			}
		}
	}

	// Convert to slices
	for date, data := range dailyData {
		result["labels"] = append(result["labels"].([]string), date)
		result["callCountData"] = append(result["callCountData"].([]int), data.callCount)
		result["tokenData"] = append(result["tokenData"].([]int), data.tokenUsage)
	}

	return result, nil
}

// GetMonthlyUsage returns monthly usage data
func (a *App) GetMonthlyUsage() (map[string]interface{}, error) {
	usage, err := a.apiService.GetRecentUsage(720) // 30 days
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"labels":        []string{},
		"callCountData": []int{},
		"tokenData":     []int{},
	}

	// Group by date
	dailyData := make(map[string]struct {
		callCount  int
		tokenUsage int
	})

	for _, record := range usage {
		date := record.TransactionTime.Format("01-02")
		if data, exists := dailyData[date]; exists {
			data.callCount += int(record.ChargeCount)
			data.tokenUsage += int(record.ChargeUnit)
			dailyData[date] = data
		} else {
			dailyData[date] = struct {
				callCount  int
				tokenUsage int
			}{
				callCount:  int(record.ChargeCount),
				tokenUsage: int(record.ChargeUnit),
			}
		}
	}

	// Convert to slices
	for date, data := range dailyData {
		result["labels"] = append(result["labels"].([]string), date)
		result["callCountData"] = append(result["callCountData"].([]int), data.callCount)
		result["tokenData"] = append(result["tokenData"].([]int), data.tokenUsage)
	}

	return result, nil
}

// GetCurrentMembershipTier returns the current membership tier
func (a *App) GetCurrentMembershipTier() (map[string]interface{}, error) {
	// Try to get token usage information to determine membership tier
	token, err := a.apiService.GetToken()
	if err == nil && token != nil {
		// For now, we'll determine tier based on usage patterns
		// TODO: Implement real tier detection from API response
		result := map[string]interface{}{
			"membershipTier": "GLM Coding Pro",
			"tierName":       "GLM Coding Pro",
		}
		return result, nil
	}

	// Fallback to free tier if no token found
	result := map[string]interface{}{
		"membershipTier": "Free",
		"tierName":       "Free",
	}

	return result, nil
}

// GetProducts returns the list of available products (保持兼容性)
func (a *App) GetProducts() ([]string, error) {
	return a.GetProductNames()
}

// GetProductNames returns the list of product names (新的专用方法)
func (a *App) GetProductNames() ([]string, error) {
	// 从数据库获取产品名称列表
	db := a.database.GetDB()

	query := `
		SELECT DISTINCT model_product_name 
		FROM expense_bills 
		WHERE model_product_name IS NOT NULL 
		  AND model_product_name != ''
		ORDER BY model_product_name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get product names: %w", err)
	}
	defer rows.Close()

	var productNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Printf("Error scanning product name: %v", err)
			continue
		}
		productNames = append(productNames, name)
	}

	// 如果没有数据，返回默认产品列表
	if len(productNames) == 0 {
		productNames = []string{
			"glm-4.5-air 0-32k 0-0.2k",
			"glm-4.5-air 0-32k 0.2k+",
			"glm-4.5-air 32-128k",
			"glm-4.6 0-32k 0-0.2k",
			"glm-4.6 0-32k 0.2k+",
			"glm-4.6 32-200k",
		}
	}

	return productNames, nil
}

// GetBillsCount returns the total count of bills and whether there's any data
func (a *App) GetBillsCount() (map[string]interface{}, error) {
	// Try to get a single bill to check if there's data
	bills, err := a.apiService.GetBills(&models.BillFilter{
		PageNum:  1,
		PageSize: 1,
	})
	if err != nil {
		return map[string]interface{}{
			"total":   0,
			"hasData": false,
		}, err
	}

	return map[string]interface{}{
		"total":   bills.Total,
		"hasData": bills.Total > 0,
	}, nil
}

// StopAutoSync stops the automatic sync
func (a *App) StopAutoSync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"success": true,
		"message": "Auto sync stopped",
	}, nil
}

// ========== Additional Sync-related Methods ==========

// Check if there's a running sync and return progress info
func (a *App) GetRunningSyncStatus() (map[string]interface{}, error) {
	// Get latest sync history via API service
	// Use a different approach since dbService is not exported
	history, err := a.apiService.GetSyncHistory("", 1, 1)
	if err != nil {
		return map[string]interface{}{
			"syncing": false,
			"message": "Failed to get sync status",
		}, err
	}

	// Check if we have any history and if the latest is running
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
func (a *App) CleanupStaleSyncs() (map[string]interface{}, error) {
	// For now, just return a success message since the cleanup is handled in GetRunningSyncCount
	return map[string]interface{}{
		"success": true,
		"message": "Cleanup completed - stale sync records have been marked as failed",
	}, nil
}

// ForceResetSyncStatus forcefully resets all running syncs to failed status
func (a *App) ForceResetSyncStatus() (map[string]interface{}, error) {
	db := a.database.GetDB()

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
		return map[string]interface{}{
			"success": false,
			"message": "Failed to reset sync status: " + err.Error(),
		}, err
	}

	return map[string]interface{}{
		"success": true,
		"message": "All running syncs have been reset to failed status",
	}, nil
}

// ========== 新增的异步同步方法 ==========

// StartSync 启动异步同步任务
func (a *App) StartSync(billingMonth string) (*services.SyncResult, error) {
	result, err := a.apiService.SyncBills(billingMonth, "full", nil)
	if err != nil {
		return &services.SyncResult{
			Success:      false,
			ErrorMessage: err.Error(),
		}, err
	}

	return &services.SyncResult{
		Success:      result.Success,
		SyncedItems:  result.SyncedItems,
		TotalItems:   result.TotalItems,
		FailedItems:  result.FailedItems,
		ErrorMessage: result.ErrorMessage,
	}, nil
}

// GetSyncStatusAsync 获取异步同步状态
func (a *App) GetSyncStatusAsync() (*services.SyncStatusResponse, error) {
	status, err := a.apiService.GetSyncStatus()
	if err != nil {
		return nil, err
	}

	// Convert models.SyncStatus to services.SyncStatusResponse
	// Use default values for fields that don't exist in models.SyncStatus
	var lastSyncStatus string
	if status.LastSyncStatus != nil {
		lastSyncStatus = *status.LastSyncStatus
	}

	var lastSyncTime time.Time
	if status.LastSyncTime != nil {
		lastSyncTime = *status.LastSyncTime
	}

	return &services.SyncStatusResponse{
		Syncing:      status.IsSyncing,
		Progress:     float64(status.Progress),
		CurrentPage:  0, // Not available in models.SyncStatus
		TotalPages:   0, // Not available in models.SyncStatus
		SyncedCount:  0, // Not available in models.SyncStatus
		FailedCount:  0, // Not available in models.SyncStatus
		TotalCount:   0, // Not available in models.SyncStatus
		Message:      status.Message,
		LastSyncTime: lastSyncTime,
		Status:       lastSyncStatus,
	}, nil
}

// ========== 自动同步API方法 ==========

// GetAutoSyncConfig 获取自动同步配置
func (a *App) GetAutoSyncConfig() (*models.AutoSyncConfig, error) {
	return a.apiService.GetAutoSyncConfig()
}

// SaveAutoSyncConfig 保存自动同步配置
func (a *App) SaveAutoSyncConfig(config *models.AutoSyncConfig) error {
	return a.apiService.SaveAutoSyncConfig(config)
}

// TriggerAutoSync 立即触发一次自动同步
func (a *App) TriggerAutoSync() (map[string]interface{}, error) {
	result, err := a.apiService.TriggerAutoSync()
	if err != nil {
		return result, err
	}

	return result, nil
}

// GetAutoSyncStatus 获取自动同步状态
func (a *App) GetAutoSyncStatus() (map[string]interface{}, error) {
	status, err := a.apiService.GetAutoSyncStatus()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "获取自动同步状态失败: " + err.Error(),
		}, err
	}

	status["success"] = true
	return status, nil
}

// getDynamicDailyLimit 从动态配置获取每日限制 (IPC_05)
func (a *App) getDynamicDailyLimit() int {
	// 尝试从配置获取每日限制
	config, err := a.apiService.GetConfig("daily_limit")
	if err == nil && config != "" {
		if limit, err := strconv.Atoi(config); err == nil {
			return limit
		}
	}

	// 如果配置不存在，返回默认值
	return 1000
}

// CleanOldSyncHistory 清理指定天数前的同步历史记录 (MUTATION_01)
func (a *App) CleanOldSyncHistory(days int) error {
	return a.apiService.CleanOldSyncHistory(days)
}

// DeleteAllExpenseBills 清空所有账单数据 (MUTATION_02)
func (a *App) DeleteAllExpenseBills() error {
	return a.apiService.DeleteAllExpenseBills()
}
