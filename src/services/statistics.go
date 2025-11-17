package services

import (
	"database/sql"
	"fmt"
	"glm-usage-monitor/models"
	"time"
)

// StatisticsService provides statistical analysis operations
type StatisticsService struct {
	db *sql.DB
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(db *sql.DB) *StatisticsService {
	return &StatisticsService{db: db}
}

// GetOverallStats retrieves overall usage statistics
func (s *StatisticsService) GetOverallStats(startDate, endDate *time.Time) (*models.StatsResponse, error) {
	stats := &models.StatsResponse{}

	// Get total records and cash cost
	whereClause := "1=1"
	args := []interface{}{}
	argIndex := 1

	if startDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) >= DATE(?%d)", argIndex)
		args = append(args, startDate.Format("2006-01-02"))
		argIndex++
	}

	if endDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) <= DATE(?%d)", argIndex)
		args = append(args, endDate.Format("2006-01-02"))
		argIndex++
	}

	// Total records and cash cost
	query := fmt.Sprintf(`
		SELECT COUNT(*), COALESCE(SUM(cash_cost), 0), COALESCE(SUM(charge_unit), 0)
		FROM expense_bills WHERE %s
	`, whereClause)

	err := s.db.QueryRow(query, args...).Scan(&stats.TotalRecords, &stats.TotalCashCost, new(float64))
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	// Get hourly usage data
	hourlyUsage, err := s.GetHourlyUsage(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly usage: %w", err)
	}
	stats.HourlyUsage = hourlyUsage

	// Get model distribution
	modelDist, err := s.GetModelDistribution(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get model distribution: %w", err)
	}
	stats.ModelDistribution = modelDist

	// Get charge type statistics
	chargeStats, err := s.GetChargeTypeStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get charge type stats: %w", err)
	}
	stats.ChargeTypeStats = chargeStats

	// Get recent usage
	recentUsage, err := s.GetRecentUsage(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent usage: %w", err)
	}
	stats.RecentUsage = recentUsage

	return stats, nil
}

