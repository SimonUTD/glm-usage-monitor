package main

import (
	"glm-usage-monitor/models"
	"glm-usage-monitor/services"
	"time"
)

// ========== Bill Management API Bindings ==========

// GetBills retrieves expense bills with filtering and pagination
func (a *App) GetBills(filter *models.BillFilter) (*models.PaginatedResult, error) {
	return a.apiService.GetBills(filter)
}

// GetBillByID retrieves a single expense bill by ID
func (a *App) GetBillByID(id int) (*models.ExpenseBill, error) {
	return a.apiService.GetBillByID(id)
}

// DeleteBill deletes an expense bill by ID
func (a *App) DeleteBill(id int) error {
	return a.apiService.DeleteBill(id)
}

// GetBillsByDateRange retrieves bills within a date range
func (a *App) GetBillsByDateRange(startDate, endDate time.Time, pageNum, pageSize int) (*models.PaginatedResult, error) {
	return a.apiService.GetBillsByDateRange(startDate, endDate, pageNum, pageSize)
}

// ========== Statistics API Bindings ==========

// GetStats retrieves overall usage statistics
func (a *App) GetStats(startDate, endDate *time.Time) (*models.StatsResponse, error) {
	return a.apiService.GetStats(startDate, endDate)
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

// SaveToken saves an API token
func (a *App) SaveToken(tokenName, tokenValue string) error {
	return a.apiService.SaveToken(tokenName, tokenValue)
}

// GetToken retrieves the active API token
func (a *App) GetToken() (*models.APIToken, error) {
	return a.apiService.GetToken()
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
	return a.apiService.GetSyncStatus()
}

// GetSyncHistory retrieves sync history
func (a *App) GetSyncHistory(syncType string, pageNum, pageSize int) (*models.PaginatedResult, error) {
	return a.apiService.GetSyncHistory(syncType, pageNum, pageSize)
}

// SyncBills starts a sync operation for billing data
func (a *App) SyncBills(year, month int) (*services.SyncResult, error) {
	return a.apiService.SyncBills(year, month, nil) // No progress callback for now
}

// SyncRecentMonths syncs billing data for recent months
func (a *App) SyncRecentMonths(months int) ([]*services.SyncResult, error) {
	return a.apiService.SyncRecentMonths(months, nil) // No progress callback for now
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