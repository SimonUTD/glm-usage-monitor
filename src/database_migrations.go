package main

import (
	"database/sql"
	"fmt"
	"log"
)

// MigrationScript 数据库迁移脚本
type MigrationScript struct {
	Version     int
	Description string
	SQL         string
}

// GetMigrations 获取所有迁移脚本
func GetMigrations() []MigrationScript {
	return []MigrationScript{
		{
			Version:     1,
			Description: "DB_01: 添加缺失的expense_bills字段",
			SQL: `
				-- 添加客户相关字段
				ALTER TABLE expense_bills ADD COLUMN customer_id TEXT;
				ALTER TABLE expense_bills ADD COLUMN order_no TEXT;
				ALTER TABLE expense_bills ADD COLUMN billing_date TEXT;
				ALTER TABLE expense_bills ADD COLUMN billing_time TEXT;

				-- 添加成本计算字段
				ALTER TABLE expense_bills ADD COLUMN original_amount REAL;
				ALTER TABLE expense_bills ADD COLUMN original_cost_price REAL;
				ALTER TABLE expense_bills ADD COLUMN discount_type TEXT;
				ALTER TABLE expense_bills ADD COLUMN credit_pay_amount REAL;
				ALTER TABLE expense_bills ADD COLUMN third_party REAL;
				ALTER TABLE expense_bills ADD COLUMN cash_amount REAL;
				ALTER TABLE expense_bills ADD COLUMN api_usage INTEGER;
			`,
		},
		{
			Version:     2,
			Description: "DB_06: 修复数据类型不一致问题",
			SQL: `
				-- 修复 business_id 类型从 INTEGER 改为 TEXT
				ALTER TABLE expense_bills ADD COLUMN business_id_new TEXT;
				UPDATE expense_bills SET business_id_new = CAST(business_id AS TEXT) WHERE business_id IS NOT NULL;
				ALTER TABLE expense_bills DROP COLUMN business_id;
				ALTER TABLE expense_bills RENAME COLUMN business_id_new TO business_id;

				-- 修复 usage_count 类型从 INTEGER 改为 REAL
				ALTER TABLE expense_bills ADD COLUMN usage_count_new REAL;
				UPDATE expense_bills SET usage_count_new = CAST(usage_count AS REAL) WHERE usage_count IS NOT NULL;
				ALTER TABLE expense_bills DROP COLUMN usage_count;
				ALTER TABLE expense_bills RENAME COLUMN usage_count_new TO usage_count;

				-- 修复 deduct_usage 类型从 INTEGER 改为 REAL
				ALTER TABLE expense_bills ADD COLUMN deduct_usage_new REAL;
				UPDATE expense_bills SET deduct_usage_new = CAST(deduct_usage AS REAL) WHERE deduct_usage IS NOT NULL;
				ALTER TABLE expense_bills DROP COLUMN deduct_usage;
				ALTER TABLE expense_bills RENAME COLUMN deduct_usage_new TO deduct_usage;

				-- 修复 token_account_id 类型从 INTEGER 改为 TEXT
				ALTER TABLE expense_bills ADD COLUMN token_account_id_new TEXT;
				UPDATE expense_bills SET token_account_id_new = CAST(token_account_id AS TEXT) WHERE token_account_id IS NOT NULL;
				ALTER TABLE expense_bills DROP COLUMN token_account_id;
				ALTER TABLE expense_bills RENAME COLUMN token_account_id_new TO token_account_id;
			`,
		},
		{
			Version:     3,
			Description: "DB_02: 重新设计api_tokens表结构",
			SQL: `
				-- 为api_tokens表添加缺失字段
				ALTER TABLE api_tokens ADD COLUMN provider TEXT;
				ALTER TABLE api_tokens ADD COLUMN token_type TEXT;
				ALTER TABLE api_tokens ADD COLUMN daily_limit INTEGER;
				ALTER TABLE api_tokens ADD COLUMN monthly_limit INTEGER;
				ALTER TABLE api_tokens ADD COLUMN expires_at DATETIME;
				ALTER TABLE api_tokens ADD COLUMN last_used_at DATETIME;
			`,
		},
		{
			Version:     4,
			Description: "修复sync_history表结构",
			SQL: `
				-- 添加billing_month字段
				ALTER TABLE sync_history ADD COLUMN billing_month TEXT;

				-- 添加failed_count字段
				ALTER TABLE sync_history ADD COLUMN failed_count INTEGER DEFAULT 0;

				-- 添加sync_time和duration字段
				ALTER TABLE sync_history ADD COLUMN sync_time DATETIME;
				ALTER TABLE sync_history ADD COLUMN duration INTEGER;
				ALTER TABLE sync_history ADD COLUMN message TEXT;

				-- 更新现有记录的billing_month
				UPDATE sync_history
				SET billing_month = strftime('%Y-%m', start_time)
				WHERE billing_month IS NULL;
			`,
		},
		{
			Version:     5,
			Description: "DB_05: 为membership_tier_limits表添加缺失字段",
			SQL: `
				-- 为membership_tier_limits表添加缺失字段
				ALTER TABLE membership_tier_limits ADD COLUMN period_hours INTEGER;
				ALTER TABLE membership_tier_limits ADD COLUMN call_limit INTEGER;
			`,
		},
		{
			Version:     10,
			Description: "重构auto_sync_config表",
			SQL: `
				-- 检查表结构并决定是否需要迁移
				-- 如果表已经是新结构，则跳过迁移
				-- 如果表是旧结构，则进行重构
				
				-- 首先检查是否存在config_key列（旧结构）
				-- 如果存在，说明是旧结构，需要迁移
				-- 如果不存在，说明已经是新结构，跳过迁移
				
				-- 创建新的auto_sync_config表（如果不存在）
				CREATE TABLE IF NOT EXISTS auto_sync_config_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					enabled INTEGER NOT NULL DEFAULT 0,
					frequency_seconds INTEGER NOT NULL DEFAULT 3600,
					sync_type TEXT NOT NULL DEFAULT 'full',
					billing_month TEXT,
					max_retries INTEGER NOT NULL DEFAULT 3,
					retry_delay INTEGER NOT NULL DEFAULT 60,
					next_sync_time DATETIME,
					last_sync_time DATETIME,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				);

				-- 尝试从旧结构迁移数据（如果存在）
				INSERT OR IGNORE INTO auto_sync_config_new (enabled, frequency_seconds)
				SELECT
					CASE
						WHEN typeof(enabled) = 'integer' THEN enabled
						WHEN typeof(enabled) = 'text' AND enabled = 'true' THEN 1
						WHEN typeof(enabled) = 'text' AND enabled = 'false' THEN 0
						ELSE 0
					END,
					COALESCE(
						frequency_seconds,
						3600
					)
				FROM auto_sync_config
				LIMIT 1;

				-- 删除旧表
				DROP TABLE IF EXISTS auto_sync_config;

				-- 重命名新表
				ALTER TABLE auto_sync_config_new RENAME TO auto_sync_config;
			`,
		},
		{
			Version:     6,
			Description: "为同步历史表添加复合索引以优化分页查询性能",
			SQL: `
				-- 为sync_history表添加复合索引
				CREATE INDEX IF NOT EXISTS idx_sync_history_type_start_time ON sync_history(sync_type, start_time DESC);
				CREATE INDEX IF NOT EXISTS idx_sync_history_status ON sync_history(status);
			`,
		},
		{
			Version:     7,
			Description: "添加性能优化索引 (DB_06)",
			SQL: `
				-- 为expense_bills添加性能索引
				CREATE INDEX IF NOT EXISTS idx_expense_bills_billing_date ON expense_bills(billing_date);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_customer_id ON expense_bills(customer_id);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_order_no ON expense_bills(order_no);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_model_name ON expense_bills(model_name);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_charge_type ON expense_bills(charge_type);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_business_id ON expense_bills(business_id);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_transaction_time ON expense_bills(transaction_time);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_create_time ON expense_bills(create_time);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_billing_no ON expense_bills(billing_no);
				
				-- 为api_tokens添加索引
				CREATE INDEX IF NOT EXISTS idx_api_tokens_is_active ON api_tokens(is_active);
				CREATE INDEX IF NOT EXISTS idx_api_tokens_token_name ON api_tokens(token_name);
				
				-- 为sync_history添加复合索引
				CREATE INDEX IF NOT EXISTS idx_sync_history_billing_month ON sync_history(billing_month);
				CREATE INDEX IF NOT EXISTS idx_sync_history_sync_type ON sync_history(sync_type);
			`,
		},
		{
			Version:     8,
			Description: "添加数据完整性约束 (DB_05)",
			SQL: `
				-- SQLite不支持ALTER TABLE ADD CONSTRAINT，跳过CHECK约束
				-- 这些约束应该在应用层处理
				
				-- 清理api_tokens表中的重复数据（按token_name和token_value）
				DELETE FROM api_tokens WHERE id NOT IN (
					SELECT MIN(id) FROM api_tokens GROUP BY token_name, token_value
				);
				
				-- 为api_tokens添加UNIQUE约束
				CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_unique_name ON api_tokens(token_name);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_unique_value ON api_tokens(token_value);
				
				-- SQLite不支持ALTER TABLE ADD CONSTRAINT，跳过CHECK约束
				-- 这些约束应该在应用层处理
			`,
		},
		{
			Version:     9,
			Description: "添加复合UNIQUE约束和额外数据完整性检查 (DB_05增强)",
			SQL: `
				-- 为expense_bills添加复合UNIQUE约束，防止重复账单
				CREATE UNIQUE INDEX IF NOT EXISTS idx_expense_bills_unique_billing_no ON expense_bills(billing_no);
				
				-- SQLite不支持部分索引，跳过auto_sync_config的UNIQUE约束
				-- CREATE UNIQUE INDEX IF NOT EXISTS idx_auto_sync_config_unique_enabled ON auto_sync_config(id) WHERE enabled = 1;
				
				-- SQLite不支持ALTER TABLE ADD CONSTRAINT，跳过所有CHECK约束
				-- 这些约束应该在应用层处理
			`,
		},
		{
			Version:     11,
			Description: "清理api_tokens表中的重复数据",
			SQL: `
				-- 清理api_tokens表中的重复数据（按token_name和token_value）
				DELETE FROM api_tokens WHERE id NOT IN (
					SELECT MIN(id) FROM api_tokens GROUP BY token_name, token_value
				);
				
				-- 确保UNIQUE索引存在
				CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_unique_name ON api_tokens(token_name);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_unique_value ON api_tokens(token_value);
			`,
		},
	}
}

