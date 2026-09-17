/**
@Time : 2026/09/17 15:45
@Author: FangYao( 方少、)
@Description: 告警记录仓储实现。
@Email: fy20030315@163.com
*/

package persistence

import (
	"context"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/alarm"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
	"strings"
	"time"
)

// AlarmRepository
type AlarmRepository struct {
	db *gorm.DB
}

func NewAlarmRepository(db *gorm.DB) *AlarmRepository {
	return &AlarmRepository{db: db}
}

// Save 保存一条告警记录,同ID存在则更新，否则插入
func (r *AlarmRepository) Save(ctx context.Context, a *alarm.Alarm) error {
	m := toAlarmModel(a)
	if m.CreateTime == "" {
		m.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	}
	return r.db.WithContext(ctx).Save(m).Error
}

// List 按条件分页查询告警记录。
func (r *AlarmRepository) List(ctx context.Context, q alarm.Query) (*alarm.Page, error) {
	db := r.db.WithContext(ctx).Model(&model.Alarm{})

	if strings.TrimSpace(q.CompanyCode) != "" {
		db = db.Where("company_code = ?", strings.TrimSpace(q.CompanyCode))
	}
	if strings.TrimSpace(q.DeviceID) != "" {
		db = db.Where("device_id = ?", strings.TrimSpace(q.DeviceID))
	}
	if strings.TrimSpace(q.AlarmStatus) != "" {
		db = db.Where("alarm_status = ?", strings.TrimSpace(q.AlarmStatus))
	}
	if strings.TrimSpace(q.StartTime) != "" {
		db = db.Where("alarm_time >= ?", strings.TrimSpace(q.StartTime))
	}
	if strings.TrimSpace(q.EndTime) != "" {
		db = db.Where("alarm_time <= ?", strings.TrimSpace(q.EndTime))
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	page, pageSize := normalizePage(q.Page, q.PageSize)
	var models []model.Alarm
	if err := db.Order("alarm_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&models).Error; err != nil {
		return nil, err
	}

	list := make([]alarm.Alarm, 0, len(models))
	for i := range models {
		list = append(list, toAlarmDomain(&models[i]))
	}

	return &alarm.Page{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// toAlarmModel 领域实体到落库模型
func toAlarmModel(a *alarm.Alarm) *model.Alarm {
	if a == nil {
		return &model.Alarm{}
	}
	return &model.Alarm{
		ID:             a.ID,
		CompanyCode:    a.CompanyCode,
		CompanyName:    a.CompanyName,
		DeviceID:       a.DeviceID,
		VideoCode:      a.VideoCode,
		VideoAlarmType: a.VideoAlarmType,
		AlarmStatus:    a.AlarmStatus,
		AlarmTime:      a.AlarmTime,
		AlarmPicture:   a.AlarmPicture,
		ReportType:     a.ReportType,
		Count:          a.Count,
		Limit:          a.Limit,
		Timestamp:      a.Timestamp,
		CreateTime:     a.CreateTime,
	}
}

// toAlarmDomain 落库模型到领域实体。
func toAlarmDomain(m *model.Alarm) alarm.Alarm {
	return alarm.Alarm{
		ID:             m.ID,
		CompanyCode:    m.CompanyCode,
		CompanyName:    m.CompanyName,
		DeviceID:       m.DeviceID,
		VideoCode:      m.VideoCode,
		VideoAlarmType: m.VideoAlarmType,
		AlarmStatus:    m.AlarmStatus,
		AlarmTime:      m.AlarmTime,
		AlarmPicture:   m.AlarmPicture,
		ReportType:     m.ReportType,
		Count:          m.Count,
		Limit:          m.Limit,
		Timestamp:      m.Timestamp,
		CreateTime:     m.CreateTime,
	}
}
