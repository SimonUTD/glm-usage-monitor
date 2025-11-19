package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TimeWindow represents parsed time window data
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ParseTimeWindow parses the time window string from the API response
// Format: "2025-11-01 00:00:00 - 2025-11-01 23:59:59"
func ParseTimeWindow(timeWindowStr string) (*TimeWindow, error) {
	if timeWindowStr == "" {
		return nil, nil
	}

	// Split by " - " to get start and end times
	parts := strings.Split(timeWindowStr, " - ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid time window format: %s", timeWindowStr)
	}

	startTimeStr := strings.TrimSpace(parts[0])
	endTimeStr := strings.TrimSpace(parts[1])

	// Parse start time
	startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time %s: %w", startTimeStr, err)
	}

	// Parse end time
	endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time %s: %w", endTimeStr, err)
	}

	return &TimeWindow{
		Start: startTime,
		End:   endTime,
	}, nil
}

// ExtractTransactionTime extracts timestamp from billingNo field
// Format: "customerId" + 13-digit timestamp (milliseconds)
// Example: "example_customer_1731781234567"
func ExtractTransactionTime(billingNo string) (time.Time, error) {
	if billingNo == "" {
		return time.Time{}, fmt.Errorf("billing no is empty")
	}

	// Look for 13-digit timestamp at the end of the string
	// Pattern: any characters followed by 13 digits
	re := regexp.MustCompile(`(\d{13})$`)
	matches := re.FindStringSubmatch(billingNo)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("no valid timestamp found in billing no: %s", billingNo)
	}

	timestampStr := matches[1]
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp %s: %w", timestampStr, err)
	}

	// Convert milliseconds timestamp to time.Time
	return time.Unix(timestamp/1000, (timestamp%1000)*1000000), nil
}

// TransformExpenseBill transforms raw expense bill data into our database model
func TransformExpenseBill(rawBill map[string]interface{}) (*ExpenseBill, error) {
	bill := &ExpenseBill{}

	// Basic string fields
	if v, ok := rawBill["charge_name"].(string); ok {
		bill.ChargeName = v
	}
	if v, ok := rawBill["charge_type"].(string); ok {
		bill.ChargeType = v
	}
	if v, ok := rawBill["model_name"].(string); ok {
		bill.ModelName = v
	}
	if v, ok := rawBill["use_group_name"].(string); ok {
		bill.UseGroupName = v
	}
	if v, ok := rawBill["group_name"].(string); ok {
		bill.GroupName = v
	}
	if v, ok := rawBill["billing_no"].(string); ok {
		bill.BillingNo = v
	}
	if v, ok := rawBill["order_time"].(string); ok {
		bill.OrderTime = v
	}
	if v, ok := rawBill["use_group_id"].(string); ok {
		bill.UseGroupID = v
	}
	if v, ok := rawBill["group_id"].(string); ok {
		bill.GroupID = v
	}
	if v, ok := rawBill["charge_unit_symbol"].(string); ok {
		bill.ChargeUnitSymbol = v
	}

	// Numeric fields
	if v, ok := rawBill["discount_rate"].(float64); ok {
		bill.DiscountRate = v
	} else if v, ok := rawBill["discount_rate"].(string); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			bill.DiscountRate = f
		}
	}

	if v, ok := rawBill["cost_rate"].(float64); ok {
		bill.CostRate = v
	} else if v, ok := rawBill["cost_rate"].(string); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			bill.CostRate = f
		}
	}

	if v, ok := rawBill["cash_cost"].(float64); ok {
		bill.CashCost = v
	} else if v, ok := rawBill["cash_cost"].(string); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			bill.CashCost = f
		}
	}

	if v, ok := rawBill["charge_unit"].(float64); ok {
		bill.ChargeUnit = v
	} else if v, ok := rawBill["charge_unit"].(string); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			bill.ChargeUnit = f
		}
	}

	if v, ok := rawBill["charge_count"].(float64); ok {
		bill.ChargeCount = v
	} else if v, ok := rawBill["charge_count"].(string); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			bill.ChargeCount = f
		}
	}

	if v, ok := rawBill["trial_cash_cost"].(float64); ok {
		bill.TrialCashCost = v
	} else if v, ok := rawBill["trial_cash_cost"].(string); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			bill.TrialCashCost = f
		}
	}

	// Time window processing
	if v, ok := rawBill["time_window"].(string); ok && v != "" {
		bill.TimeWindow = v
		if timeWindow, err := ParseTimeWindow(v); err == nil {
			bill.TimeWindowStart = timeWindow.Start
			bill.TimeWindowEnd = timeWindow.End
		}
	}

	// Transaction time extraction from billing_no
	if bill.BillingNo != "" {
		if transactionTime, err := ExtractTransactionTime(bill.BillingNo); err == nil {
			bill.TransactionTime = transactionTime
		}
	}

	// Set creation time
	bill.CreateTime = time.Now()

	return bill, nil
}

