package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"glm-usage-monitor/models"
	"time"
)

// ZhipuAPIService provides integration with Zhipu AI API
type ZhipuAPIService struct {
	baseURL      string
	httpClient   *http.Client
	apiToken     string
	errorHandler ErrorHandler
}

// NewZhipuAPIService creates a new Zhipu API service
func NewZhipuAPIService(apiToken string) *ZhipuAPIService {
	return &ZhipuAPIService{
		baseURL: "https://bigmodel.cn/api/finance/expenseBill",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiToken:     apiToken,
		errorHandler: NewErrorHandler(),
	}
}

// BillingRequest represents the request parameters for billing API
type BillingRequest struct {
	BillingMonth string `json:"billingMonth"`
	PageNum      int    `json:"pageNum"`
	PageSize     int    `json:"pageSize"`
}

// BillingResponse represents the response from billing API
type BillingResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

// Data represents the data structure in billing response
type Data struct {
	BillList    []BillItem `json:"billList"`
	Total       int        `json:"total"`
	PageNum     int        `json:"pageNum"`
	PageSize    int        `json:"pageSize"`
	TotalPages  int        `json:"totalPages"`
	HasMore     bool       `json:"hasMore"`
}

// BillItem represents a single bill item from the API
type BillItem struct {
	ChargeName       string  `json:"chargeName"`
	ChargeType       string  `json:"chargeType"`
	ModelName        string  `json:"modelName"`
	UseGroupName     string  `json:"useGroupName"`
	GroupName        string  `json:"groupName"`
	DiscountRate     float64 `json:"discountRate"`
	CostRate         float64 `json:"costRate"`
	CashCost         float64 `json:"cashCost"`
	BillingNo        string  `json:"billingNo"`
	OrderTime        string  `json:"orderTime"`
	UseGroupID       string  `json:"useGroupId"`
	GroupID          string  `json:"groupId"`
	ChargeUnit       float64 `json:"chargeUnit"`
	ChargeCount      float64 `json:"chargeCount"`
	ChargeUnitSymbol string  `json:"chargeUnitSymbol"`
	TrialCashCost    float64 `json:"trialCashCost"`
	TimeWindow       string  `json:"timeWindow"`
}

// SyncProgress represents sync progress information
type SyncProgress struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
	SyncedItems int `json:"synced_items"`
	Progress    int `json:"progress"` // 0-100
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Success       bool           `json:"success"`
	Message       string         `json:"message"`
	TotalItems    int            `json:"total_items"`
	SyncedItems   int            `json:"synced_items"`
	FailedItems   int            `json:"failed_items"`
	SkippedItems  int            `json:"skipped_items"`
	Duration      time.Duration  `json:"duration"`
	ErrorMessage  string         `json:"error_message,omitempty"`
	ProcessedBills []models.ExpenseBill `json:"processed_bills,omitempty"`
}

// BillingMonth represents a billing month
type BillingMonth struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

// GetAvailableBillingMonths retrieves available billing months
func (s *ZhipuAPIService) GetAvailableBillingMonths() ([]BillingMonth, error) {
	// Zhipu AI API doesn't provide a direct endpoint for this
	// We'll generate months from current date back 12 months
	var months []BillingMonth
	now := time.Now()

	for i := 0; i < 12; i++ {
		date := now.AddDate(0, -i, 0)
		months = append(months, BillingMonth{
			Year:  date.Year(),
			Month: int(date.Month()),
		})
	}

	return months, nil
}

