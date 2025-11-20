package services

import (
	"database/sql"
	"fmt"
	"glm-usage-monitor/models"
	"log"
	"strings"
	"time"
)

// DatabaseService provides CRUD operations for all database tables
type DatabaseService struct {
	db *sql.DB
}

// NewDatabaseService creates a new database service
func NewDatabaseService(db *sql.DB) *DatabaseService {
	return &DatabaseService{db: db}
}

// GetDB returns the underlying database connection
func (s *DatabaseService) GetDB() *sql.DB {
	return s.db
}

// ========== ExpenseBill Operations ==========

// CreateExpenseBill creates a new expense bill record
func (s *DatabaseService) CreateExpenseBill(bill *models.ExpenseBill) error {
	query := `
		INSERT INTO expense_bills (
			charge_name, charge_type, model_name, use_group_name, group_name,
			discount_rate, cost_rate, cash_cost, billing_no, order_time,
			use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			trial_cash_cost, transaction_time, time_window_start, time_window_end,
			time_window, create_time,
			
			-- 新增字段：模型信息字段
			api_key, model_code, model_product_type, model_product_subtype, model_product_code, model_product_name,
			
			-- 新增字段：支付和成本信息字段
			payment_type, start_time, end_time, business_id, cost_price, cost_unit, usage_count, usage_exempt, usage_unit, currency,
			
			-- 新增字段：金额信息字段
			settlement_amount, gift_deduct_amount, due_amount, paid_amount, unpaid_amount, billing_status, invoicing_amount, invoiced_amount,
			
			-- 新增字段：Token业务字段
			token_account_id, token_resource_no, token_resource_name, deduct_usage, deduct_after, token_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		// 原有字段
		bill.ChargeName, bill.ChargeType, bill.ModelName, bill.UseGroupName, bill.GroupName,
		bill.DiscountRate, bill.CostRate, bill.CashCost, bill.BillingNo, bill.OrderTime,
		bill.UseGroupID, bill.GroupID, bill.ChargeUnit, bill.ChargeCount, bill.ChargeUnitSymbol,
		bill.TrialCashCost, bill.TransactionTime, bill.TimeWindowStart, bill.TimeWindowEnd,
		bill.TimeWindow, bill.CreateTime,

		// 模型信息字段
		bill.APIKey, bill.ModelCode, bill.ModelProductType, bill.ModelProductSubtype, bill.ModelProductCode, bill.ModelProductName,

		// 支付和成本信息字段
		bill.PaymentType, bill.StartTime, bill.EndTime, bill.BusinessID, bill.CostPrice, bill.CostUnit, bill.UsageCount, bill.UsageExempt, bill.UsageUnit, bill.Currency,

		// 金额信息字段
		bill.SettlementAmount, bill.GiftDeductAmount, bill.DueAmount, bill.PaidAmount, bill.UnpaidAmount, bill.BillingStatus, bill.InvoicingAmount, bill.InvoicedAmount,

		// Token业务字段
		bill.TokenAccountID, bill.TokenResourceNo, bill.TokenResourceName, bill.DeductUsage, bill.DeductAfter, bill.TokenType,
	)

	if err != nil {
		return fmt.Errorf("failed to create expense bill: %w", err)
	}

	return nil
}

// BatchCreateExpenseBills creates multiple expense bills in a transaction
func (s *DatabaseService) BatchCreateExpenseBills(bills []*models.ExpenseBill) error {
	if len(bills) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO expense_bills (
			charge_name, charge_type, model_name, use_group_name, group_name,
			discount_rate, cost_rate, cash_cost, billing_no, order_time,
			use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			trial_cash_cost, transaction_time, time_window_start, time_window_end,
			time_window, create_time,
			
			-- DB_01: 缺失的关键字段
			billing_date, billing_time, customer_id, order_no, original_amount, original_cost_price,
			discount_type, credit_pay_amount, third_party, cash_amount, api_usage,
			
			-- 模型信息字段
			api_key, model_code, model_product_type, model_product_subtype, model_product_code, model_product_name,
			
			-- 支付和成本信息字段
			payment_type, start_time, end_time, business_id, cost_price, cost_unit, usage_count, usage_exempt, usage_unit, currency,
			
			-- 金额信息字段
			settlement_amount, gift_deduct_amount, due_amount, paid_amount, unpaid_amount, billing_status, invoicing_amount, invoiced_amount,
			
			-- Token业务字段
			token_account_id, token_resource_no, token_resource_name, deduct_usage, deduct_after, token_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, bill := range bills {
		if err := models.ValidateExpenseBill(bill); err != nil {
			log.Printf("Invalid bill data, skipping: %v", err)
			continue
		}

		_, err := stmt.Exec(
			// 原有字段
			bill.ChargeName, bill.ChargeType, bill.ModelName, bill.UseGroupName, bill.GroupName,
			bill.DiscountRate, bill.CostRate, bill.CashCost, bill.BillingNo, bill.OrderTime,
			bill.UseGroupID, bill.GroupID, bill.ChargeUnit, bill.ChargeCount, bill.ChargeUnitSymbol,
			bill.TrialCashCost, bill.TransactionTime, bill.TimeWindowStart, bill.TimeWindowEnd,
			bill.TimeWindow, bill.CreateTime,

			// DB_01: 缺失的关键字段
			bill.BillingDate, bill.BillingTime, bill.CustomerID, bill.OrderNo, bill.OriginalAmount, bill.OriginalCostPrice,
			bill.DiscountType, bill.CreditPayAmount, bill.ThirdParty, bill.CashAmount, bill.APIUsage,

			// 模型信息字段
			bill.APIKey, bill.ModelCode, bill.ModelProductType, bill.ModelProductSubtype, bill.ModelProductCode, bill.ModelProductName,

			// 支付和成本信息字段
			bill.PaymentType, bill.StartTime, bill.EndTime, bill.BusinessID, bill.CostPrice, bill.CostUnit, bill.UsageCount, bill.UsageExempt, bill.UsageUnit, bill.Currency,

			// 金额信息字段
			bill.SettlementAmount, bill.GiftDeductAmount, bill.DueAmount, bill.PaidAmount, bill.UnpaidAmount, bill.BillingStatus, bill.InvoicingAmount, bill.InvoicedAmount,

			// Token业务字段
			bill.TokenAccountID, bill.TokenResourceNo, bill.TokenResourceName, bill.DeductUsage, bill.DeductAfter, bill.TokenType,
		)

		if err != nil {
			return fmt.Errorf("failed to insert bill %s: %w", bill.BillingNo, err)
		}
	}

	return tx.Commit()
}

// GetExpenseBills retrieves expense bills with filtering and pagination
func (s *DatabaseService) GetExpenseBills(filter *models.BillFilter) (*models.PaginatedResult, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	// Build WHERE conditions
	if filter.StartDate != nil {
		whereConditions = append(whereConditions, "DATE(transaction_time) >= DATE(?)")
		args = append(args, filter.StartDate.Format("2006-01-02"))
	}

	if filter.EndDate != nil {
		whereConditions = append(whereConditions, "DATE(transaction_time) <= DATE(?)")
		args = append(args, filter.EndDate.Format("2006-01-02"))
	}

	if filter.ModelName != nil && *filter.ModelName != "" {
		whereConditions = append(whereConditions, "model_name = ?")
		args = append(args, *filter.ModelName)
	}

	if filter.ChargeType != nil && *filter.ChargeType != "" {
		whereConditions = append(whereConditions, "charge_type = ?")
		args = append(args, *filter.ChargeType)
	}

	if filter.GroupName != nil && *filter.GroupName != "" {
		whereConditions = append(whereConditions, "group_name LIKE ?")
		args = append(args, "%"+*filter.GroupName+"%")
	}

	if filter.MinCashCost != nil {
		whereConditions = append(whereConditions, "cash_cost >= ?")
		args = append(args, *filter.MinCashCost)
	}

	if filter.MaxCashCost != nil {
		whereConditions = append(whereConditions, "cash_cost <= ?")
		args = append(args, *filter.MaxCashCost)
	}

	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		whereConditions = append(whereConditions, "(charge_name LIKE ? OR model_name LIKE ? OR billing_no LIKE ?)")
		searchTerm := "%" + *filter.SearchTerm + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM expense_bills WHERE %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count expense bills: %w", err)
	}

	// Calculate pagination
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	pageNum := filter.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}

	offset := (pageNum - 1) * pageSize

	// Get data
	query := fmt.Sprintf(`
		SELECT id, charge_name, charge_type, model_name, use_group_name, group_name,
			   discount_rate, cost_rate, cash_cost, billing_no, order_time,
			   use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			   trial_cash_cost, transaction_time, time_window_start, time_window_end,
			   time_window, create_time,
			   
			   -- DB_01: 缺失的关键字段
			   billing_date, billing_time, customer_id, order_no, original_amount, original_cost_price,
			   discount_type, credit_pay_amount, third_party, cash_amount, api_usage,
			   
			   -- 模型信息字段
			   api_key, model_code, model_product_type, model_product_subtype, model_product_code, model_product_name,
			   
			   -- 支付和成本信息字段
			   payment_type, start_time, end_time, business_id, cost_price, cost_unit, usage_count, usage_exempt, usage_unit, currency,
			   
			   -- 金额信息字段
			   settlement_amount, gift_deduct_amount, due_amount, paid_amount, unpaid_amount, billing_status, invoicing_amount, invoiced_amount,
			   
			   -- Token业务字段
			   token_account_id, token_resource_no, token_resource_name, deduct_usage, deduct_after, token_type
		FROM expense_bills
		WHERE %s
		ORDER BY transaction_time DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query expense bills: %w", err)
	}
	defer rows.Close()

	var bills []models.ExpenseBill
	for rows.Next() {
		var bill models.ExpenseBill
		err := rows.Scan(
			&bill.ID, &bill.ChargeName, &bill.ChargeType, &bill.ModelName, &bill.UseGroupName, &bill.GroupName,
			&bill.DiscountRate, &bill.CostRate, &bill.CashCost, &bill.BillingNo, &bill.OrderTime,
			&bill.UseGroupID, &bill.GroupID, &bill.ChargeUnit, &bill.ChargeCount, &bill.ChargeUnitSymbol,
			&bill.TrialCashCost, &bill.TransactionTime, &bill.TimeWindowStart, &bill.TimeWindowEnd,
			&bill.TimeWindow, &bill.CreateTime,

			// DB_01: 缺失的关键字段
			&bill.BillingDate, &bill.BillingTime, &bill.CustomerID, &bill.OrderNo, &bill.OriginalAmount, &bill.OriginalCostPrice,
			&bill.DiscountType, &bill.CreditPayAmount, &bill.ThirdParty, &bill.CashAmount, &bill.APIUsage,

			// 模型信息字段
			&bill.APIKey, &bill.ModelCode, &bill.ModelProductType, &bill.ModelProductSubtype, &bill.ModelProductCode, &bill.ModelProductName,

			// 支付和成本信息字段
			&bill.PaymentType, &bill.StartTime, &bill.EndTime, &bill.BusinessID, &bill.CostPrice, &bill.CostUnit, &bill.UsageCount, &bill.UsageExempt, &bill.UsageUnit, &bill.Currency,

			// 金额信息字段
			&bill.SettlementAmount, &bill.GiftDeductAmount, &bill.DueAmount, &bill.PaidAmount, &bill.UnpaidAmount, &bill.BillingStatus, &bill.InvoicingAmount, &bill.InvoicedAmount,

			// Token业务字段
			&bill.TokenAccountID, &bill.TokenResourceNo, &bill.TokenResourceName, &bill.DeductUsage, &bill.DeductAfter, &bill.TokenType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense bill: %w", err)
		}
		bills = append(bills, bill)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expense bills: %w", err)
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
		Data:       bills,
		Pagination: pagination,
	}, nil
}

