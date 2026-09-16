/**
@Time : 2026/09/16 14:33
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package snapshot

import (
	"context"
	"errors"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/snapshot"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
	"io"
	"net/url"
)

// imageProxyPath 图片流式代理接口路径前缀。
// 列表返回的是相对路径（如 /server/file/edge-01/composite/...），前端同源直接可用；
// 跨域使用时自行拼上中心平台地址即可。
const imageProxyPath = "/file/"

// Service 快照应用服务。
type Service struct {
	repo          snapshot.Repository
	storeProvider func() port.ObjectStore
}

func NewService(repo snapshot.Repository, storeProvider func() port.ObjectStore) *Service {
	return &Service{repo: repo, storeProvider: storeProvider}
}

// ListByDeviceAndDate 按设备 + 日期文件夹分页查询快照，并附带图片访问相对路径。
func (s *Service) ListByDeviceAndDate(ctx context.Context, q snapshot.Query) (*snapshot.Page, error) {
	if q.DeviceID == "" {
		return nil, errors.New("deviceId 不能为空")
	}
	if q.Date == "" {
		return nil, errors.New("date 不能为空")
	}

	page, err := s.repo.ListByDeviceAndDate(ctx, q)
	if err != nil {
		return nil, err
	}

	// 生成短代理 URL（相对路径），由中心平台流式转发图片，避免长预签名 URL 且不过期
	for i := range page.List {
		u := &url.URL{Path: imageProxyPath + page.List[i].StoragePath}
		page.List[i].URL = u.String()
	}
	return page, nil
}

// OpenImage 打开图片对象流，供图片代理接口读取。
func (s *Service) OpenImage(ctx context.Context, key string) (io.ReadCloser, error) {
	store := s.storeProvider()
	return store.Get(ctx, key)
}
