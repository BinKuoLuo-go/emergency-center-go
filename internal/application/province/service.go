/**
@Time : 2026/09/24 11:42
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package province

import (
	"context"
	"encoding/base64"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/alarm"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/province"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
	applog "github.com/BinKuoLuo-go/emergency-center-go/pkg/log"
	"github.com/google/uuid"
	"io"
)

// maxPictureBytes 省平台要求报警图片解码后 ≤1MB。
const maxPictureBytes = 1 << 20

// Service 省平台上报服务。
type Service struct {
	client        province.Client
	storeProvider func() port.ObjectStore
	enabled       bool
}

// NewService 构建上报服务。client 为 nil 或未启用时，Report 直接跳过。
func NewService(enabled bool, client province.Client, storeProvider func() port.ObjectStore) *Service {
	return &Service{enabled: enabled, client: client, storeProvider: storeProvider}
}

// Report 将一条边缘端告警转换为省平台格式并上报
func (s *Service) Report(ctx context.Context, a *alarm.Alarm) error {
	if !s.enabled || s.client == nil || a == nil {
		return nil
	}

	item := province.FromAlarm(a)

	// 下载告警图并 base64,失败不阻断上报，仅告警图留空
	item.AlarmPic = s.loadPictureBase64(ctx, a.AlarmPicture)

	return s.client.Report(ctx, uuid.NewString(), []province.GatherAlarm{item})
}

// loadPictureBase64 从对象存储下载图片并返回 base64 字符串；失败或为空返回空串。
func (s *Service) loadPictureBase64(ctx context.Context, key string) string {
	if key == "" || s.storeProvider == nil {
		return ""
	}
	store := s.storeProvider()
	if store == nil {
		return ""
	}
	rc, err := store.Get(ctx, key)
	if err != nil {
		applog.Warn("省平台上报拉取告警图失败", "key", key, "err", err)
		return ""
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		applog.Warn("省平台上报读取告警图失败", "key", key, "err", err)
		return ""
	}
	if len(data) == 0 {
		return ""
	}
	if len(data) > maxPictureBytes {
		applog.Warn("告警图超过 1MB，省平台可能拒收", "key", key, "bytes", len(data))
	}
	return base64.StdEncoding.EncodeToString(data)
}
