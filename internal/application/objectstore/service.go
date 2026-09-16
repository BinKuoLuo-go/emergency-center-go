/**
@Time : 2026/09/16 15:03
@Author: FangYao( 方少、)
@Description: 对象存储
@Email: fy20030315@163.com
*/

package objectstoreapp

import (
	"context"
	"errors"
	objadapter "github.com/BinKuoLuo-go/emergency-center-go/internal/adapter/objectstore"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/model"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
	"strings"
	"sync"

	"gorm.io/gorm"
)

type Service struct {
	repo *persistence.ObjectStoreConfigRepository

	mu    sync.RWMutex
	store port.ObjectStore
}

func NewService(repo *persistence.ObjectStoreConfigRepository) *Service {
	s := &Service{repo: repo, store: objadapter.NewNoop()}
	_ = s.reload()
	return s
}

func (s *Service) Store() port.ObjectStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.store
}

func (s *Service) GetOrEmpty() (*model.ObjectStoreConfig, error) {
	row, err := s.repo.Get()
	if err == nil {
		return row, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return persistence.DefaultObjectStoreConfig(), nil
	}
	return nil, err
}

func (s *Service) Save(row *model.ObjectStoreConfig) error {
	if row == nil {
		return errors.New("参数错误")
	}
	row.Provider = strings.ToLower(strings.TrimSpace(row.Provider))
	if row.Provider == "" {
		row.Provider = "noop"
	}
	switch row.Provider {
	case "noop", "minio":
	default:
		return errors.New("provider 仅支持 noop / minio")
	}
	if row.Enabled && row.Provider != "noop" {
		if strings.TrimSpace(row.Bucket) == "" {
			return errors.New("启用时必须填写 Bucket")
		}
		if strings.TrimSpace(row.AccessKey) == "" || strings.TrimSpace(row.SecretKey) == "" {
			return errors.New("启用时必须填写 AccessKey / SecretKey")
		}
		if strings.TrimSpace(row.Endpoint) == "" {
			return errors.New("MinIO 必须填写 Endpoint")
		}
	}
	if err := s.repo.Save(row); err != nil {
		return err
	}
	return s.reload()
}

func (s *Service) Health(ctx context.Context) error {
	return s.Store().Health(ctx)
}

func (s *Service) reload() error {
	row, err := s.repo.Get()
	cfg := port.ObjectStoreConfig{Provider: "noop"}
	if err == nil {
		cfg = s.repo.ToPortConfig(row)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	store, ferr := objadapter.Factory(cfg)
	if ferr != nil {
		store = objadapter.NewNoop()
	}
	s.mu.Lock()
	s.store = store
	s.mu.Unlock()
	return ferr
}
