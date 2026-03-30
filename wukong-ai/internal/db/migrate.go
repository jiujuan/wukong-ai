package db

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jiujuan/wukong-ai/pkg/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations 执行数据库迁移
func RunMigrations() error {
	db := Get()

	// 确保 migrations 表存在
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 读取所有迁移文件
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// 按文件名排序确保执行顺序
	sort.Strings(migrationFiles)

	// 执行每个迁移
	for _, filename := range migrationFiles {
		// 检查是否已应用
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", filename).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", filename, err)
		}

		if count > 0 {
			logger.Info("migration already applied", "file", filename)
			continue
		}

		// 读取迁移内容
		content, err := migrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// 执行迁移
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// 记录迁移
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", filename)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		logger.Info("migration applied successfully", "file", filename)
	}

	logger.Info("all migrations completed")
	return nil
}
