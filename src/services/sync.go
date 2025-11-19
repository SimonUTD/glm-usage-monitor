package services

import (
	"database/sql"
	"fmt"
	"glm-usage-monitor/models"
	"strconv"
	"time"
)

// ========== SyncHistory Operations ==========

// CreateSyncHistory creates a new sync history record
func (s *DatabaseService) CreateSyncHistory(history *models.SyncHistory) error {
	query := `
		INSERT INTO sync_history (
			sync_type, start_time, end_time, status, records_synced, error_message,
			total_records, page_synced, total_pages, billing_month, failed_count,
			sync_time, duration, message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		history.SyncType, history.StartTime, history.EndTime, history.Status,
		history.RecordsSynced, history.ErrorMessage, history.TotalRecords,
		history.PageSynced, history.TotalPages, history.BillingMonth, history.FailedCount,
		history.SyncTime, history.Duration, history.Message,
	)

	if err != nil {
		return fmt.Errorf("failed to create sync history: %w", err)
	}

	return nil
}

// UpdateSyncHistory updates an existing sync history record
func (s *DatabaseService) UpdateSyncHistory(id int, history *models.SyncHistory) error {
	query := `
		UPDATE sync_history
		SET end_time = ?, status = ?, records_synced = ?, error_message = ?,
		    total_records = ?, page_synced = ?, total_pages = ?,
		    billing_month = ?, failed_count = ?, sync_time = ?,
		    duration = ?, message = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		history.EndTime, history.Status, history.RecordsSynced, history.ErrorMessage,
		history.TotalRecords, history.PageSynced, history.TotalPages,
		history.BillingMonth, history.FailedCount, history.SyncTime,
		history.Duration, history.Message, id,
	)

	if err != nil {
		return fmt.Errorf("failed to update sync history: %w", err)
	}

	return nil
}

