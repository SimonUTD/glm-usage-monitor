package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Database holds the database connection and related information
type Database struct {
	DB     *sql.DB
	dbPath string
}

// SQLDatabase alias for type compatibility
type SQLDatabase = Database

// NewDatabase creates a new database instance with proper cross-platform path configuration
func NewDatabase() (*Database, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		DB:     db,
		dbPath: dbPath,
	}

	// Initialize database schema
	log.Printf("DEBUG: Initializing database schema...")
	if err := database.initSchema(); err != nil {
		log.Printf("DEBUG: Database schema initialization failed: %v", err)
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}
	log.Printf("DEBUG: Database schema initialized successfully")

	return database, nil
}

// getDatabasePath returns the appropriate database path for the current operating system
func getDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".glm-usage-monitor")

	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "expense_bills.db"), nil
}

// initSchema creates all necessary database tables
func (db *Database) initSchema() error {
	// First, create tables with all fields including new ones
	if err := db.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Run database migrations to handle schema changes
	if err := RunMigrations(db.DB); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Then, migrate existing data by adding new columns if they don't exist
	if err := db.migrateExistingTables(); err != nil {
		return fmt.Errorf("failed to migrate existing tables: %w", err)
	}

	// Create indexes for better performance
	if err := db.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Insert default configuration values
	if err := db.insertDefaultConfigs(); err != nil {
		return fmt.Errorf("failed to insert default configs: %w", err)
	}

	return nil
}

