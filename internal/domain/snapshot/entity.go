/**
@Time : 2026/09/16 10:36
@Author: FangYao( 方少、)
@Description:  边缘设备快照上传
@Email: fy20030315@163.com
*/

package snapshot

import "context"

type Snapshot struct {
	ID          string `json:"id"`          // 唯一 id
	DeviceID    string `json:"deviceId"`    // 边缘设备 ID
	ParkName    string `json:"parkName"`    // 园区/公司名
	StoragePath string `json:"storagePath"` // 对象存储路径
	CaptureTime string `json:"captureTime"` // 抓拍时间
	UpdateTime  int64  `json:"updateTime"`  // 上传时间（Unix 毫秒）
	Size        int64  `json:"size"`        // 文件大小
	URL         string `json:"url"`         // 图片访问直链（预签名，可能为空）
}

// Repository 快照列表查询端口。
type Repository interface {
	// ListByDeviceAndDate 按设备 + 日期文件夹分页查询快照列表。
	ListByDeviceAndDate(ctx context.Context, q Query) (*Page, error)
}

// Query 快照查询条件。
type Query struct {
	DeviceID string // 边缘设备 ID
	Date     string // 日期文件夹名（如 20260916）
	Company  string // 公司/园区名（可选，用于精确前缀，提升查询效率）
	Page     int    // 页码，从 1 开始
	PageSize int    // 每页数量
}

// Page 分页结果。
type Page struct {
	List     []Snapshot `json:"list"`     // 当前页数据
	Total    int64      `json:"total"`    // 满足条件的总数
	Page     int        `json:"page"`     // 当前页码
	PageSize int        `json:"pageSize"` // 每页数量
}
