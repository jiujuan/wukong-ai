package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

var (
	db   *sql.DB
	once sync.Once
)

// Init 初始化数据库连接
func Init(dsn string) *sql.DB {
	once.Do(func() {
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			panic(fmt.Sprintf("failed to open database: %v", err))
		}

		// 设置连接池参数
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)

		// 测试连接
		if err = db.Ping(); err != nil {
			panic(fmt.Sprintf("failed to ping database: %v", err))
		}

		logger.Info("database connected successfully")
	})
	return db
}

// Get 获取数据库实例
func Get() *sql.DB {
	if db == nil {
		panic("database not initialized, call Init() first")
	}
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Transaction 执行事务
func Transaction(fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
