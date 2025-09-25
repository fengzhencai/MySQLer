package database

import (
	"fmt"
	"time"

	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize 初始化数据库连接
func Initialize(cfg *config.Config) (*gorm.DB, error) {
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 如果是生产环境，调整日志级别
	if cfg.AppEnv == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移数据库表
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	// 修复旧版本列名（target_table_name -> table_name）
	if err := fixExecutionRecordsSchema(db, cfg.DBName); err != nil {
		return nil, fmt.Errorf("failed to fix schema: %w", err)
	}

	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Connection{},
		&models.ExecutionRecord{},
		&models.AuditLog{},
	)
}

// fixExecutionRecordsSchema 兼容旧列名，确保仅存在 table_name 列
func fixExecutionRecordsSchema(db *gorm.DB, dbName string) error {
	type cnt struct{ C int }
	var a, b cnt
	// 检查是否存在旧列 target_table_name
	if err := db.Raw(
		"SELECT COUNT(*) AS c FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME='execution_records' AND COLUMN_NAME='target_table_name'",
		dbName,
	).Scan(&a).Error; err != nil {
		return err
	}
	if a.C == 0 {
		return nil
	}
	// 检查是否已存在新列 table_name
	if err := db.Raw(
		"SELECT COUNT(*) AS c FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME='execution_records' AND COLUMN_NAME='table_name'",
		dbName,
	).Scan(&b).Error; err != nil {
		return err
	}

	if b.C == 0 {
		// 重命名旧列为新列
		if err := db.Exec("ALTER TABLE execution_records CHANGE COLUMN target_table_name table_name varchar(200) NOT NULL").Error; err != nil {
			return err
		}
	} else {
		// 已有新列，直接删除旧列，避免 NOT NULL 约束导致插入失败
		if err := db.Exec("ALTER TABLE execution_records DROP COLUMN target_table_name").Error; err != nil {
			return err
		}
	}
	return nil
}