// GetExpenseBillByID retrieves a single expense bill by ID
func (s *DatabaseService) GetExpenseBillByID(id string) (*models.ExpenseBill, error) {
	query := `
		SELECT id, charge_name, charge_type, model_name, use_group_name, group_name,
			   discount_rate, cost_rate, cash_cost, billing_no, order_time,
			   use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			   trial_cash_cost, transaction_time, time_window_start, time_window_end,
			   time_window, create_time,
			   
			   -- DB_01: 缺失的关键字段
			   billing_date, billing_time, customer_id, order_no, original_amount, original_cost_price,
			   discount_type, credit_pay_amount, third_party, cash_amount, api_usage,
			   
			   -- 模型信息字段
			   api_key, model_code, model_product_type, model_product_subtype, model_product_code, model_product_name,
			   
			   -- 支付和成本信息字段
			   payment_type, start_time, end_time, business_id, cost_price, cost_unit, usage_count, usage_exempt, usage_unit, currency,
			   
			   -- 金额信息字段
			   settlement_amount, gift_deduct_amount, due_amount, paid_amount, unpaid_amount, billing_status, invoicing_amount, invoiced_amount,
			   
			   -- Token业务字段
			   token_account_id, token_resource_no, token_resource_name, deduct_usage, deduct_after, token_type
		FROM expense_bills
		WHERE id = ?
	`

	var bill models.ExpenseBill
	err := s.db.QueryRow(query, id).Scan(
		&bill.ID, &bill.ChargeName, &bill.ChargeType, &bill.ModelName, &bill.UseGroupName, &bill.GroupName,
		&bill.DiscountRate, &bill.CostRate, &bill.CashCost, &bill.BillingNo, &bill.OrderTime,
		&bill.UseGroupID, &bill.GroupID, &bill.ChargeUnit, &bill.ChargeCount, &bill.ChargeUnitSymbol,
		&bill.TrialCashCost, &bill.TransactionTime, &bill.TimeWindowStart, &bill.TimeWindowEnd,
		&bill.TimeWindow, &bill.CreateTime,

		// DB_01: 缺失的关键字段
		&bill.BillingDate, &bill.BillingTime, &bill.CustomerID, &bill.OrderNo, &bill.OriginalAmount, &bill.OriginalCostPrice,
		&bill.DiscountType, &bill.CreditPayAmount, &bill.ThirdParty, &bill.CashAmount, &bill.APIUsage,

		// 模型信息字段
		&bill.APIKey, &bill.ModelCode, &bill.ModelProductType, &bill.ModelProductSubtype, &bill.ModelProductCode, &bill.ModelProductName,

		// 支付和成本信息字段
		&bill.PaymentType, &bill.StartTime, &bill.EndTime, &bill.BusinessID, &bill.CostPrice, &bill.CostUnit, &bill.UsageCount, &bill.UsageExempt, &bill.UsageUnit, &bill.Currency,

		// 金额信息字段
		&bill.SettlementAmount, &bill.GiftDeductAmount, &bill.DueAmount, &bill.PaidAmount, &bill.UnpaidAmount, &bill.BillingStatus, &bill.InvoicingAmount, &bill.InvoicedAmount,

		// Token业务字段
		&bill.TokenAccountID, &bill.TokenResourceNo, &bill.TokenResourceName, &bill.DeductUsage, &bill.DeductAfter, &bill.TokenType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expense bill not found")
		}
		return nil, fmt.Errorf("failed to get expense bill: %w", err)
	}

	return &bill, nil
}