// ValidateExpenseBill validates the expense bill data
func ValidateExpenseBill(bill *ExpenseBill) error {
	if bill.BillingNo == "" {
		return fmt.Errorf("billing no is required")
	}

	if bill.TransactionTime.IsZero() {
		return fmt.Errorf("transaction time is required")
	}

	if bill.CashCost < 0 {
		return fmt.Errorf("cash cost cannot be negative")
	}

	if bill.ChargeUnit < 0 {
		return fmt.Errorf("charge unit cannot be negative")
	}

	return nil
}

// ConvertToDatabaseFriendly converts the expense bill to database-friendly format
func (bill *ExpenseBill) ConvertToDatabaseFriendly() interface{} {
	// Return a map with database-friendly field names and types
	return map[string]interface{}{
		"charge_name":        bill.ChargeName,
		"charge_type":        bill.ChargeType,
		"model_name":         bill.ModelName,
		"use_group_name":     bill.UseGroupName,
		"group_name":         bill.GroupName,
		"discount_rate":      bill.DiscountRate,
		"cost_rate":          bill.CostRate,
		"cash_cost":          bill.CashCost,
		"billing_no":         bill.BillingNo,
		"order_time":         bill.OrderTime,
		"use_group_id":       bill.UseGroupID,
		"group_id":           bill.GroupID,
		"charge_unit":        bill.ChargeUnit,
		"charge_count":       bill.ChargeCount,
		"charge_unit_symbol": bill.ChargeUnitSymbol,
		"trial_cash_cost":    bill.TrialCashCost,
		"transaction_time":   bill.TransactionTime.Format("2006-01-02 15:04:05"),
		"time_window_start":  bill.TimeWindowStart.Format("2006-01-02 15:04:05"),
		"time_window_end":    bill.TimeWindowEnd.Format("2006-01-02 15:04:05"),
		"time_window":        bill.TimeWindow,
		"create_time":        bill.CreateTime.Format("2006-01-02 15:04:05"),
	}
}

// FormatCashCost formats cash cost to a readable string
func FormatCashCost(cost float64) string {
	return fmt.Sprintf("¥%.4f", cost)
}

// FormatChargeUnit formats charge unit with its symbol
func FormatChargeUnit(unit float64, symbol string) string {
	if symbol == "" {
		symbol = "tokens"
	}
	return fmt.Sprintf("%.2f %s", unit, symbol)
}

// MIDDLEWARE_01: 统一时间格式化函数
// FormatDateTime 统一日期时间格式化为YYYY-MM-DD HH:mm:ss
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 统一日期格式化为YYYY-MM-DD
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDateTimeForAPI 为API响应格式化日期时间
func FormatDateTimeForAPI(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDateForAPI 为API响应格式化日期
func FormatDateForAPI(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatTimeForAPI 为API响应格式化时间
func FormatTimeForAPI(t time.Time) string {
	return t.Format("15:04:05")
}
