package services

import (
	"fmt"
	"glm-usage-monitor/models"
	"reflect"
	"time"
)

// ServiceContainer 服务容器，用于依赖注入
type ServiceContainer struct {
	services map[string]interface{}
}

// NewServiceContainer 创建新的服务容器
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		services: make(map[string]interface{}),
	}
}

// RegisterService 注册服务到容器
func (c *ServiceContainer) RegisterService(name string, service interface{}) {
	c.services[name] = service
}

// GetService 从容器获取服务
func (c *ServiceContainer) GetService(name string) (interface{}, error) {
	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not registered", name)
	}
	return service, nil
}

// APIServiceDependency API服务依赖接口
type APIServiceDependency interface {
	GetStats(startDate, endDate *time.Time, period string) (*models.StatsResponse, error)
	GetBills(filter *models.BillFilter) (*models.PaginatedResult, error)
	GetBillByID(id int) (*models.ExpenseBill, error)
	DeleteBill(id int) error
	GetBillsByDateRange(startDate, endDate time.Time, pageNum, pageSize int) (*models.PaginatedResult, error)
	SaveToken(token string) error
	GetToken() (*models.APIToken, error)
	GetAllTokens() ([]models.APIToken, error)
	DeleteToken(id int) error
	ValidateToken(token string) error
	ValidateSavedToken() (bool, error)
	GetConfig(key string) (string, error)
	SetConfig(key, value, description string) error
	GetAllConfigs() ([]models.AutoSyncConfig, error)
	GetHourlyUsage(hours int) ([]models.HourlyUsageData, error)
	GetModelDistribution(startDate, endDate *time.Time) ([]models.ModelDistributionData, error)
	GetRecentUsage(limit int) ([]models.ExpenseBill, error)
	GetUsageTrend(days int) ([]models.HourlyUsageData, error)
	GetDatabaseInfo() (map[string]interface{}, error)
	CheckAPIConnectivity() (map[string]interface{}, error)
	GetSyncStatus() (*models.SyncStatus, error)
	GetSyncHistory(syncType string, pageNum, pageSize int) (*models.PaginatedResult, error)
	SyncBills(billingMonth, syncType string, progressCallback func(*SyncProgress)) (*SyncResult, error)
	SyncRecentMonths(months int, progressCallback func(month, totalMonths int, monthProgress *SyncProgress)) ([]*SyncResult, error)
	ForceResetSyncStatus() error
	GetAutoSyncConfig() (*models.AutoSyncConfig, error)
	SaveAutoSyncConfig(config *models.AutoSyncConfig) error
	TriggerAutoSync() (map[string]interface{}, error)
	StopAutoSync() (map[string]interface{}, error)
	GetAutoSyncStatus() (map[string]interface{}, error)
	GetProductNames() ([]string, error)
	GetApiUsageProgress() (map[string]interface{}, error)
	GetTokenUsageProgress() (map[string]interface{}, error)
	GetTotalCostProgress() (map[string]interface{}, error)
}

// DependencyInjector 依赖注入器
type DependencyInjector struct {
	container *ServiceContainer
}

// NewDependencyInjector 创建新的依赖注入器
func NewDependencyInjector(container *ServiceContainer) *DependencyInjector {
	return &DependencyInjector{
		container: container,
	}
}

// InjectDependencies 注入依赖到目标结构体
func (d *DependencyInjector) InjectDependencies(target interface{}) error {
	// 获取目标结构体的类型信息
	targetType := reflect.TypeOf(target)

	// 遍历目标结构体的所有字段
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldType := field.Type

		// 只处理可导出的字段
		if field.PkgPath != "" {
			fieldName := field.Name

			// 尝试从容器获取对应的服务
			service, err := d.container.GetService(fieldName)
			if err != nil {
				// 如果服务不存在，跳过此字段
				continue
			}

			// 检查服务是否实现了期望的接口
			serviceType := reflect.TypeOf(service)
			if !serviceType.Implements(fieldType) {
				return fmt.Errorf("service %s does not implement expected interface for field %s", fieldName)
			}

			// 设置字段值
			fieldValue := reflect.ValueOf(service)
			fieldValue.Set(fieldValue)
		}
	}

	return nil
}

// InjectAPIService 注入APIService依赖
func (d *DependencyInjector) InjectAPIService(target interface{}) error {
	// 直接注入APIService实例
	apiService, err := d.container.GetService("apiService")
	if err != nil {
		return fmt.Errorf("failed to get apiService: %w", err)
	}

	// 使用反射设置apiService字段
	targetValue := reflect.ValueOf(target)
	apiServiceField := targetValue.Elem().FieldByName("ApiService")
	if !apiServiceField.IsValid() || !apiServiceField.CanSet() {
		return fmt.Errorf("target does not have ApiService field or field is not settable")
	}

	apiServiceField.Set(reflect.ValueOf(apiService))
	return nil
}

// GetServiceWithInjection 获取带依赖注入的服务
func (d *DependencyInjector) GetServiceWithInjection(serviceName string, target interface{}) (interface{}, error) {
	// 获取服务
	service, err := d.container.GetService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	// 注入依赖
	injector := NewDependencyInjector(d.container)
	err = injector.InjectDependencies(target)
	if err != nil {
		return nil, fmt.Errorf("failed to inject dependencies: %w", err)
	}

	return service, nil
}
