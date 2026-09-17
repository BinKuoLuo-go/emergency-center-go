/**
@Time : 2026/09/17 15:39
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package alarm

import (
	"context"
	"errors"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/alarm"
	"github.com/google/uuid"
	"strings"
	"time"
)

// Service 告警应用服务。
type Service struct {
	repo alarm.Repository
}

func NewService(repo alarm.Repository) *Service {
	return &Service{repo: repo}
}

// Record 校验并保存一条告警、销警记录。
func (s *Service) Record(ctx context.Context, a *alarm.Alarm) error {
	if a == nil {
		return errors.New("告警消息为空")
	}
	if strings.TrimSpace(a.ID) == "" {
		a.ID = uuid.NewString()
	}
	if strings.TrimSpace(a.CompanyCode) == "" {
		return errors.New("companyCode 不能为空")
	}
	// 状态
	if a.AlarmStatus != "1" {
		a.AlarmStatus = "0"
	}
	if strings.TrimSpace(a.AlarmTime) == "" {
		a.AlarmTime = time.Now().Format("2006-01-02 15:04:05")
	}
	return s.repo.Save(ctx, a)
}

// List 分页查询告警记录。
func (s *Service) List(ctx context.Context, q alarm.Query) (*alarm.Page, error) {
	return s.repo.List(ctx, q)
}
