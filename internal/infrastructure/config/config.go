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
	Mysql  MySQLConfig  `mapstructure:"mysql" json:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis" json:"redis"`
	Mqtt   MQTTConfig   `mapstructure:"mqtt" json:"mqtt"`
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

// MySQL配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func (c MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

// Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
	PoolSize int    `mapstructure:"pool_size"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// MQTT配置 订阅边缘端上报的告警、心跳、设备状态消息
type MQTTConfig struct {
	Enabled     bool   `mapstructure:"enabled" json:"enabled"`          // 是否启用 MQTT 订阅
	Broker      string `mapstructure:"broker" json:"broker"`            // Broker 地址，如 tcp://127.0.0.1:1883
	ClientID    string `mapstructure:"client_id" json:"clientId"`       // 客户端 ID
	Username    string `mapstructure:"username" json:"username"`        // 用户名
	Password    string `mapstructure:"password" json:"password"`        // 密码
	Topic       string `mapstructure:"topic" json:"topic"`              // 告警订阅主题
	StatusTopic string `mapstructure:"status_topic" json:"statusTopic"` // 设备状态订阅主题
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
