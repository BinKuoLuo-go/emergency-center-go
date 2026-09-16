/**
@Time : 2026/09/16 15:06
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package objectstore

import (
	"context"
	"fmt"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"strings"
	"time"
)

// MinIO S3 兼容对象存储适配器。
type MinIO struct {
	cfg    port.ObjectStoreConfig
	client *minio.Client
	bucket string
}

func NewMinIO(cfg port.ObjectStoreConfig) (*MinIO, error) {
	client, bucket, err := newMinIOClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("minio: %w", err)
	}
	return &MinIO{cfg: cfg, client: client, bucket: bucket}, nil
}

func (m *MinIO) Provider() string { return "minio" }

func (m *MinIO) Health(ctx context.Context) error {
	ok, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("minio: bucket %s not found", m.bucket)
	}
	return nil
}

func (m *MinIO) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{}
	if contentType != "" {
		opts.ContentType = contentType
	}
	_, err := m.client.PutObject(ctx, m.bucket, strings.TrimLeft(key, "/"), r, size, opts)
	return err
}

func (m *MinIO) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, strings.TrimLeft(key, "/"), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (m *MinIO) Delete(ctx context.Context, key string) error {
	return m.client.RemoveObject(ctx, m.bucket, strings.TrimLeft(key, "/"), minio.RemoveObjectOptions{})
}

func (m *MinIO) PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if m.cfg.PublicBase != "" {
		return strings.TrimRight(m.cfg.PublicBase, "/") + "/" + strings.TrimLeft(key, "/"), nil
	}
	if expiry <= 0 {
		expiry = time.Hour
	}
	u, err := m.client.PresignedGetObject(ctx, m.bucket, strings.TrimLeft(key, "/"), expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (m *MinIO) List(ctx context.Context, prefix string) ([]port.ObjectInfo, error) {
	objCh := m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Prefix:    strings.TrimLeft(prefix, "/"),
		Recursive: true,
	})
	out := make([]port.ObjectInfo, 0)
	for obj := range objCh {
		if obj.Err != nil {
			return nil, obj.Err
		}
		out = append(out, port.ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ETag:         strings.Trim(obj.ETag, `"`),
			ContentType:  obj.ContentType,
		})
	}
	return out, nil
}

func newMinIOClient(cfg port.ObjectStoreConfig) (*minio.Client, string, error) {
	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		return nil, "", fmt.Errorf("bucket required")
	}
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, "", fmt.Errorf("accessKey and secretKey required")
	}

	endpoint, secure, err := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)
	if err != nil {
		return nil, "", err
	}
	if endpoint == "" {
		return nil, "", fmt.Errorf("endpoint required")
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
	}
	if cfg.PathStyle {
		opts.BucketLookup = minio.BucketLookupPath
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, "", err
	}
	return client, bucket, nil
}

func normalizeEndpoint(raw string, useSSL bool) (host string, secure bool, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", useSSL, nil
	}
	secure = useSSL
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, perr := url.Parse(raw)
		if perr != nil {
			return "", false, perr
		}
		secure = u.Scheme == "https"
		host = u.Host
		if host == "" {
			return "", false, fmt.Errorf("invalid endpoint")
		}
		return host, secure, nil
	}
	return strings.TrimRight(raw, "/"), secure, nil
}

var _ port.ObjectStore = (*MinIO)(nil)
