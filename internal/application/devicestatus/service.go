/**
@Time : 2026/09/17 15:49
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package devicestatus

import (
	"context"
	"encoding/json"
	"errors"
	devicestatusdomain "github.com/BinKuoLuo-go/emergency-center-go/internal/domain/devicestatus"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/redis"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"sort"
	"strings"
	"time"
)

const (
	onlineKeyPrefix = "emergency:online:"        // 在线状态 key 前缀
	statusKeyPrefix = "emergency:device:status:" // 设备资源状态 key 前缀
	onlineTTL       = 30 * time.Second           // 在线状态 TTL
	statusTTL       = 90 * time.Second           // 设备状态 TTL
)

// Service 设备状态应用服务
type Service struct {
	rdb *redis.Client
}

func NewService(rdb *redis.Client) *Service {
	return &Service{rdb: rdb}
}

// MarkOnline 记录设备在线由心跳触发覆盖写最新在线信息并刷新 TTL。
func (s *Service) MarkOnline(ctx context.Context, deviceID, companyCode, companyName string) error {
	if s.rdb == nil {
		return nil // Redis 不可用，静默降级
	}
	info := devicestatusdomain.OnlineInfo{
		DeviceID:          deviceID,
		CompanyCode:       companyCode,
		CompanyName:       companyName,
		LastHeartbeatTime: time.Now().Format("2006-01-02 15:04:05"),
		Timestamp:         time.Now().UnixMilli(),
	}
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return s.rdb.SetEX(ctx, onlineKeyPrefix+deviceID, string(b), onlineTTL)
}

// Online 查询设备在线状态
func (s *Service) Online(ctx context.Context, deviceID string) (*devicestatusdomain.OnlineInfo, error) {
	if s.rdb == nil {
		return nil, errors.New("Redis 不可用")
	}
	raw, err := s.rdb.Get(ctx, onlineKeyPrefix+deviceID)
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil // 离线
		}
		return nil, err
	}
	var info devicestatusdomain.OnlineInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// MarkStatus 记录最新设备资源状态，覆盖写并刷新 TTL。
func (s *Service) MarkStatus(ctx context.Context, d *devicestatusdomain.DeviceStatus) error {
	if s.rdb == nil {
		return nil // Redis 不可用，静默降级
	}
	if d == nil || d.DeviceID == "" {
		return errors.New("deviceId 不能为空")
	}
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	if d.Status == "" {
		d.Status = "online"
	}
	if d.ReportTime == "" {
		d.ReportTime = time.Now().Format("2006-01-02 15:04:05")
	}
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return s.rdb.SetEX(ctx, statusKeyPrefix+d.DeviceID, string(b), statusTTL)
}

// LatestStatus 查询设备最新资源状态
func (s *Service) LatestStatus(ctx context.Context, deviceID string) (*devicestatusdomain.DeviceStatus, error) {
	if s.rdb == nil {
		return nil, errors.New("Redis 不可用")
	}
	raw, err := s.rdb.Get(ctx, statusKeyPrefix+deviceID)
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil // 无数据
		}
		return nil, err
	}
	var d devicestatusdomain.DeviceStatus
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDevices 返回所有已知设备的 deviceId
func (s *Service) ListDevices(ctx context.Context) ([]string, error) {
	if s.rdb == nil {
		return nil, errors.New("Redis 不可用")
	}
	set := make(map[string]struct{})
	for _, prefix := range []string{onlineKeyPrefix, statusKeyPrefix} {
		keys, err := s.rdb.Keys(ctx, prefix+"*")
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			id := strings.TrimPrefix(k, prefix)
			if id != "" && id != k {
				set[id] = struct{}{}
			}
		}
	}
	list := make([]string, 0, len(set))
	for id := range set {
		list = append(list, id)
	}
	sort.Strings(list)
	return list, nil
}