// GetHourlyUsage retrieves hourly usage statistics for the last 5 hours
func (s *StatisticsService) GetHourlyUsage(startDate, endDate *time.Time) ([]models.HourlyUsageData, error) {
	whereClause := "1=1"
	args := []interface{}{}
	argIndex := 1

	if startDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) >= DATE(?%d)", argIndex)
		args = append(args, startDate.Format("2006-01-02"))
		argIndex++
	}

	if endDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) <= DATE(?%d)", argIndex)
		args = append(args, endDate.Format("2006-01-02"))
		argIndex++
	}

	// If no specific date range, get last 5 hours
	if startDate == nil && endDate == nil {
		whereClause += fmt.Sprintf(" AND transaction_time >= DATETIME('now', '-5 hours')")
	}

	query := fmt.Sprintf(`
		SELECT
			CASE
				WHEN transaction_time >= DATETIME('now', '-5 hours') AND transaction_time < DATETIME('now', '-4 hours') THEN -4
				WHEN transaction_time >= DATETIME('now', '-4 hours') AND transaction_time < DATETIME('now', '-3 hours') THEN -3
				WHEN transaction_time >= DATETIME('now', '-3 hours') AND transaction_time < DATETIME('now', '-2 hours') THEN -2
				WHEN transaction_time >= DATETIME('now', '-2 hours') AND transaction_time < DATETIME('now', '-1 hours') THEN -1
				WHEN transaction_time >= DATETIME('now', '-1 hours') THEN 0
				ELSE strftime('%H', transaction_time)
			END as hour,
			COUNT(*) as call_count,
			COALESCE(SUM(charge_unit), 0) as token_usage,
			COALESCE(SUM(cash_cost), 0) as cash_cost
		FROM expense_bills
		WHERE %s
		GROUP BY hour
		ORDER BY hour
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query hourly usage: %w", err)
	}
	defer rows.Close()

	var hourlyData []models.HourlyUsageData
	for rows.Next() {
		var data models.HourlyUsageData
		err := rows.Scan(&data.Hour, &data.CallCount, &data.TokenUsage, &data.CashCost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hourly usage: %w", err)
		}
		hourlyData = append(hourlyData, data)
	}

	return hourlyData, nil
}

// GetModelDistribution retrieves usage distribution by model
func (s *StatisticsService) GetModelDistribution(startDate, endDate *time.Time) ([]models.ModelDistributionData, error) {
	whereClause := "1=1"
	args := []interface{}{}
	argIndex := 1

	if startDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) >= DATE(?%d)", argIndex)
		args = append(args, startDate.Format("2006-01-02"))
		argIndex++
	}

	if endDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) <= DATE(?%d)", argIndex)
		args = append(args, endDate.Format("2006-01-02"))
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT
			model_name,
			COUNT(*) as call_count,
			COALESCE(SUM(charge_unit), 0) as token_usage,
			COALESCE(SUM(cash_cost), 0) as cash_cost
		FROM expense_bills
		WHERE %s AND model_name IS NOT NULL AND model_name != ''
		GROUP BY model_name
		ORDER BY cash_cost DESC
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query model distribution: %w", err)
	}
	defer rows.Close()

	var modelData []models.ModelDistributionData
	var totalCashCost float64

	// First pass: collect data and total
	for rows.Next() {
		var data models.ModelDistributionData
		err := rows.Scan(&data.ModelName, &data.CallCount, &data.TokenUsage, &data.CashCost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model distribution: %w", err)
		}
		modelData = append(modelData, data)
		totalCashCost += data.CashCost
	}

	// Calculate percentages
	for i := range modelData {
		if totalCashCost > 0 {
			modelData[i].Percentage = (modelData[i].CashCost / totalCashCost) * 100
		}
	}

	return modelData, nil
}

// GetChargeTypeStats retrieves statistics by charge type
func (s *StatisticsService) GetChargeTypeStats(startDate, endDate *time.Time) ([]models.ChargeTypeStatsData, error) {
	whereClause := "1=1"
	args := []interface{}{}
	argIndex := 1

	if startDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) >= DATE(?%d)", argIndex)
		args = append(args, startDate.Format("2006-01-02"))
		argIndex++
	}

	if endDate != nil {
		whereClause += fmt.Sprintf(" AND DATE(transaction_time) <= DATE(?%d)", argIndex)
		args = append(args, endDate.Format("2006-01-02"))
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT
			charge_type,
			COUNT(*) as call_count,
			COALESCE(SUM(cash_cost), 0) as cash_cost
		FROM expense_bills
		WHERE %s AND charge_type IS NOT NULL AND charge_type != ''
		GROUP BY charge_type
		ORDER BY cash_cost DESC
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query charge type stats: %w", err)
	}
	defer rows.Close()

	var chargeData []models.ChargeTypeStatsData
	var totalCashCost float64

	// First pass: collect data and total
	for rows.Next() {
		var data models.ChargeTypeStatsData
		err := rows.Scan(&data.ChargeType, &data.CallCount, &data.CashCost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan charge type stats: %w", err)
		}
		chargeData = append(chargeData, data)
		totalCashCost += data.CashCost
	}

	// Calculate percentages
	for i := range chargeData {
		if totalCashCost > 0 {
			chargeData[i].Percentage = (chargeData[i].CashCost / totalCashCost) * 100
		}
	}

	return chargeData, nil
}

// GetRecentUsage retrieves recent usage records
func (s *StatisticsService) GetRecentUsage(limit int) ([]models.ExpenseBill, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, charge_name, charge_type, model_name, use_group_name, group_name,
			   discount_rate, cost_rate, cash_cost, billing_no, order_time,
			   use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			   trial_cash_cost, transaction_time, time_window_start, time_window_end,
			   time_window, create_time
		FROM expense_bills
		ORDER BY transaction_time DESC, create_time DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent usage: %w", err)
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
			return nil, fmt.Errorf("failed to scan recent usage: %w", err)
		}
		bills = append(bills, bill)
	}

	return bills, nil
}

// GetUsageTrend retrieves usage trend data for the specified period
func (s *StatisticsService) GetUsageTrend(days int) ([]models.HourlyUsageData, error) {
	if days <= 0 {
		days = 7
	}

	query := `
		SELECT
			DATE(transaction_time) as date,
			COUNT(*) as call_count,
			COALESCE(SUM(charge_unit), 0) as token_usage,
			COALESCE(SUM(cash_cost), 0) as cash_cost
		FROM expense_bills
		WHERE transaction_time >= DATE('now', '-%d days')
		GROUP BY DATE(transaction_time)
		ORDER BY date ASC
	`

	rows, err := s.db.Query(query, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage trend: %w", err)
	}
	defer rows.Close()

	var trendData []models.HourlyUsageData
	for rows.Next() {
		var data models.HourlyUsageData
		var dateStr string
		err := rows.Scan(&dateStr, &data.CallCount, &data.TokenUsage, &data.CashCost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage trend: %w", err)
		}

		// Convert date string to hour representation for simplicity
		// In a real implementation, you might want a different structure
		trendData = append(trendData, data)
	}

	return trendData, nil
}

// GetTopExpenses retrieves top expenses by amount
func (s *StatisticsService) GetTopExpenses(limit int) ([]models.ExpenseBill, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, charge_name, charge_type, model_name, use_group_name, group_name,
			   discount_rate, cost_rate, cash_cost, billing_no, order_time,
			   use_group_id, group_id, charge_unit, charge_count, charge_unit_symbol,
			   trial_cash_cost, transaction_time, time_window_start, time_window_end,
			   time_window, create_time
		FROM expense_bills
		WHERE cash_cost > 0
		ORDER BY cash_cost DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top expenses: %w", err)
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
			return nil, fmt.Errorf("failed to scan top expenses: %w", err)
		}
		bills = append(bills, bill)
	}

	return bills, nil
}