// DeleteExpenseBill deletes an expense bill by ID
func (s *DatabaseService) DeleteExpenseBill(id string) error {
	query := "DELETE FROM expense_bills WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense bill: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("expense bill not found")
	}

	return nil
}

// GetExpenseBillsByBillingNo retrieves expense bills by billing number
func (s *DatabaseService) GetExpenseBillsByBillingNo(billingNo string) ([]models.ExpenseBill, error) {
	query := `
		SELECT id, charge_name, charge_type, model_name, use_group_name, group_name,
			   discount_rate, cost_rate, cash_cost, billing_no, order_time,
			   use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			   trial_cash_cost, transaction_time, time_window_start, time_window_end,
			   time_window, create_time,
			   
			   -- DB_01: 缺失的关键字段
			   billing_date, billing_time, customer_id, order_no, original_amount, original_cost_price,
			   discount_type, credit_pay_amount, third_party, cash_amount, api_usage,
			   
			   -- 模型信息字段
			   api_key, model_code, model_product_type, model_product_subtype, model_product_code, model_product_name,
			   
			   -- 支付和成本信息字段
			   payment_type, start_time, end_time, business_id, cost_price, cost_unit, usage_count, usage_exempt, usage_unit, currency,
			   
			   -- 金额信息字段
			   settlement_amount, gift_deduct_amount, due_amount, paid_amount, unpaid_amount, billing_status, invoicing_amount, invoiced_amount,
			   
			   -- Token业务字段
			   token_account_id, token_resource_no, token_resource_name, deduct_usage, deduct_after, token_type
		FROM expense_bills
		WHERE billing_no = ?
		ORDER BY transaction_time DESC
	`

	rows, err := s.db.Query(query, billingNo)
	if err != nil {
		return nil, fmt.Errorf("failed to query expense bills by billing no: %w", err)
	}
	defer rows.Close()

	var bills []models.ExpenseBill
	for rows.Next() {
		var bill models.ExpenseBill
		err := rows.Scan(
			&bill.ID, &bill.ChargeName, &bill.ChargeType, &bill.ModelName, &bill.UseGroupName, &bill.GroupName,
			&bill.DiscountRate, &bill.CostRate, &bill.CashCost, &bill.BillingNo, &bill.OrderTime,
			&bill.UseGroupID, &bill.GroupID, &bill.ChargeUnit, &bill.ChargeCount, &bill.ChargeUnitSymbol,
			&bill.TrialCashCost, &bill.TransactionTime, &bill.TimeWindowStart, &bill.TimeWindowEnd,
			&bill.TimeWindow, &bill.CreateTime,

			// DB_01: 缺失的关键字段
			&bill.BillingDate, &bill.BillingTime, &bill.CustomerID, &bill.OrderNo, &bill.OriginalAmount, &bill.OriginalCostPrice,
			&bill.DiscountType, &bill.CreditPayAmount, &bill.ThirdParty, &bill.CashAmount, &bill.APIUsage,

			// 模型信息字段
			&bill.APIKey, &bill.ModelCode, &bill.ModelProductType, &bill.ModelProductSubtype, &bill.ModelProductCode, &bill.ModelProductName,

			// 支付和成本信息字段
			&bill.PaymentType, &bill.StartTime, &bill.EndTime, &bill.BusinessID, &bill.CostPrice, &bill.CostUnit, &bill.UsageCount, &bill.UsageExempt, &bill.UsageUnit, &bill.Currency,

			// 金额信息字段
			&bill.SettlementAmount, &bill.GiftDeductAmount, &bill.DueAmount, &bill.PaidAmount, &bill.UnpaidAmount, &bill.BillingStatus, &bill.InvoicingAmount, &bill.InvoicedAmount,

			// Token业务字段
			&bill.TokenAccountID, &bill.TokenResourceNo, &bill.TokenResourceName, &bill.DeductUsage, &bill.DeductAfter, &bill.TokenType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense bill: %w", err)
		}
		bills = append(bills, bill)
	}

	return bills, nil
}

