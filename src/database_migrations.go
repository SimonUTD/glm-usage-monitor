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
			Description: "添加缺失的expense_bills字段",
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
			Description: "修复sync_history表结构",
			SQL: `
				-- 添加billing_month字段
				ALTER TABLE sync_history ADD COLUMN billing_month TEXT;

				-- 添加failed_count字段
				ALTER TABLE sync_history ADD COLUMN failed_count INTEGER DEFAULT 0;

				-- 更新现有记录的billing_month
				UPDATE sync_history
				SET billing_month = strftime('%Y-%m', start_time)
				WHERE billing_month IS NULL;
			`,
		},
		{
			Version:     3,
			Description: "添加关键索引",
			SQL: `
				-- 为expense_bills添加索引
				CREATE INDEX IF NOT EXISTS idx_expense_bills_customer_id ON expense_bills(customer_id);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_billing_date ON expense_bills(billing_date);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_order_no ON expense_bills(order_no);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_time_window_start ON expense_bills(time_window_start);
				CREATE INDEX IF NOT EXISTS idx_expense_bills_transaction_time ON expense_bills(transaction_time);
			`,
		},
		{
			Version:     4,
			Description: "重构auto_sync_config表",
			SQL: `
				-- 创建新的auto_sync_config表
				CREATE TABLE IF NOT EXISTS auto_sync_config_v2 (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					enabled INTEGER NOT NULL DEFAULT 0,
					frequency_seconds INTEGER NOT NULL DEFAULT 10,
					next_sync_time DATETIME,
					last_sync_time DATETIME,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				);

				-- 迁移现有配置（如果有）
				INSERT INTO auto_sync_config_v2 (enabled, frequency_seconds)
				SELECT
					CASE WHEN config_value = 'true' THEN 1 ELSE 0 END,
					10
				FROM auto_sync_config
				WHERE config_key = 'enabled'
				LIMIT 1;

				-- 删除旧表
				DROP TABLE IF EXISTS auto_sync_config;

				-- 重命名新表
				ALTER TABLE auto_sync_config_v2 RENAME TO auto_sync_config;
			`,
		},
		{
			Version: 4,
			Description: "为同步历史表添加复合索引以优化分页查询性能",
			SQL: `
				-- 为sync_history表添加复合索引
				CREATE INDEX IF NOT EXISTS idx_sync_history_type_start_time ON sync_history(sync_type, start_time DESC);
				CREATE INDEX IF NOT EXISTS idx_sync_history_status ON sync_history(status);
			`,
		},
	}
}

// RunMigrations 执行数据库迁移
func RunMigrations(db *sql.DB) error {
	// 创建迁移历史表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

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
	for _, migration := range migrations {
		if appliedVersions[migration.Version] {
			log.Printf("Migration %d already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Description)

		// 开始事务
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// 执行迁移SQL
		_, err = tx.Exec(migration.SQL)
		if err != nil {
			tx.Rollback()
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
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		log.Printf("Migration %d applied successfully", migration.Version)
	}

	return nil
}