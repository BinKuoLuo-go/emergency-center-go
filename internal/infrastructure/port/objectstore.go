/**
@Time : 2026/09/16 15:09
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package port

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrStoreDisabled    = errors.New("object store disabled")
	ErrStoreUnsupported = errors.New("object store provider unsupported")
)

// ObjectStore 对象存储端口：对接 MinIO，平台自身不实现存储引擎。
// 用于抓拍图、录像文件归档、导出包等「放对象」场景。
type ObjectStore interface {
	// Provider 返回驱动名：noop | minio
	Provider() string
	// Health 探测连通性
	Health(ctx context.Context) error
	// Put 上传对象
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	// Get 下载对象
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete 删除对象
	Delete(ctx context.Context, key string) error
	// PresignGet 生成限时下载 URL
	PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error)
	// List 列出指定前缀下的全部对象，按 Key 字典序返回
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
}

// ObjectStoreConfig 运行时配置视图
type ObjectStoreConfig struct {
	Enabled    bool
	Provider   string // noop | minio
	Endpoint   string
	Bucket     string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	PathStyle  bool   // MinIO 常用 path-style
	PublicBase string // 可选：对外访问前缀
}

// ObjectInfo 对象元信息。
type ObjectInfo struct {
	Key          string    // 对象 key（含前缀路径）
	Size         int64     // 文件大小（字节）
	LastModified time.Time // 最后修改时间
	ETag         string    // 内容哈希（去重用）
	ContentType  string    // MIME 类型
}