// GetSyncHistory retrieves sync history with pagination and optional sync type filtering
func (s *DatabaseService) GetSyncHistory(syncType string, pageNum, pageSize int) (*models.PaginatedResult, error) {
	// 参数验证和默认值
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	// 构建WHERE子句和参数
	whereClause := "1=1"
	args := []interface{}{}

	if syncType != "" {
		whereClause += " AND sync_type = ?"
		args = append(args, syncType)
	}

	// 优化：对于第一页，如果不需要精确总数，可以使用估算来优化性能
	var total int
	var err error

	// 性能优化：使用索引来加速COUNT查询
	var countQuery string
	if syncType != "" {
		// 如果有sync_type过滤，使用复合索引
		countQuery = `
			SELECT COUNT(*)
			FROM sync_history INDEXED BY idx_sync_history_type_start_time
			WHERE sync_type = ?
		`
		err = s.db.QueryRow(countQuery, syncType).Scan(&total)
	} else {
		// 如果没有过滤条件，直接查询全表
		countQuery = "SELECT COUNT(*) FROM sync_history"
		err = s.db.QueryRow(countQuery).Scan(&total)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to count sync history: %w", err)
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 优化查询：使用索引提示，确保使用复合索引
	query := fmt.Sprintf(`
		SELECT id, sync_type, start_time, end_time, status, records_synced, error_message,
		       total_records, page_synced, total_pages, billing_month, failed_count,
		       sync_time, duration, message
		FROM sync_history INDEXED BY idx_sync_history_type_start_time
		WHERE %s
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	// 添加分页参数
	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, pageSize, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		// 如果索引提示失败，回退到普通查询
		query = fmt.Sprintf(`
			SELECT id, sync_type, start_time, end_time, status, records_synced, error_message,
			       total_records, page_synced, total_pages
			FROM sync_history
			WHERE %s
			ORDER BY start_time DESC
			LIMIT ? OFFSET ?
		`, whereClause)

		rows, err = s.db.Query(query, queryArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to query sync history: %w", err)
		}
	}
	defer rows.Close()

	var history []models.SyncHistory
	for rows.Next() {
		var h models.SyncHistory
		err := rows.Scan(
			&h.ID, &h.SyncType, &h.StartTime, &h.EndTime, &h.Status,
			&h.RecordsSynced, &h.ErrorMessage, &h.TotalRecords,
			&h.PageSynced, &h.TotalPages, &h.BillingMonth, &h.FailedCount,
			&h.SyncTime, &h.Duration, &h.Message,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sync history: %w", err)
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sync history: %w", err)
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
		Data:       history,
		Pagination: pagination,
	}, nil
}

// GetLatestSyncHistory retrieves the latest sync history record
func (s *DatabaseService) GetLatestSyncHistory() (*models.SyncHistory, error) {
	// 性能优化：使用索引来优化查询
	query := `
		SELECT id, sync_type, start_time, end_time, status, records_synced, error_message,
		       total_records, page_synced, total_pages, billing_month, failed_count,
		       sync_time, duration, message
		FROM sync_history INDEXED BY idx_sync_history_type_start_time
		ORDER BY start_time DESC
		LIMIT 1
	`

	var history models.SyncHistory
	err := s.db.QueryRow(query).Scan(
		&history.ID, &history.SyncType, &history.StartTime, &history.EndTime, &history.Status,
		&history.RecordsSynced, &history.ErrorMessage, &history.TotalRecords,
		&history.PageSynced, &history.TotalPages, &history.BillingMonth, &history.FailedCount,
		&history.SyncTime, &history.Duration, &history.Message,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest sync history: %w", err)
	}

	return &history, nil
}

// GetRunningSyncCount counts the number of currently running syncs
func (s *DatabaseService) GetRunningSyncCount() (int, error) {
	// First, clean up any stale running syncs (older than 10 minutes)
	err := s.CleanupStaleRunningSyncs()
	if err != nil {
		// Log error but don't fail the count
		fmt.Printf("Warning: failed to cleanup stale running syncs: %v\n", err)
	}

	var count int
	// 性能优化：使用status索引来加速查询
	err = s.db.QueryRow("SELECT COUNT(*) FROM sync_history INDEXED BY idx_sync_history_status WHERE status = 'running'").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count running syncs: %w", err)
	}
	return count, nil
}

// CleanupStaleRunningSyncs marks syncs that have been running too long as failed
func (s *DatabaseService) CleanupStaleRunningSyncs() error {
	// Mark syncs older than 10 minutes as failed
	query := `
		UPDATE sync_history
		SET status = 'failed',
		    end_time = ?,
		    error_message = ?
		WHERE status = 'running'
		  AND start_time < datetime('now', '-10 minutes')
	`

	errorMessage := "Sync marked as failed due to timeout"
	_, err := s.db.Exec(query, time.Now(), errorMessage)
	if err != nil {
		return fmt.Errorf("failed to cleanup stale running syncs: %w", err)
	}

	return nil
}

// ========== AutoSyncConfig Operations ==========

// GetAutoSyncConfig retrieves a configuration value by key
func (s *DatabaseService) GetAutoSyncConfig(key string) (string, error) {
	query := "SELECT config_value FROM auto_sync_config WHERE config_key = ?"
	var value string
	err := s.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("config key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get config value: %w", err)
	}
	return value, nil
}

// SetAutoSyncConfig saves a configuration value
func (s *DatabaseService) SetAutoSyncConfig(key, value, description string) error {
	query := `
		INSERT INTO auto_sync_config (config_key, config_value, description, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(config_key) DO UPDATE SET
			config_value = excluded.config_value,
			description = COALESCE(excluded.description, excluded.description),
			updated_at = excluded.updated_at
	`

	_, err := s.db.Exec(query, key, value, description, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save config value: %w", err)
	}

	return nil
}

// GetAllAutoSyncConfigs retrieves all configuration values (DB_03: 适配新结构)
func (s *DatabaseService) GetAllAutoSyncConfigs() ([]models.AutoSyncConfig, error) {
	// 检查表结构，如果还是旧的键值对结构，需要转换
	var isOldStructure bool

	// 尝试查询新结构的字段
	checkQuery := "SELECT COUNT(*) FROM pragma_table_info('auto_sync_config') WHERE name = 'enabled'"
	var count int
	err := s.db.QueryRow(checkQuery).Scan(&count)
	if err != nil {
		// 如果pragma查询失败，假设是新结构
		isOldStructure = false
	} else {
		isOldStructure = count == 0
	}

	if isOldStructure {
		// 从旧结构转换
		return s.convertOldConfigToNew()
	}

	// 查询新结构
	query := `
		SELECT id, enabled, frequency_seconds, last_sync_time, next_sync_time,
		       sync_type, billing_month, max_retries, retry_delay, created_at, updated_at
		FROM auto_sync_config
		ORDER BY id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query auto sync configs: %w", err)
	}
	defer rows.Close()

	var configs []models.AutoSyncConfig
	for rows.Next() {
		var config models.AutoSyncConfig
		err := rows.Scan(
			&config.ID, &config.Enabled, &config.FrequencySeconds,
			&config.LastSyncTime, &config.NextSyncTime, &config.SyncType,
			&config.BillingMonth, &config.MaxRetries, &config.RetryDelay,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auto sync config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// convertOldConfigToNew converts old key-value config to new structured config
func (s *DatabaseService) convertOldConfigToNew() ([]models.AutoSyncConfig, error) {
	// 获取旧配置
	oldQuery := "SELECT config_key, config_value FROM auto_sync_config"
	rows, err := s.db.Query(oldQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query old auto sync configs: %w", err)
	}
	defer rows.Close()

	// 默认配置
	config := models.AutoSyncConfig{
		Enabled:          false,
		FrequencySeconds: 300, // 5 minutes
		SyncType:         "incremental",
		MaxRetries:       3,
		RetryDelay:       60, // 1 minute
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 解析旧配置值
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}

		switch key {
		case "enabled":
			config.Enabled = value == "true"
		case "frequency_seconds":
			if freq, err := strconv.Atoi(value); err == nil {
				config.FrequencySeconds = freq
			}
		}
	}

	return []models.AutoSyncConfig{config}, nil
}

// GetAutoSyncStatus retrieves the current auto-sync status
func (s *DatabaseService) GetAutoSyncStatus() (*models.SyncStatus, error) {
	status := &models.SyncStatus{
		IsSyncing: false,
		Progress:  0,
		Message:   "Idle",
	}

	// Check if there are running syncs
	runningCount, err := s.GetRunningSyncCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get running sync count: %w", err)
	}

	if runningCount > 0 {
		status.IsSyncing = true
		status.Message = fmt.Sprintf("%d sync operation(s) in progress", runningCount)
		return status, nil
	}

	// Get last sync history
	latestHistory, err := s.GetLatestSyncHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sync history: %w", err)
	}

	if latestHistory != nil {
		status.LastSyncTime = &latestHistory.StartTime
		status.LastSyncStatus = &latestHistory.Status

		switch latestHistory.Status {
		case "completed":
			status.Message = fmt.Sprintf("Last sync completed successfully at %s",
				latestHistory.StartTime.Format("2006-01-02 15:04:05"))
		case "failed":
			status.Message = fmt.Sprintf("Last sync failed at %s",
				latestHistory.StartTime.Format("2006-01-02 15:04:05"))
			if latestHistory.ErrorMessage != nil {
				status.Message += fmt.Sprintf(": %s", *latestHistory.ErrorMessage)
			}
		default:
			status.Message = "Last sync status unknown"
		}
	}

	return status, nil
}

// ========== MembershipTierLimit Operations ==========

// ResetRunningSyncs forces reset of all running syncs to failed status
func (s *DatabaseService) ResetRunningSyncs() error {
	query := `
		UPDATE sync_history
		SET status = 'failed',
		    end_time = ?,
		    error_message = ?
		WHERE status = 'running'
	`

	errorMessage := "Sync status reset manually"
	_, err := s.db.Exec(query, time.Now(), errorMessage)
	if err != nil {
		return fmt.Errorf("failed to reset running syncs: %w", err)
	}

	return nil
}

// GetMembershipTierLimits retrieves membership tier information by name (plural version for compatibility)
func (s *DatabaseService) GetMembershipTierLimits(tierName string) (*models.MembershipTierLimit, error) {
	return s.GetMembershipTierLimit(tierName)
}

// GetMembershipTierLimit retrieves membership tier information by name
func (s *DatabaseService) GetMembershipTierLimit(tierName string) (*models.MembershipTierLimit, error) {
	query := `
		SELECT id, tier_name, daily_limit, monthly_limit, max_tokens, max_context_length,
		       features, description, updated_at
		FROM membership_tier_limits
		WHERE tier_name = ?
	`

	var limit models.MembershipTierLimit
	err := s.db.QueryRow(query, tierName).Scan(
		&limit.ID, &limit.TierName, &limit.DailyLimit, &limit.MonthlyLimit,
		&limit.MaxTokens, &limit.MaxContextLength, &limit.Features,
		&limit.Description, &limit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("membership tier not found: %s", tierName)
		}
		return nil, fmt.Errorf("failed to get membership tier limit: %w", err)
	}

	return &limit, nil
}

// SaveMembershipTierLimit saves a membership tier limit
func (s *DatabaseService) SaveMembershipTierLimit(limit *models.MembershipTierLimit) error {
	query := `
		INSERT INTO membership_tier_limits
		(tier_name, daily_limit, monthly_limit, max_tokens, max_context_length,
		 features, description, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(tier_name) DO UPDATE SET
			daily_limit = excluded.daily_limit,
			monthly_limit = excluded.monthly_limit,
			max_tokens = excluded.max_tokens,
			max_context_length = excluded.max_context_length,
			features = excluded.features,
			description = excluded.description,
			updated_at = excluded.updated_at
	`

	_, err := s.db.Exec(query,
		limit.TierName, limit.DailyLimit, limit.MonthlyLimit, limit.MaxTokens,
		limit.MaxContextLength, limit.Features, limit.Description, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save membership tier limit: %w", err)
	}

	return nil
}

// GetAllMembershipTierLimits retrieves all membership tier limits
func (s *DatabaseService) GetAllMembershipTierLimits() ([]models.MembershipTierLimit, error) {
	query := `
		SELECT id, tier_name, daily_limit, monthly_limit, max_tokens, max_context_length,
		       features, description, updated_at
		FROM membership_tier_limits
		ORDER BY tier_name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query membership tier limits: %w", err)
	}
	defer rows.Close()

	var limits []models.MembershipTierLimit
	for rows.Next() {
		var limit models.MembershipTierLimit
		err := rows.Scan(
			&limit.ID, &limit.TierName, &limit.DailyLimit, &limit.MonthlyLimit,
			&limit.MaxTokens, &limit.MaxContextLength, &limit.Features,
			&limit.Description, &limit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan membership tier limit: %w", err)
		}
		limits = append(limits, limit)
	}

	return limits, nil
}
