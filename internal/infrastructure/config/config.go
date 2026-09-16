/**
@Time : 2026/09/16 09:38
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

// 全局配置结构体
type Config struct {
	Server ServerConfig `mapstructure:"server" json:"server"`
	Sqlite SqliteConfig `mapstructure:"sqlite" json:"sqlite"`
	Log    LogConfig    `mapstructure:"log" json:"log"`
}

// 系统配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// 日志配置
type LogConfig struct {
	Level  string        `mapstructure:"level"`
	Format string        `mapstructure:"format"`
	Output string        `mapstructure:"output"`
	File   LogFileConfig `mapstructure:"file"`
}

// 日志配置
type LogFileConfig struct {
	Path       string `mapstructure:"path"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}
type SqliteConfig struct {
	Path    string `mapstructure:"path" json:"path"`
	LogMode bool   `mapstructure:"log-mode" json:"logMode"`
}

// 加载配置文件
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// 可选本地覆盖：configs/config.local.yaml（不入库，放密码等）
	localPath := filepath.Join(filepath.Dir(path), "config.local.yaml")
	if st, err := os.Stat(localPath); err == nil && !st.IsDir() {
		v.SetConfigFile(localPath)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge local config %s: %w", localPath, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
