/**
@Time : 2026/09/16 15:19
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package objectstore

import (
	"context"
	"fmt"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
	"io"
	"strings"
	"time"
)

// Noop 未配置对象存储时的默认实现：所有写读明确失败，避免静默丢数据。
type Noop struct{}

func NewNoop() *Noop { return &Noop{} }

func (n *Noop) Provider() string { return "noop" }

func (n *Noop) Health(ctx context.Context) error {
	return port.ErrStoreDisabled
}

func (n *Noop) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	return port.ErrStoreDisabled
}

func (n *Noop) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, port.ErrStoreDisabled
}

func (n *Noop) Delete(ctx context.Context, key string) error {
	return port.ErrStoreDisabled
}

func (n *Noop) PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", port.ErrStoreDisabled
}

func (n *Noop) List(ctx context.Context, prefix string) ([]port.ObjectInfo, error) {
	return nil, port.ErrStoreDisabled
}

// Factory 按配置构建 ObjectStore。未启用或未知驱动 → Noop。
func Factory(cfg port.ObjectStoreConfig) (port.ObjectStore, error) {
	if !cfg.Enabled {
		return NewNoop(), nil
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", "noop":
		return NewNoop(), nil
	case "minio":
		return NewMinIO(cfg)
	default:
		return nil, fmt.Errorf("%w: %s", port.ErrStoreUnsupported, cfg.Provider)
	}
}
