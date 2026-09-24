/**
@Time : 2026/09/17 11:04
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package alarm

import "context"

// Alarm 告警记录
type Alarm struct {
	ID             string `json:"id"`             // 消息唯一ID UUID
	CompanyCode    string `json:"companyCode"`    // 公司编码
	CompanyName    string `json:"companyName"`    // 公司名称
	DeviceID       string `json:"deviceId"`       // 边缘设备ID
	VideoCode      string `json:"videoCode"`      // 视频编码，如 h264
	VideoAlarmType string `json:"videoAlarmType"` // 告警类型
	AlarmStatus    string `json:"alarmStatus"`    // 0=销警/正常 1=报警
	AlarmTime      string `json:"alarmTime"`      // 报警/销警时间 yyyy-MM-dd HH:mm:ss
	AlarmPicture   string `json:"alarmPicture"`   // 告警图片
	ReportType     string `json:"reportType"`     // normal / alarm / cancel
	Count          int    `json:"count"`          // 当前人数
	Limit          int    `json:"limit"`          // 超员阈值
	Timestamp      int64  `json:"timestamp"`      // 上报时间戳
	CreateTime     string `json:"createTime"`     // 入库时间
}

type Repository interface {
	// Save 保存一条告警记录
	Save(ctx context.Context, a *Alarm) error
	// List 按条件分页查询告警记录
	List(ctx context.Context, q Query) (*Page, error)
}

// Query 告警查询条件。
type Query struct {
	CompanyCode string // 公司编码
	DeviceID    string // 边缘设备ID
	AlarmStatus string // 报警状态 0/1
	StartTime   string // 起始时间（alarm_time >=）
	EndTime     string // 结束时间（alarm_time <=）
	Page        int    // 页码，从 1 开始
	PageSize    int    // 每页数量
}

// Page 分页结果。
type Page struct {
	List     []Alarm `json:"list"`
	Total    int64   `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
}
