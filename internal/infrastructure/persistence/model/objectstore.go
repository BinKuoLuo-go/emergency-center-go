/**
@Time : 2026/09/16 15:07
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package model

// ObjectStoreConfig 对象存储对接配置，平台不自建存储。
type ObjectStoreConfig struct {
	ID         int    `gorm:"column:id;primaryKey" json:"id"`                                // 主键
	Enabled    bool   `gorm:"column:enabled;not null;default:0" json:"enabled"`              // 是否启用对象存储（false 时走 noop，仅本地不落对象）
	Provider   string `gorm:"column:provider;size:32;not null;default:noop" json:"provider"` // 驱动名：noop | minio
	Endpoint   string `gorm:"column:endpoint;size:255" json:"endpoint"`                      // MinIO 地址
	Bucket     string `gorm:"column:bucket;size:128" json:"bucket"`                          // 桶名
	AccessKey  string `gorm:"column:access_key;size:128" json:"accessKey"`                   // 访问密钥 AccessKey
	SecretKey  string `gorm:"column:secret_key;size:256" json:"secretKey"`                   // 访问密钥 SecretKey
	UseSSL     bool   `gorm:"column:use_ssl;not null;default:0" json:"useSSL"`               // 是否使用 HTTPS 连接 MinIO
	PathStyle  bool   `gorm:"column:path_style;not null;default:1" json:"pathStyle"`         // 是否使用 path-style 访问
	PublicBase string `gorm:"column:public_base;size:255" json:"publicBase"`                 // 对外访问 URL 前缀；配置后图片直链用它拼接，否则走预签名/代理
	CreateTime string `gorm:"column:create_time;size:50" json:"createTime"`                  // 创建时间（yyyy-MM-dd HH:mm:ss）
	UpdateTime string `gorm:"column:update_time;size:50" json:"updateTime"`                  // 更新时间（yyyy-MM-dd HH:mm:ss）
}

func (ObjectStoreConfig) TableName() string { return "object_store_config" }