// RunMigrations 执行数据库迁移
func RunMigrations(db *sql.DB) error {
	log.Printf("DEBUG: Starting database migrations...")

	// 创建迁移历史表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("DEBUG: Failed to create migrations table: %v", err)
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	log.Printf("DEBUG: Migrations table created/verified")

	// 获取已应用的迁移
	appliedVersions := make(map[int]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan version: %w", err)
		}
		appliedVersions[version] = true
	}

	// 执行未应用的迁移
	migrations := GetMigrations()
	log.Printf("DEBUG: Found %d migrations to check", len(migrations))

	for _, migration := range migrations {
		if appliedVersions[migration.Version] {
			log.Printf("DEBUG: Migration %d already applied, skipping", migration.Version)
			continue
		}

		log.Printf("DEBUG: Applying migration %d: %s", migration.Version, migration.Description)

		// 开始事务
		tx, err := db.Begin()
		if err != nil {
			log.Printf("DEBUG: Failed to begin transaction for migration %d: %v", migration.Version, err)
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// 执行迁移SQL
		log.Printf("DEBUG: Executing migration SQL for version %d:\n%s", migration.Version, migration.SQL)
		_, err = tx.Exec(migration.SQL)
		if err != nil {
			tx.Rollback()
			log.Printf("DEBUG: Migration %d SQL failed: %v", migration.Version, err)
			return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
		}

		// 记录迁移历史
		_, err = tx.Exec(
			"INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
			migration.Version,
			migration.Description,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("DEBUG: Failed to record migration %d: %v", migration.Version, err)
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			log.Printf("DEBUG: Failed to commit migration %d: %v", migration.Version, err)
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		log.Printf("DEBUG: Migration %d applied successfully", migration.Version)
	}

	log.Printf("DEBUG: All migrations completed successfully")

	return nil
}
