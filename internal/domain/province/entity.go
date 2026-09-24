/**
@Time : 2026/09/24 11:43
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package province

import (
	"context"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/alarm"
	"time"
)

// GatherAlarm 省平台 装置聚集报警数据业务字段
type GatherAlarm struct {
	ID          string `json:"id"`          // 主键，36/32 位 UUID，用于增量同步、幂等、报警与销警对应
	CompanyCode string `json:"companyCode"` // 企业编码
	AlarmType   string `json:"alarmType"`   // 报警类型：1=生产车间(含装置)超2人 ...
	AlarmNumber int    `json:"alarmNumber"` // 报警人数
	AlarmArea   string `json:"alarmArea"`   // 聚集报警区域
	Describe    string `json:"describe"`    // 报警说明（可选）
	AlarmStatus string `json:"alarmStatus"` // 1 报警；0 销警
	AlarmTime   string `json:"alarmTime"`   // yyyy-MM-dd HH:mm:ss
	AlarmPic    string `json:"alarmPic"`    // 报警图片 base64（可选）
	AlarmDevice string `json:"alarmDevice"` // 1 视频监控；2 道闸
	Deleted     string `json:"deleted"`     // 0 正常；1 已删除
}

type Client interface {
	Report(ctx context.Context, batchID string, items []GatherAlarm) error
}

// FromAlarm 将边缘端告警转换为省平台业务格式
func FromAlarm(a *alarm.Alarm) GatherAlarm {
	status := a.AlarmStatus
	if status != "1" {
		status = "0"
	}
	return GatherAlarm{
		ID:          a.ID,
		CompanyCode: a.CompanyCode,
		AlarmType:   "1",
		AlarmNumber: a.Count,
		AlarmArea:   "",
		Describe:    "超员告警",
		AlarmStatus: status,
		AlarmTime:   a.AlarmTime,
		AlarmDevice: "1",
		Deleted:     "0",
	}
}

// Credential 省平台访问凭据
type Credential struct {
	Token    string    // Authorization 请求头使用的 token
	AESKey   string    // AES-GCM 密钥（hex）
	AESIV    string    // AES-GCM IV（hex）
	ExpireAt time.Time // 过期时间
}

// TokenProvider 凭据提供端口
type TokenProvider interface {
	// Get 返回当前有效的凭据（。
	Get(ctx context.Context) (*Credential, error)
}