// ========== APIToken Operations ==========

// SaveAPIToken saves an API token (single token design)
func (s *DatabaseService) SaveAPIToken(token *models.APIToken) error {
	// 使用INSERT OR REPLACE简化逻辑：如果存在就替换，不存在就插入
	query := `
		INSERT OR REPLACE INTO api_tokens (
			id, token_name, token_value, provider, token_type, is_active,
			daily_limit, monthly_limit, expires_at, last_used_at,
			created_at, updated_at
		) VALUES (
			1, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := s.db.Exec(query,
		token.TokenName, token.TokenValue, token.Provider, token.TokenType,
		token.DailyLimit, token.MonthlyLimit, token.ExpiresAt, token.LastUsedAt,
		token.CreatedAt, token.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save API token: %w", err)
	}

	log.Printf("DEBUG: Successfully saved/updated token")
	return nil
}

// GetActiveAPIToken retrieves the API token (single token design)
func (s *DatabaseService) GetActiveAPIToken() (*models.APIToken, error) {
	query := `
		SELECT id, token_name, token_value,
			   CASE WHEN provider IS NULL THEN '' ELSE provider END as provider,
			   CASE WHEN token_type IS NULL THEN '' ELSE token_type END as token_type,
			   is_active,
			   daily_limit, monthly_limit, expires_at, last_used_at,
			   created_at, updated_at
		FROM api_tokens
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var token models.APIToken
	err := s.db.QueryRow(query).Scan(
		&token.ID, &token.TokenName, &token.TokenValue, &token.Provider, &token.TokenType, &token.IsActive,
		&token.DailyLimit, &token.MonthlyLimit, &token.ExpiresAt, &token.LastUsedAt,
		&token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// 没有token是正常情况，返回nil而不是错误
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API token: %w", err)
	}

	return &token, nil
}

