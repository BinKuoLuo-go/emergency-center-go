/**
@Time : 2026/09/17 15:37
@Author: FangYao( 方少、)
@Description: 告警记录表
@Email: fy20030315@163.com
*/

package model

// Alarm 告警记录表
type Alarm struct {
	ID             string `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`                                // 消息唯一 ID（UUID）
	CompanyCode    string `gorm:"column:company_code;type:varchar(64);index:idx_company_code" json:"companyCode"` // 公司编码
	CompanyName    string `gorm:"column:company_name;type:varchar(128)" json:"companyName"`                       // 公司名称
	DeviceID       string `gorm:"column:device_id;type:varchar(64);index:idx_device_id" json:"deviceId"`          // 边缘设备 ID
	VideoCode      string `gorm:"column:video_code;type:varchar(64)" json:"videoCode"`                            // 视频编码，如 h264
	VideoAlarmType string `gorm:"column:video_alarm_type;type:varchar(64)" json:"videoAlarmType"`                 // 告警类型，如 overcrowding
	AlarmStatus    string `gorm:"column:alarm_status;type:varchar(8);index:idx_alarm_status" json:"alarmStatus"`  // 0=销警/正常 1=报警
	AlarmTime      string `gorm:"column:alarm_time;type:varchar(32)" json:"alarmTime"`                            // 报警/销警时间 yyyy-MM-dd HH:mm:ss
	AlarmPicture   string `gorm:"column:alarm_picture;type:varchar(512)" json:"alarmPicture"`                     // 告警图片（MinIO 对象键）
	ReportType     string `gorm:"column:report_type;type:varchar(16)" json:"reportType"`                          // normal / alarm / cancel
	Count          int    `gorm:"column:count;type:int" json:"count"`                                             // 当前人数
	Limit          int    `gorm:"column:limit;type:int" json:"limit"`                                             // 超员阈值
	Timestamp      int64  `gorm:"column:timestamp;type:bigint" json:"timestamp"`                                  // 上报时间戳（Unix 毫秒）
	CreateTime     string `gorm:"column:create_time;type:varchar(32)" json:"createTime"`                          // 入库时间
}

func (Alarm) TableName() string { return "alarm" }