// GetBillingData retrieves billing data for a specific month
func (s *ZhipuAPIService) GetBillingData(request *BillingRequest) (*BillingResponse, error) {
	if s.apiToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// Build URL with query parameters
	baseURL, err := url.Parse(s.baseURL + "/expenseBillList")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	params := url.Values{}
	params.Add("billingMonth", request.BillingMonth)
	params.Add("pageNum", fmt.Sprintf("%d", request.PageNum))
	params.Add("pageSize", fmt.Sprintf("%d", request.PageSize))
	baseURL.RawQuery = params.Encode()

	// Create HTTP request
	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var billingResp BillingResponse
	if err := json.Unmarshal(body, &billingResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check API response code
	if billingResp.Code != 200 {
		return nil, fmt.Errorf("API returned error code %d: %s", billingResp.Code, billingResp.Message)
	}

	return &billingResp, nil
}

// ValidateAPIToken validates the API token by making a test request
func (s *ZhipuAPIService) ValidateAPIToken() error {
	// 验证API令牌是否为空
	if s.apiToken == "" {
		err := NewAuthError(ErrCodeInvalidToken, "API token is empty")
		s.errorHandler.HandleError(err, map[string]interface{}{
			"operation": "ValidateAPIToken",
		})
		return err
	}

	// 使用重试机制进行测试请求
	var validationErr error
	err := RetryWithBackUp(DefaultRetryConfig, func() error {
		// 创建测试请求
		request := &BillingRequest{
			BillingMonth: time.Now().Format("2006-01"),
			PageNum:      1,
			PageSize:     1,
		}

		var apiErr error
		_, apiErr = s.GetBillingData(request)
		if apiErr != nil {
			validationErr = WrapError(apiErr, ErrorTypeAPI, ErrCodeAPIUnauthorized, "API token validation failed")
			return validationErr
		}
		return nil
	})

	if err != nil {
		// 增强错误信息
		if validationErr != nil {
			validationErr = validationErr.(*AppError).WithContext("operation", "ValidateAPIToken").
				WithContext("timestamp", time.Now()).
				WithDetails("Failed to validate API token with test request")
			s.errorHandler.HandleError(validationErr, map[string]interface{}{
				"operation": "ValidateAPIToken",
				"retries":   DefaultRetryConfig.MaxRetries,
			})
			return validationErr
		}

		fallbackErr := NewInternalError(ErrCodeInternalError, "API token validation failed with unknown error").
			WithCause(err).
			WithDetails("Unknown error during token validation")
		s.errorHandler.HandleError(fallbackErr, map[string]interface{}{
			"operation": "ValidateAPIToken",
		})
		return fallbackErr
	}

	return nil
}

// SyncFullMonth syncs all billing data for a specific month
func (s *ZhipuAPIService) SyncFullMonth(year, month int, progressCallback func(*SyncProgress)) (*SyncResult, error) {
	billingMonth := fmt.Sprintf("%04d-%02d", year, month)

	startTime := time.Now()
	result := &SyncResult{
		Success:      true,
		TotalItems:   0,
		SyncedItems:  0,
		FailedItems:  0,
		SkippedItems: 0,
		ProcessedBills: []models.ExpenseBill{},
	}

	// Start with first page
	pageNum := 1
	pageSize := 100
	totalPages := 1

	for pageNum <= totalPages {
		request := &BillingRequest{
			BillingMonth: billingMonth,
			PageNum:      pageNum,
			PageSize:     pageSize,
		}

		// Get billing data for current page
		billingResp, err := s.GetBillingData(request)
		if err != nil {
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("Failed to fetch page %d: %v", pageNum, err)
			return result, nil
		}

		// Update total information (should be consistent across pages)
		if pageNum == 1 {
			result.TotalItems = billingResp.Data.Total
			totalPages = billingResp.Data.TotalPages
		}

		// Process bill items
		for _, billItem := range billingResp.Data.BillList {
			// Convert to map for transformation
			billMap, err := s.BillItemToMap(&billItem)
			if err != nil {
				result.FailedItems++
				continue
			}

			// Transform to expense bill
			expenseBill, err := models.TransformExpenseBill(billMap)
			if err != nil {
				result.FailedItems++
				continue
			}

			// Validate the bill
			if err := models.ValidateExpenseBill(expenseBill); err != nil {
				result.SkippedItems++
				continue
			}

			result.ProcessedBills = append(result.ProcessedBills, *expenseBill)
			result.SyncedItems++
		}

		// Update progress
		if progressCallback != nil {
			progress := &SyncProgress{
				CurrentPage: pageNum,
				TotalPages:  totalPages,
				TotalItems:  result.TotalItems,
				SyncedItems: result.SyncedItems,
			}
			if result.TotalItems > 0 {
				progress.Progress = int(float64(result.SyncedItems) / float64(result.TotalItems) * 100)
			}
			progressCallback(progress)
		}

		// Check if there are more pages
		if !billingResp.Data.HasMore {
			break
		}

		pageNum++
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// SyncRecentMonths syncs billing data for recent months
func (s *ZhipuAPIService) SyncRecentMonths(months int, progressCallback func(month, totalMonths int, monthProgress *SyncProgress)) ([]*SyncResult, error) {
	if months <= 0 {
		months = 3 // Default to last 3 months
	}

	var results []*SyncResult
	now := time.Now()

	for i := 0; i < months; i++ {
		date := now.AddDate(0, -i, 0)
		year := date.Year()
		month := int(date.Month())

		// Create progress callback for individual month
		var monthProgressCallback func(*SyncProgress)
		if progressCallback != nil {
			monthProgressCallback = func(progress *SyncProgress) {
				progressCallback(i+1, months, progress)
			}
		}

		// Sync the month
		result, err := s.SyncFullMonth(year, month, monthProgressCallback)
		if err != nil {
			return results, fmt.Errorf("failed to sync month %04d-%02d: %w", year, month, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// BillItemToMap converts BillItem to map for transformation
func (s *ZhipuAPIService) BillItemToMap(item *BillItem) (map[string]interface{}, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bill item: %w", err)
	}

	var billMap map[string]interface{}
	if err := json.Unmarshal(data, &billMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return billMap, nil
}

// GetAPIToken returns the current API token
func (s *ZhipuAPIService) GetAPIToken() string {
	return s.apiToken
}

// SetAPIToken updates the API token
func (s *ZhipuAPIService) SetAPIToken(token string) {
	s.apiToken = token
}

// GetBaseURL returns the base URL
func (s *ZhipuAPIService) GetBaseURL() string {
	return s.baseURL
}

// EstimateSyncTime estimates the time required for syncing
func (s *ZhipuAPIService) EstimateSyncTime(months int) time.Duration {
	// Rough estimation: 30 seconds per month
	// This can be adjusted based on actual performance
	return time.Duration(months*30) * time.Second
}

// GetSyncStatistics returns statistics about previous syncs
func (s *ZhipuAPIService) GetSyncStatistics(dbService *DatabaseService) (map[string]interface{}, error) {
	// Get latest sync history
	latestHistory, err := dbService.GetLatestSyncHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest sync history: %w", err)
	}

	stats := map[string]interface{}{
		"last_sync_time":    nil,
		"last_sync_status":  "never",
		"last_sync_records": 0,
		"last_sync_duration": 0,
	}

	if latestHistory != nil {
		stats["last_sync_time"] = latestHistory.StartTime
		stats["last_sync_status"] = latestHistory.Status
		stats["last_sync_records"] = latestHistory.RecordsSynced
		if latestHistory.EndTime != nil {
			stats["last_sync_duration"] = latestHistory.EndTime.Sub(latestHistory.StartTime).String()
		}
	}

	return stats, nil
}

// GetExpenseBillsPage 获取指定页数的账单数据
func (s *ZhipuAPIService) GetExpenseBillsPage(year, month, pageNum, pageSize int) (*BillingResponse, error) {
	billingMonth := fmt.Sprintf("%04d-%02d", year, month)

	request := &BillingRequest{
		BillingMonth: billingMonth,
		PageNum:      pageNum,
		PageSize:     pageSize,
	}

	return s.GetBillingData(request)
}