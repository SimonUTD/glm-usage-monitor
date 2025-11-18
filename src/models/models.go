package models

import (
	"time"
)

// ExpenseBill represents the expense_bills table structure
type ExpenseBill struct {
	ID                int       `json:"id" db:"id"`
	ChargeName        string    `json:"charge_name" db:"charge_name"`
	ChargeType        string    `json:"charge_type" db:"charge_type"`
	ModelName         string    `json:"model_name" db:"model_name"`
	UseGroupName      string    `json:"use_group_name" db:"use_group_name"`
	GroupName         string    `json:"group_name" db:"group_name"`
	DiscountRate      float64   `json:"discount_rate" db:"discount_rate"`
	CostRate          float64   `json:"cost_rate" db:"cost_rate"`
	CashCost          float64   `json:"cash_cost" db:"cash_cost"`
	BillingNo         string    `json:"billing_no" db:"billing_no"`
	OrderTime         string    `json:"order_time" db:"order_time"`
	UseGroupID        string    `json:"use_group_id" db:"use_group_id"`
	GroupID           string    `json:"group_id" db:"group_id"`
	ChargeUnit        float64   `json:"charge_unit" db:"charge_unit"`
	ChargeCount       float64   `json:"charge_count" db:"charge_count"`
	ChargeUnitSymbol  string    `json:"charge_unit_symbol" db:"charge_unit_symbol"`
	TrialCashCost     float64   `json:"trial_cash_cost" db:"trial_cash_cost"`
	TransactionTime   time.Time `json:"transaction_time" db:"transaction_time"`
	TimeWindowStart   time.Time `json:"time_window_start" db:"time_window_start"`
	TimeWindowEnd     time.Time `json:"time_window_end" db:"time_window_end"`
	TimeWindow        string    `json:"time_window" db:"time_window"`
	CreateTime        time.Time `json:"create_time" db:"create_time"`
	
	// === 新增字段：模型信息字段 ===
	APIKey            string    `json:"api_key" db:"api_key"`
	ModelCode         string    `json:"model_code" db:"model_code"`
	ModelProductType  string    `json:"model_product_type" db:"model_product_type"`
	ModelProductSubtype string  `json:"model_product_subtype" db:"model_product_subtype"`
	ModelProductCode  string    `json:"model_product_code" db:"model_product_code"`
	ModelProductName  string    `json:"model_product_name" db:"model_product_name"`
	
	// === 新增字段：支付和成本信息字段 ===
	PaymentType       string    `json:"payment_type" db:"payment_type"`
	StartTime         string    `json:"start_time" db:"start_time"`
	EndTime           string    `json:"end_time" db:"end_time"`
	BusinessID        string    `json:"business_id" db:"business_id"`
	CostPrice         float64   `json:"cost_price" db:"cost_price"`
	CostUnit          string    `json:"cost_unit" db:"cost_unit"`
	UsageCount        float64   `json:"usage_count" db:"usage_count"`
	UsageExempt       float64   `json:"usage_exempt" db:"usage_exempt"`
	UsageUnit         string    `json:"usage_unit" db:"usage_unit"`
	Currency          string    `json:"currency" db:"currency"`
	
	// === 新增字段：金额信息字段 ===
	SettlementAmount  float64   `json:"settlement_amount" db:"settlement_amount"`
	GiftDeductAmount  float64   `json:"gift_deduct_amount" db:"gift_deduct_amount"`
	DueAmount         float64   `json:"due_amount" db:"due_amount"`
	PaidAmount        float64   `json:"paid_amount" db:"paid_amount"`
	UnpaidAmount      float64   `json:"unpaid_amount" db:"unpaid_amount"`
	BillingStatus     string    `json:"billing_status" db:"billing_status"`
	InvoicingAmount   float64   `json:"invoicing_amount" db:"invoicing_amount"`
	InvoicedAmount    float64   `json:"invoiced_amount" db:"invoiced_amount"`
	
	// === 新增字段：Token业务字段 ===
	TokenAccountID    string    `json:"token_account_id" db:"token_account_id"`
	TokenResourceNo   string    `json:"token_resource_no" db:"token_resource_no"`
	TokenResourceName string    `json:"token_resource_name" db:"token_resource_name"`
	DeductUsage       float64   `json:"deduct_usage" db:"deduct_usage"`
	DeductAfter       string    `json:"deduct_after" db:"deduct_after"`
	TokenType         string    `json:"token_type" db:"token_type"`
}