// GetAllAPITokens retrieves all API tokens
func (s *DatabaseService) GetAllAPITokens() ([]models.APIToken, error) {
	query := `
		SELECT id, token_name, token_value,
			   CASE WHEN provider IS NULL THEN '' ELSE provider END as provider,
			   CASE WHEN token_type IS NULL THEN '' ELSE token_type END as token_type,
			   is_active,
			   daily_limit, monthly_limit, expires_at, last_used_at,
			   created_at, updated_at
		FROM api_tokens
		ORDER BY updated_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query API tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.APIToken
	for rows.Next() {
		var token models.APIToken
		err := rows.Scan(
			&token.ID, &token.TokenName, &token.TokenValue, &token.Provider, &token.TokenType, &token.IsActive,
			&token.DailyLimit, &token.MonthlyLimit, &token.ExpiresAt, &token.LastUsedAt,
			&token.CreatedAt, &token.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// DeactivateAPIToken deactivates an API token by ID
func (s *DatabaseService) DeactivateAPIToken(id int) error {
	query := "UPDATE api_tokens SET is_active = 0, updated_at = ? WHERE id = ?"
	_, err := s.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to deactivate API token: %w", err)
	}
	return nil
}

// DeleteAPIToken deletes an API token by ID
func (s *DatabaseService) DeleteAPIToken(id int) error {
	query := "DELETE FROM api_tokens WHERE id = ?"
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API token: %w", err)
	}
	return nil
}

// ========== 事务相关方法 ==========

// BeginTx 开始一个数据库事务
func (s *DatabaseService) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}

// CreateOrUpdateExpenseBillInTx 在事务中创建或更新账单
func (s *DatabaseService) CreateOrUpdateExpenseBillInTx(tx *sql.Tx, bill *models.ExpenseBill) error {
	// 首先检查是否已存在
	var count int
	checkQuery := "SELECT COUNT(*) FROM expense_bills WHERE billing_no = ?"
	err := tx.QueryRow(checkQuery, bill.BillingNo).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing bill: %w", err)
	}

	if count > 0 {
		// 更新现有记录
		return s.updateExpenseBillInTx(tx, bill)
	} else {
		// 创建新记录
		return s.createExpenseBillInTx(tx, bill)
	}
}

// createExpenseBillInTx 在事务中创建账单
func (s *DatabaseService) createExpenseBillInTx(tx *sql.Tx, bill *models.ExpenseBill) error {
	query := `
		INSERT INTO expense_bills (
			charge_name, charge_type, model_name, use_group_name, group_name,
			discount_rate, cost_rate, cash_cost, billing_no, order_time,
			use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			trial_cash_cost, transaction_time, time_window_start, time_window_end,
			time_window, create_time,

			-- 模型信息字段
			api_key, model_code, model_product_type, model_product_subtype, model_product_code, model_product_name,

			-- 支付和成本信息字段
			payment_type, start_time, end_time, business_id, cost_price, cost_unit, usage_count, usage_exempt, usage_unit, currency,

			-- 金额信息字段
			settlement_amount, gift_deduct_amount, due_amount, paid_amount, unpaid_amount, billing_status, invoicing_amount, invoiced_amount,

			-- Token业务字段
			token_account_id, token_resource_no, token_resource_name, deduct_usage, deduct_after, token_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tx.Exec(query,
		// 原有字段
		bill.ChargeName, bill.ChargeType, bill.ModelName, bill.UseGroupName, bill.GroupName,
		bill.DiscountRate, bill.CostRate, bill.CashCost, bill.BillingNo, bill.OrderTime,
		bill.UseGroupID, bill.GroupID, bill.ChargeUnit, bill.ChargeCount, bill.ChargeUnitSymbol,
		bill.TrialCashCost, bill.TransactionTime, bill.TimeWindowStart, bill.TimeWindowEnd,
		bill.TimeWindow, bill.CreateTime,

		// 模型信息字段
		bill.APIKey, bill.ModelCode, bill.ModelProductType, bill.ModelProductSubtype, bill.ModelProductCode, bill.ModelProductName,

		// 支付和成本信息字段
		bill.PaymentType, bill.StartTime, bill.EndTime, bill.BusinessID, bill.CostPrice, bill.CostUnit, bill.UsageCount, bill.UsageExempt, bill.UsageUnit, bill.Currency,

		// 金额信息字段
		bill.SettlementAmount, bill.GiftDeductAmount, bill.DueAmount, bill.PaidAmount, bill.UnpaidAmount, bill.BillingStatus, bill.InvoicingAmount, bill.InvoicedAmount,

		// Token业务字段
		bill.TokenAccountID, bill.TokenResourceNo, bill.TokenResourceName, bill.DeductUsage, bill.DeductAfter, bill.TokenType,
	)

	if err != nil {
		return fmt.Errorf("failed to create expense bill in transaction: %w", err)
	}

	return nil
}

// updateExpenseBillInTx 在事务中更新账单
func (s *DatabaseService) updateExpenseBillInTx(tx *sql.Tx, bill *models.ExpenseBill) error {
	query := `
		UPDATE expense_bills SET
			charge_name = ?, charge_type = ?, model_name = ?, use_group_name = ?, group_name = ?,
			discount_rate = ?, cost_rate = ?, cash_cost = ?, order_time = ?,
			use_group_id = ?, group_id = ?, charge_unit = ?, charge_count = ?, charge_unit_symbol = ?,
			trial_cash_cost = ?, transaction_time = ?, time_window_start = ?, time_window_end = ?,
			time_window = ?,

			-- 模型信息字段
			api_key = ?, model_code = ?, model_product_type = ?, model_product_subtype = ?, model_product_code = ?, model_product_name = ?,

			-- 支付和成本信息字段
			payment_type = ?, start_time = ?, end_time = ?, business_id = ?, cost_price = ?, cost_unit = ?, usage_count = ?, usage_exempt = ?, usage_unit = ?, currency = ?,

			-- 金额信息字段
			settlement_amount = ?, gift_deduct_amount = ?, due_amount = ?, paid_amount = ?, unpaid_amount = ?, billing_status = ?, invoicing_amount = ?, invoiced_amount = ?,

			-- Token业务字段
			token_account_id = ?, token_resource_no = ?, token_resource_name = ?, deduct_usage = ?, deduct_after = ?, token_type = ?
		WHERE billing_no = ?
	`

	_, err := tx.Exec(query,
		// 原有字段
		bill.ChargeName, bill.ChargeType, bill.ModelName, bill.UseGroupName, bill.GroupName,
		bill.DiscountRate, bill.CostRate, bill.CashCost, bill.OrderTime,
		bill.UseGroupID, bill.GroupID, bill.ChargeUnit, bill.ChargeCount, bill.ChargeUnitSymbol,
		bill.TrialCashCost, bill.TransactionTime, bill.TimeWindowStart, bill.TimeWindowEnd,
		bill.TimeWindow,

		// 模型信息字段
		bill.APIKey, bill.ModelCode, bill.ModelProductType, bill.ModelProductSubtype, bill.ModelProductCode, bill.ModelProductName,

		// 支付和成本信息字段
		bill.PaymentType, bill.StartTime, bill.EndTime, bill.BusinessID, bill.CostPrice, bill.CostUnit, bill.UsageCount, bill.UsageExempt, bill.UsageUnit, bill.Currency,

		// 金额信息字段
		bill.SettlementAmount, bill.GiftDeductAmount, bill.DueAmount, bill.PaidAmount, bill.UnpaidAmount, bill.BillingStatus, bill.InvoicingAmount, bill.InvoicedAmount,

		// Token业务字段
		bill.TokenAccountID, bill.TokenResourceNo, bill.TokenResourceName, bill.DeductUsage, bill.DeductAfter, bill.TokenType,

		// WHERE 条件
		bill.BillingNo,
	)

	if err != nil {
		return fmt.Errorf("failed to update expense bill in transaction: %w", err)
	}

	return nil
}

// ========== 会员等级智能匹配功能 ==========

// GetCurrentMembershipTier 获取当前用户的会员等级
func (s *DatabaseService) GetCurrentMembershipTier() (string, error) {
	db := s.db

	// 从expense_bills表获取最新的token_resource_name
	query := `
		SELECT token_resource_name 
		FROM expense_bills 
		WHERE token_resource_name IS NOT NULL 
		  AND token_resource_name != ''
		ORDER BY transaction_time DESC 
		LIMIT 1
	`

	var tokenResourceName string
	err := db.QueryRow(query).Scan(&tokenResourceName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "free", nil // 默认为免费版
		}
		return "", fmt.Errorf("failed to get current membership tier: %w", err)
	}

	// 智能匹配会员等级
	tier := s.matchMembershipTier(tokenResourceName)
	return tier, nil
}

