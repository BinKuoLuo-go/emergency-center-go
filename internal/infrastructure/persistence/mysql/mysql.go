/**
@Time : 2026/09/04 13:41
@Author: FangYao( 方少、)
@Description: mysql
@Email: fy20030315@163.com
*/

package mysql

import (
	"database/sql"
	"fmt"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/config"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/model"
	applog "github.com/BinKuoLuo-go/emergency-center-go/pkg/log"
	_ "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

// NewMysqlDB 连接 MySQL：库不存在则创建。
func NewMysqlDB(cfg config.MySQLConfig) (*gorm.DB, error) {
	if cfg.Charset == "" {
		cfg.Charset = "utf8mb4"
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("mysql.database 不能为空")
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 100
	}

	if err := ensureDatabase(cfg); err != nil {
		return nil, err
	}

	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(gormmysql.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := autoMigrateAll(db); err != nil {
		return nil, fmt.Errorf("自动迁移: %w", err)
	}
	applog.Info("MySQL 准备就绪", "database", cfg.Database)
	return db, nil
}

func ensureDatabase(cfg config.MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Charset)
	raw, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql bootstrap: %w", err)
	}
	defer raw.Close()

	raw.SetConnMaxLifetime(time.Minute)
	if err := raw.Ping(); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	stmt := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.Database,
	)
	if _, err := raw.Exec(stmt); err != nil {
		return fmt.Errorf("create database %s: %w", cfg.Database, err)
	}
	return nil
}

// autoMigrateAll 自动表迁移
func autoMigrateAll(db *gorm.DB) error {
	models := []any{
		&model.ObjectStoreConfig{},
		&model.Alarm{},
	}
	for _, m := range models {
		exists := db.Migrator().HasTable(m) // 迁移前记录表是否已存在
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("%T: %w", m, err) // 错误中带上模型类型，方便定位
		}
		if !exists {
			applog.Info("created table", "model", fmt.Sprintf("%T", m))
		}
	}
	return nil
}