// createTables creates the database tables
func (db *Database) createTables() error {
	schemas := []string{
		// expense_bills table - main table for storing GLM billing data
		`CREATE TABLE IF NOT EXISTS expense_bills (
			id TEXT PRIMARY KEY,
			charge_name TEXT,
			charge_type TEXT,
			model_name TEXT,
			use_group_name TEXT,
			group_name TEXT,
			discount_rate REAL,
			cost_rate REAL,
			cash_cost REAL,
			billing_no TEXT,
			order_time TEXT,
			use_group_id TEXT,
			group_id TEXT,
			charge_unit REAL,
			charge_count REAL,
			charge_unit_symbol TEXT,
			trial_cash_cost REAL,
			transaction_time DATETIME,
			time_window_start DATETIME,
			time_window_end DATETIME,
			time_window TEXT,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			
			-- === 模型信息字段 ===
			api_key TEXT,
			model_code TEXT,
			model_product_type TEXT,
			model_product_subtype TEXT,
			model_product_code TEXT,
			model_product_name TEXT,
			
			-- === 支付和成本信息字段 ===
			payment_type TEXT,
			start_time TEXT,
			end_time TEXT,
			business_id TEXT,
			cost_price REAL,
			cost_unit TEXT,
			usage_count REAL,
			usage_exempt REAL,
			usage_unit TEXT,
			currency TEXT DEFAULT 'CNY',
			
			-- === 金额信息字段 ===
			settlement_amount REAL,
			gift_deduct_amount REAL DEFAULT 0,
			due_amount REAL,
			paid_amount REAL DEFAULT 0,
			unpaid_amount REAL DEFAULT 0,
			billing_status TEXT DEFAULT 'unpaid',
			invoicing_amount REAL DEFAULT 0,
			invoiced_amount REAL DEFAULT 0,
			
			-- === Token业务字段 ===
			token_account_id TEXT,
			token_resource_no TEXT,
			token_resource_name TEXT,
			deduct_usage REAL DEFAULT 0,
			deduct_after TEXT,
			token_type TEXT
		)`,

		// api_tokens table - for storing API tokens (DB_02: 重新设计)
		`CREATE TABLE IF NOT EXISTS api_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_name TEXT NOT NULL,
			token_value TEXT NOT NULL,
			provider TEXT,
			token_type TEXT,
			is_active INTEGER DEFAULT 1,
			daily_limit INTEGER,
			monthly_limit INTEGER,
			expires_at DATETIME,
			last_used_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// sync_history table - for tracking synchronization history
		`CREATE TABLE IF NOT EXISTS sync_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sync_type TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			status TEXT NOT NULL,
			records_synced INTEGER DEFAULT 0,
			error_message TEXT,
			total_records INTEGER DEFAULT 0,
			page_synced INTEGER DEFAULT 0,
			total_pages INTEGER DEFAULT 0
		)`,

		// auto_sync_config table - for storing auto-sync configuration (DB_03: 重新设计)
		`CREATE TABLE IF NOT EXISTS auto_sync_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			enabled INTEGER DEFAULT 0,
			frequency_seconds INTEGER DEFAULT 3600,
			last_sync_time DATETIME,
			next_sync_time DATETIME,
			sync_type TEXT DEFAULT 'full',
			billing_month TEXT,
			max_retries INTEGER DEFAULT 3,
			retry_delay INTEGER DEFAULT 5,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// membership_tier_limits table - for storing membership tier information
		`CREATE TABLE IF NOT EXISTS membership_tier_limits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tier_name TEXT UNIQUE NOT NULL,
			daily_limit INTEGER,
			monthly_limit INTEGER,
			max_tokens INTEGER,
			max_context_length INTEGER,
			features TEXT,
			description TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		_, err := db.DB.Exec(schema)
		if err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

// migrateExistingTables adds new columns to existing tables
func (db *Database) migrateExistingTables() error {
	migrations := []string{
		// === DB_01: 添加缺失的expense_bills字段 ===
		"ALTER TABLE expense_bills ADD COLUMN billing_date TEXT",
		"ALTER TABLE expense_bills ADD COLUMN billing_time TEXT",
		"ALTER TABLE expense_bills ADD COLUMN customer_id TEXT",
		"ALTER TABLE expense_bills ADD COLUMN order_no TEXT",
		"ALTER TABLE expense_bills ADD COLUMN original_amount REAL",
		"ALTER TABLE expense_bills ADD COLUMN original_cost_price REAL",
		"ALTER TABLE expense_bills ADD COLUMN discount_type TEXT",
		"ALTER TABLE expense_bills ADD COLUMN credit_pay_amount REAL",
		"ALTER TABLE expense_bills ADD COLUMN third_party REAL",
		"ALTER TABLE expense_bills ADD COLUMN cash_amount REAL",
		"ALTER TABLE expense_bills ADD COLUMN api_usage INTEGER",

		// === 模型信息字段 ===
		"ALTER TABLE expense_bills ADD COLUMN api_key TEXT",
		"ALTER TABLE expense_bills ADD COLUMN model_code TEXT",
		"ALTER TABLE expense_bills ADD COLUMN model_product_type TEXT",
		"ALTER TABLE expense_bills ADD COLUMN model_product_subtype TEXT",
		"ALTER TABLE expense_bills ADD COLUMN model_product_code TEXT",
		"ALTER TABLE expense_bills ADD COLUMN model_product_name TEXT",

		// === 支付和成本信息字段 ===
		"ALTER TABLE expense_bills ADD COLUMN payment_type TEXT",
		"ALTER TABLE expense_bills ADD COLUMN start_time TEXT",
		"ALTER TABLE expense_bills ADD COLUMN end_time TEXT",
		"ALTER TABLE expense_bills ADD COLUMN business_id TEXT",
		"ALTER TABLE expense_bills ADD COLUMN cost_price REAL",
		"ALTER TABLE expense_bills ADD COLUMN cost_unit TEXT",
		"ALTER TABLE expense_bills ADD COLUMN usage_count REAL",
		"ALTER TABLE expense_bills ADD COLUMN usage_exempt REAL",
		"ALTER TABLE expense_bills ADD COLUMN usage_unit TEXT",
		"ALTER TABLE expense_bills ADD COLUMN currency TEXT DEFAULT 'CNY'",

		// === 金额信息字段 ===
		"ALTER TABLE expense_bills ADD COLUMN settlement_amount REAL",
		"ALTER TABLE expense_bills ADD COLUMN gift_deduct_amount REAL DEFAULT 0",
		"ALTER TABLE expense_bills ADD COLUMN due_amount REAL",
		"ALTER TABLE expense_bills ADD COLUMN paid_amount REAL DEFAULT 0",
		"ALTER TABLE expense_bills ADD COLUMN unpaid_amount REAL DEFAULT 0",
		"ALTER TABLE expense_bills ADD COLUMN billing_status TEXT DEFAULT 'unpaid'",
		"ALTER TABLE expense_bills ADD COLUMN invoicing_amount REAL DEFAULT 0",
		"ALTER TABLE expense_bills ADD COLUMN invoiced_amount REAL DEFAULT 0",

		// === Token业务字段 ===
		"ALTER TABLE expense_bills ADD COLUMN token_account_id TEXT",
		"ALTER TABLE expense_bills ADD COLUMN token_resource_no TEXT",
		"ALTER TABLE expense_bills ADD COLUMN token_resource_name TEXT",
		"ALTER TABLE expense_bills ADD COLUMN deduct_usage REAL DEFAULT 0",
		"ALTER TABLE expense_bills ADD COLUMN deduct_after TEXT",
		"ALTER TABLE expense_bills ADD COLUMN token_type TEXT",

		// === DB_02: 为api_tokens表添加缺失字段 ===
		"ALTER TABLE api_tokens ADD COLUMN provider TEXT",
		"ALTER TABLE api_tokens ADD COLUMN token_type TEXT",
		"ALTER TABLE api_tokens ADD COLUMN daily_limit INTEGER",
		"ALTER TABLE api_tokens ADD COLUMN monthly_limit INTEGER",
		"ALTER TABLE api_tokens ADD COLUMN expires_at DATETIME",
		"ALTER TABLE api_tokens ADD COLUMN last_used_at DATETIME",

		// === DB_05: 为membership_tier_limits表添加缺失字段 ===
		"ALTER TABLE membership_tier_limits ADD COLUMN period_hours INTEGER",
		"ALTER TABLE membership_tier_limits ADD COLUMN call_limit INTEGER",
	}

	for _, migration := range migrations {
		_, err := db.DB.Exec(migration)
		if err != nil {
			// Ignore "duplicate column name" errors for existing databases
			if !strings.Contains(err.Error(), "duplicate column name") {
				log.Printf("Migration failed (continuing anyway): %s - %v", migration, err)
			}
		}
	}

	return nil
}

// createIndexes creates database indexes for better performance
func (db *Database) createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_transaction_time ON expense_bills(transaction_time)",
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_billing_no ON expense_bills(billing_no)",
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_model_name ON expense_bills(model_name)",
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_charge_type ON expense_bills(charge_type)",
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_api_key ON expense_bills(api_key)",
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_payment_type ON expense_bills(payment_type)",
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_billing_status ON expense_bills(billing_status)",
		"CREATE INDEX IF NOT EXISTS idx_expense_bills_token_account_id ON expense_bills(token_account_id)",
		"CREATE INDEX IF NOT EXISTS idx_sync_history_start_time ON sync_history(start_time)",
		"CREATE INDEX IF NOT EXISTS idx_api_tokens_is_active ON api_tokens(is_active)",
	}

	for _, index := range indexes {
		_, err := db.DB.Exec(index)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Insert default configuration values
	if err := db.insertDefaultConfigs(); err != nil {
		return fmt.Errorf("failed to insert default configs: %w", err)
	}

	return nil
}

// insertDefaultConfigs inserts default configuration values
func (db *Database) insertDefaultConfigs() error {
	// 检查是否已有配置记录
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM auto_sync_config").Scan(&count)
	if err != nil {
		log.Printf("DEBUG: Failed to check auto_sync_config count: %v", err)
		// 继续执行，可能是表不存在
	}

	// 如果已有记录，跳过插入默认值
	if count > 0 {
		log.Printf("DEBUG: auto_sync_config already has %d records, skipping default insertion", count)
		return nil
	}

	// 插入默认配置记录
	_, err = db.DB.Exec(`
		INSERT INTO auto_sync_config (enabled, frequency_seconds, sync_type, max_retries, retry_delay)
		VALUES (0, 3600, 'full', 3, 60)
	`)
	if err != nil {
		log.Printf("DEBUG: Failed to insert default auto_sync_config: %v", err)
		return fmt.Errorf("failed to insert default auto_sync_config: %w", err)
	}

	log.Printf("DEBUG: Default auto_sync_config inserted successfully")
	return nil
}

// getDefaultConfigDescription returns a description for default configuration keys
func getDefaultConfigDescription(key string) string {
	descriptions := map[string]string{
		"auto_sync_enabled":  "Enable automatic data synchronization",
		"sync_interval":      "Synchronization interval in seconds",
		"last_sync_time":     "Timestamp of last successful synchronization",
		"sync_on_startup":    "Sync data when application starts",
		"max_retry_attempts": "Maximum number of retry attempts for failed syncs",
		"retry_delay":        "Delay between retry attempts in milliseconds",
	}

	if desc, exists := descriptions[key]; exists {
		return desc
	}
	return "Configuration value"
}

// Close closes the database connection
func (db *Database) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// GetDB returns the underlying database connection
func (db *Database) GetDB() *sql.DB {
	return db.DB
}

// GetPath returns the current database file path
func (db *Database) GetPath() string {
	return db.dbPath
}

// GetDatabasePath returns the current database file path (for backward compatibility)
func (db *Database) GetDatabasePath() string {
	return db.dbPath
}
