/**
@Time : 2026/09/16 09:37
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package sqlite

import (
	"fmt"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/config"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/model"
	applog "github.com/BinKuoLuo-go/emergency-center-go/pkg/log"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"path/filepath"
)

func NewSqlite(cfg config.SqliteConfig) (*gorm.DB, error) {
	// 纯 Go 驱动不会自动创建父目录，先确保目录存在
	if dir := filepath.Dir(cfg.Path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s", cfg.Path)
	showDsn := fmt.Sprintf("file:%s[隐藏参数]", cfg.Path)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("初始化SQLite数据库异常: %w", err)
	}

	if cfg.LogMode {
		db = db.Debug()
	}

	if err := autoMigrateAll(db); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	applog.Infof("初始化SQLite数据库完成! 连接信息: %s", showDsn)
	return db, nil
}

// autoMigrateAll 自动创建/更新业务表结构。
// 后续新增表只需在 models 里追加对应的 model。
func autoMigrateAll(db *gorm.DB) error {
	models := []any{
		&model.ObjectStoreConfig{},
	}
	for _, m := range models {
		existed := db.Migrator().HasTable(m)
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("%T: %w", m, err)
		}
		if !existed {
			applog.Infof("创建数据表: %T", m)
		}
	}
	return nil
}
