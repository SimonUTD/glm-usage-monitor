package services

import (
	"encoding/json"
	"fmt"
	"glm-usage-monitor/models"
	"log"
	"time"
)

// AutoSyncService 自动同步服务
type AutoSyncService struct {
	apiService *APIService
	dbService  *DatabaseService
	ticker     *time.Ticker
	stopChan   chan bool
	running    bool
	config     *models.AutoSyncConfig
}

// NewAutoSyncService 创建自动同步服务
func NewAutoSyncService(apiService *APIService, dbService *DatabaseService) *AutoSyncService {
	return &AutoSyncService{
		apiService: apiService,
		dbService:  dbService,
		stopChan:   make(chan bool, 1),
		running:    false,
	}
}

// GetConfig 获取自动同步配置
func (s *AutoSyncService) GetConfig() (*models.AutoSyncConfig, error) {
	// 从数据库获取配置
	configJSON, err := s.dbService.GetAutoSyncConfig("auto_sync_enabled")
	if err != nil {
		// 如果没有配置，返回默认配置
		return &models.AutoSyncConfig{
			Enabled:          false,
			FrequencySeconds: 3600, // 默认1小时
		}, nil
	}

	// 解析配置
	var config models.AutoSyncConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		log.Printf("Failed to parse auto sync config: %v", err)
		return &models.AutoSyncConfig{
			Enabled:          false,
			FrequencySeconds: 3600,
		}, nil
	}

	return &config, nil
}

// SaveConfig 保存自动同步配置
func (s *AutoSyncService) SaveConfig(config *models.AutoSyncConfig) error {
	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// 保存到数据库
	err = s.dbService.SetAutoSyncConfig("auto_sync_enabled", string(configJSON), "自动同步配置")
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// 更新内存中的配置
	s.config = config

	// 重新启动自动同步
	if config.Enabled {
		s.Stop()
		err = s.Start(config.FrequencySeconds)
		if err != nil {
			return fmt.Errorf("failed to restart auto sync: %w", err)
		}
	} else {
		s.Stop()
	}

	return nil
}

// Start 启动自动同步
func (s *AutoSyncService) Start(intervalSeconds int) error {
	if s.running {
		log.Println("Auto sync is already running")
		return nil
	}

	if intervalSeconds < 60 {
		return fmt.Errorf("interval too short, minimum 60 seconds")
	}

	s.running = true
	s.ticker = time.NewTicker(time.Duration(intervalSeconds) * time.Second)

	log.Printf("Auto sync started with interval: %d seconds", intervalSeconds)

	// 启动goroutine执行同步
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Auto sync goroutine panic recovered: %v", r)
				s.running = false
			}
		}()

		// 立即执行一次同步
		s.performAutoSync()

		// 定时执行同步
		for {
			select {
			case <-s.ticker.C:
				s.performAutoSync()
			case <-s.stopChan:
				log.Println("Auto sync goroutine stopped")
				return
			}
		}
	}()

	return nil
}

// Stop 停止自动同步
func (s *AutoSyncService) Stop() error {
	if !s.running {
		return nil
	}

	s.running = false
	
	// 停止定时器
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}

	// 发送停止信号
	select {
	case s.stopChan <- true:
	default:
		// 通道已满，避免阻塞
	}

	log.Println("Auto sync stopped")
	return nil
}

// TriggerNow 立即触发一次同步
func (s *AutoSyncService) TriggerNow() error {
	return s.performAutoSync()
}

// IsRunning 检查是否正在运行
func (s *AutoSyncService) IsRunning() bool {
	return s.running
}

// performAutoSync 执行自动同步
func (s *AutoSyncService) performAutoSync() error {
	log.Printf("Performing auto sync at %s", time.Now().Format("2006-01-02 15:04:05"))

	// 检查是否有正在运行的同步
	runningCount, err := s.dbService.GetRunningSyncCount()
	if err != nil {
		log.Printf("Failed to check running syncs: %v", err)
		return err
	}

	if runningCount > 0 {
		log.Printf("Skip auto sync: %d sync operation(s) already running", runningCount)
		return nil
	}

	// 获取当前月份
	now := time.Now()
	billingMonth := now.Format("2006-01")

	// 调用同步服务启动同步
	response, err := s.apiService.StartSync(billingMonth)
	if err != nil {
		log.Printf("Auto sync failed to start: %v", err)
		return err
	}

	if !response.Success {
		log.Printf("Auto sync start failed: %s", response.Message)
		return fmt.Errorf("auto sync start failed: %s", response.Message)
	}

	// 更新最后同步时间
	err = s.updateLastSyncTime(now)
	if err != nil {
		log.Printf("Failed to update last sync time: %v", err)
	}

	log.Printf("Auto sync started successfully for month: %s, sync ID: %d", billingMonth, response.SyncID)
	return nil
}

// updateLastSyncTime 更新最后同步时间
func (s *AutoSyncService) updateLastSyncTime(syncTime time.Time) error {
	timeStr := syncTime.Format("2006-01-02 15:04:05")
	return s.dbService.SetAutoSyncConfig("last_sync_time", timeStr, "最后同步时间")
}

// GetLastSyncTime 获取最后同步时间
func (s *AutoSyncService) GetLastSyncTime() (*time.Time, error) {
	timeStr, err := s.dbService.GetAutoSyncConfig("last_sync_time")
	if err != nil {
		return nil, err
	}

	if timeStr == "" {
		return nil, nil
	}

	syncTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last sync time: %w", err)
	}

	return &syncTime, nil
}

// GetStatus 获取自动同步状态
func (s *AutoSyncService) GetStatus() (map[string]interface{}, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	lastSyncTime, err := s.GetLastSyncTime()
	if err != nil {
		log.Printf("Failed to get last sync time: %v", err)
	}

	status := map[string]interface{}{
		"enabled":           s.running,
		"frequency_seconds": config.FrequencySeconds,
		"next_sync_time":    nil,
		"last_sync_time":    lastSyncTime,
	}

	// 计算下次同步时间
	if s.running && s.ticker != nil {
		nextSync := time.Now().Add(time.Duration(config.FrequencySeconds) * time.Second)
		status["next_sync_time"] = nextSync.Format("2006-01-02 15:04:05")
	}

	return status, nil
}