// matchMembershipTier 智能匹配会员等级
func (s *DatabaseService) matchMembershipTier(tokenResourceName string) string {
	// 转换为小写便于匹配
	name := strings.ToLower(tokenResourceName)

	// 匹配规则
	if strings.Contains(name, "lite") {
		return "lite"
	} else if strings.Contains(name, "plus") {
		return "plus"
	} else if strings.Contains(name, "pro") {
		return "pro"
	} else if strings.Contains(name, "enterprise") || strings.Contains(name, "企业") {
		return "enterprise"
	} else if strings.Contains(name, "free") || strings.Contains(name, "免费") {
		return "free"
	}

	// 默认返回pro（如果无法匹配）
	return "pro"
}

// ========== MUTATION_01: 清理旧同步历史功能缺失 ==========

// CleanOldSyncHistory 清理指定天数前的同步历史记录 (MUTATION_01)
func (s *DatabaseService) CleanOldSyncHistory(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	query := "DELETE FROM sync_history WHERE start_time < ?"
	_, err := s.db.Exec(query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to clean old sync history: %w", err)
	}

	log.Printf("Cleaned sync history older than %d days", days)
	return nil
}

// ========== AutoSyncConfig Operations ==========

// GetAutoSyncConfigRecord retrieves the auto sync configuration record
func (s *DatabaseService) GetAutoSyncConfigRecord() (*models.AutoSyncConfig, error) {
	query := `
		SELECT id, enabled, frequency_seconds, last_sync_time, next_sync_time,
			   sync_type, billing_month, max_retries, retry_delay,
			   created_at, updated_at
		FROM auto_sync_config
		ORDER BY id DESC
		LIMIT 1
	`

	var config models.AutoSyncConfig
	err := s.db.QueryRow(query).Scan(
		&config.ID, &config.Enabled, &config.FrequencySeconds, &config.LastSyncTime, &config.NextSyncTime,
		&config.SyncType, &config.BillingMonth, &config.MaxRetries, &config.RetryDelay,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 返回默认配置
			return &models.AutoSyncConfig{
				Enabled:          false,
				FrequencySeconds: 3600,
				SyncType:         "full",
				MaxRetries:       3,
				RetryDelay:       60,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get auto sync config: %w", err)
	}

	return &config, nil
}

// SaveAutoSyncConfigRecord saves the auto sync configuration record
func (s *DatabaseService) SaveAutoSyncConfigRecord(config *models.AutoSyncConfig) error {
	query := `
		INSERT OR REPLACE INTO auto_sync_config (
			id, enabled, frequency_seconds, last_sync_time, next_sync_time,
			sync_type, billing_month, max_retries, retry_delay,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		config.ID, config.Enabled, config.FrequencySeconds, config.LastSyncTime, config.NextSyncTime,
		config.SyncType, config.BillingMonth, config.MaxRetries, config.RetryDelay,
		config.CreatedAt, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save auto sync config: %w", err)
	}

	return nil
}

// UpdateAutoSyncLastSyncTime updates the last sync time
func (s *DatabaseService) UpdateAutoSyncLastSyncTime(syncTime time.Time) error {
	query := `
		UPDATE auto_sync_config
		SET last_sync_time = ?, updated_at = ?
		WHERE id = (SELECT id FROM auto_sync_config ORDER BY id DESC LIMIT 1)
	`

	_, err := s.db.Exec(query, syncTime, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last sync time: %w", err)
	}

	return nil
}

// ========== MUTATION_02: 清空账单数据功能缺失 ==========

// DeleteAllExpenseBills 清空所有账单数据 (MUTATION_02)
func (s *DatabaseService) DeleteAllExpenseBills() error {
	query := "DELETE FROM expense_bills"
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all expense bills: %w", err)
	}

	log.Printf("Deleted all expense bills data")
	return nil
}

// ========== SyncHistory Operations ==========

// SaveSyncHistory saves a sync history record
func (s *DatabaseService) SaveSyncHistory(history *models.SyncHistory) error {
	query := `
		INSERT INTO sync_history (
			sync_type, start_time, end_time, status, records_synced,
			error_message, total_records, page_synced, total_pages,
			billing_month, failed_count, sync_time, duration, message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		history.SyncType, history.StartTime, history.EndTime, history.Status, history.RecordsSynced,
		history.ErrorMessage, history.TotalRecords, history.PageSynced, history.TotalPages,
		history.BillingMonth, history.FailedCount, history.SyncTime, history.Duration, history.Message,
	)

	if err != nil {
		return fmt.Errorf("failed to save sync history: %w", err)
	}

	return nil
}