// APIToken represents the api_tokens table structure
type APIToken struct {
	ID         int       `json:"id" db:"id"`
	TokenName  string    `json:"token_name" db:"token_name"`
	TokenValue string    `json:"token_value" db:"token_value"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// SyncHistory represents the sync_history table structure
type SyncHistory struct {
	ID            int        `json:"id" db:"id"`
	SyncType      string     `json:"sync_type" db:"sync_type"`
	StartTime     time.Time  `json:"start_time" db:"start_time"`
	EndTime       *time.Time `json:"end_time" db:"end_time"`
	Status        string     `json:"status" db:"status"`
	RecordsSynced int        `json:"records_synced" db:"records_synced"`
	ErrorMessage  *string    `json:"error_message" db:"error_message"`
	TotalRecords  int        `json:"total_records" db:"total_records"`
	PageSynced    int        `json:"page_synced" db:"page_synced"`
	TotalPages    int        `json:"total_pages" db:"total_pages"`
	BillingMonth  string     `json:"billing_month" db:"billing_month"`
	FailedCount   int        `json:"failed_count" db:"failed_count"`
}

// AutoSyncConfig represents the auto_sync_config table structure
type AutoSyncConfig struct {
	ID          int       `json:"id" db:"id"`
	ConfigKey   string    `json:"config_key" db:"config_key"`
	ConfigValue string    `json:"config_value" db:"config_value"`
	Description *string   `json:"description" db:"description"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	
	// 以下字段用于前端API响应，不存储在数据库中
	Enabled          bool   `json:"enabled"`
	FrequencySeconds int    `json:"frequency_seconds"`
	LastSyncTime     string `json:"last_sync_time,omitempty"`
	NextSyncTime     string `json:"next_sync_time,omitempty"`
}

// MembershipTierLimit represents the membership_tier_limits table structure
type MembershipTierLimit struct {
	ID                int       `json:"id" db:"id"`
	TierName          string    `json:"tier_name" db:"tier_name"`
	DailyLimit        *int      `json:"daily_limit" db:"daily_limit"`
	MonthlyLimit      *int      `json:"monthly_limit" db:"monthly_limit"`
	MaxTokens         *int      `json:"max_tokens" db:"max_tokens"`
	MaxContextLength  *int      `json:"max_context_length" db:"max_context_length"`
	Features          *string   `json:"features" db:"features"`
	Description       *string   `json:"description" db:"description"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// SyncStatus represents the current sync status
type SyncStatus struct {
	IsSyncing      bool      `json:"is_syncing"`
	LastSyncTime   *time.Time `json:"last_sync_time"`
	LastSyncStatus *string   `json:"last_sync_status"`
	Progress       int       `json:"progress"` // 0-100
	Message        string    `json:"message"`
}

// BillFilter represents filtering options for expense bills
type BillFilter struct {
	PageNum       int        `json:"page_num"`
	PageSize      int        `json:"page_size"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	ModelName     *string    `json:"model_name"`
	ChargeType    *string    `json:"charge_type"`
	GroupName     *string    `json:"group_name"`
	MinCashCost   *float64   `json:"min_cash_cost"`
	MaxCashCost   *float64   `json:"max_cash_cost"`
	SearchTerm    *string    `json:"search_term"`
}

// StatsResponse represents statistics response
type StatsResponse struct {
	TotalRecords        int                    `json:"total_records"`
	TotalCashCost       float64                `json:"total_cash_cost"`
	HourlyUsage         []HourlyUsageData      `json:"hourly_usage"`
	ModelDistribution   []ModelDistributionData `json:"model_distribution"`
	ChargeTypeStats     []ChargeTypeStatsData  `json:"charge_type_stats"`
	RecentUsage         []ExpenseBill          `json:"recent_usage"`
	SyncStatus          SyncStatus             `json:"sync_status"`
	MembershipInfo      *MembershipTierLimit   `json:"membership_info,omitempty"`
}

// HourlyUsageData represents hourly usage statistics
type HourlyUsageData struct {
	Hour        int     `json:"hour"`
	CallCount   int     `json:"call_count"`
	TokenUsage  float64 `json:"token_usage"`
	CashCost    float64 `json:"cash_cost"`
}

// ModelDistributionData represents model usage distribution
type ModelDistributionData struct {
	ModelName    string  `json:"model_name"`
	CallCount    int     `json:"call_count"`
	TokenUsage   float64 `json:"token_usage"`
	CashCost     float64 `json:"cash_cost"`
	Percentage   float64 `json:"percentage"`
}

// ChargeTypeStatsData represents charge type statistics
type ChargeTypeStatsData struct {
	ChargeType  string  `json:"charge_type"`
	CallCount   int     `json:"call_count"`
	CashCost    float64 `json:"cash_cost"`
	Percentage  float64 `json:"percentage"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page    int `json:"page"`
	Size    int `json:"size"`
	Total   int `json:"total"`
	HasNext bool `json:"has_next"`
}

// PaginatedResult represents paginated response data
type PaginatedResult struct {
	Data       interface{}       `json:"data"`
	Pagination PaginationParams  `json:"pagination"`
	Total      int               `json:"total"`
}