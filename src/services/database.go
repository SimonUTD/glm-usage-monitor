package services

import (
	"database/sql"
	"fmt"
	"log"
	"glm-usage-monitor/models"
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

// ========== ExpenseBill Operations ==========

// CreateExpenseBill creates a new expense bill record
func (s *DatabaseService) CreateExpenseBill(bill *models.ExpenseBill) error {
	query := `
		INSERT INTO expense_bills (
			charge_name, charge_type, model_name, use_group_name, group_name,
			discount_rate, cost_rate, cash_cost, billing_no, order_time,
			use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			trial_cash_cost, transaction_time, time_window_start, time_window_end,
			time_window, create_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		bill.ChargeName, bill.ChargeType, bill.ModelName, bill.UseGroupName, bill.GroupName,
		bill.DiscountRate, bill.CostRate, bill.CashCost, bill.BillingNo, bill.OrderTime,
		bill.UseGroupID, bill.GroupID, bill.ChargeUnit, bill.ChargeCount, bill.ChargeUnitSymbol,
		bill.TrialCashCost, bill.TransactionTime, bill.TimeWindowStart, bill.TimeWindowEnd,
		bill.TimeWindow, bill.CreateTime,
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
			time_window, create_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			bill.ChargeName, bill.ChargeType, bill.ModelName, bill.UseGroupName, bill.GroupName,
			bill.DiscountRate, bill.CostRate, bill.CashCost, bill.BillingNo, bill.OrderTime,
			bill.UseGroupID, bill.GroupID, bill.ChargeUnit, bill.ChargeCount, bill.ChargeUnitSymbol,
			bill.TrialCashCost, bill.TransactionTime, bill.TimeWindowStart, bill.TimeWindowEnd,
			bill.TimeWindow, bill.CreateTime,
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
	argIndex := 1

	// Build WHERE conditions
	if filter.StartDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("DATE(transaction_time) >= DATE(?%d)", argIndex))
		args = append(args, filter.StartDate.Format("2006-01-02"))
		argIndex++
	}

	if filter.EndDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("DATE(transaction_time) <= DATE(?%d)", argIndex))
		args = append(args, filter.EndDate.Format("2006-01-02"))
		argIndex++
	}

	if filter.ModelName != nil && *filter.ModelName != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("model_name = ?%d", argIndex))
		args = append(args, *filter.ModelName)
		argIndex++
	}

	if filter.ChargeType != nil && *filter.ChargeType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("charge_type = ?%d", argIndex))
		args = append(args, *filter.ChargeType)
		argIndex++
	}

	if filter.GroupName != nil && *filter.GroupName != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("group_name LIKE ?%d", argIndex))
		args = append(args, "%"+*filter.GroupName+"%")
		argIndex++
	}

	if filter.MinCashCost != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cash_cost >= ?%d", argIndex))
		args = append(args, *filter.MinCashCost)
		argIndex++
	}

	if filter.MaxCashCost != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cash_cost <= ?%d", argIndex))
		args = append(args, *filter.MaxCashCost)
		argIndex++
	}

	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(charge_name LIKE ?%d OR model_name LIKE ?%d OR billing_no LIKE ?%d)", argIndex, argIndex+1, argIndex+2))
		searchTerm := "%" + *filter.SearchTerm + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		argIndex += 3
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
			   time_window, create_time
		FROM expense_bills
		WHERE %s
		ORDER BY transaction_time DESC
		LIMIT ?%d OFFSET ?%d
	`, whereClause, argIndex, argIndex+1)

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
func (s *DatabaseService) GetExpenseBillByID(id int) (*models.ExpenseBill, error) {
	query := `
		SELECT id, charge_name, charge_type, model_name, use_group_name, group_name,
			   discount_rate, cost_rate, cash_cost, billing_no, order_time,
			   use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			   trial_cash_cost, transaction_time, time_window_start, time_window_end,
			   time_window, create_time
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
func (s *DatabaseService) DeleteExpenseBill(id int) error {
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
			   time_window, create_time
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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense bill: %w", err)
		}
		bills = append(bills, bill)
	}

	return bills, nil
}

// ========== APIToken Operations ==========

// SaveAPIToken saves an API token
func (s *DatabaseService) SaveAPIToken(token *models.APIToken) error {
	query := `
		INSERT INTO api_tokens (token_name, token_value, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(token_name) DO UPDATE SET
			token_value = excluded.token_value,
			is_active = excluded.is_active,
			updated_at = excluded.updated_at
	`

	_, err := s.db.Exec(query, token.TokenName, token.TokenValue, token.IsActive, token.CreatedAt, token.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save API token: %w", err)
	}

	return nil
}

// GetActiveAPIToken retrieves the active API token
func (s *DatabaseService) GetActiveAPIToken() (*models.APIToken, error) {
	query := `
		SELECT id, token_name, token_value, is_active, created_at, updated_at
		FROM api_tokens
		WHERE is_active = 1
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var token models.APIToken
	err := s.db.QueryRow(query).Scan(&token.ID, &token.TokenName, &token.TokenValue, &token.IsActive, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active API token found")
		}
		return nil, fmt.Errorf("failed to get active API token: %w", err)
	}

	return &token, nil
}

// GetAllAPITokens retrieves all API tokens
func (s *DatabaseService) GetAllAPITokens() ([]models.APIToken, error) {
	query := `
		SELECT id, token_name, token_value, is_active, created_at, updated_at
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
		err := rows.Scan(&token.ID, &token.TokenName, &token.TokenValue, &token.IsActive, &token.CreatedAt, &token.UpdatedAt)
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