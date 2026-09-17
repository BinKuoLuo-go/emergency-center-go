/**
@Time : 2026/09/17 15:50
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package devicestatus

// DeviceStatus 设备资源状态
type DeviceStatus struct {
	ID          string  `json:"id"`          // 消息唯一 ID（UUID）
	DeviceID    string  `json:"deviceId"`    // 边缘设备 ID
	DeviceName  string  `json:"deviceName"`  // 设备名称（主机名）
	Status      string  `json:"status"`      // online / offline
	ReportType  string  `json:"reportType"`  // status
	CpuPercent  float64 `json:"cpuPercent"`  // CPU 总使用率（%）
	MemPercent  float64 `json:"memPercent"`  // 内存使用率（%）
	MemUsed     uint64  `json:"memUsed"`     // 内存已用（字节）
	MemTotal    uint64  `json:"memTotal"`    // 内存总量（字节）
	DiskPercent float64 `json:"diskPercent"` // 磁盘使用率（%）
	DiskUsed    uint64  `json:"diskUsed"`    // 磁盘已用（字节）
	DiskTotal   uint64  `json:"diskTotal"`   // 磁盘总量（字节）
	Uptime      uint64  `json:"uptime"`      // 系统运行时长（秒）
	Hostname    string  `json:"hostname"`    // 主机名
	OS          string  `json:"os"`          // 操作系统：windows / linux
	Platform    string  `json:"platform"`    // 平台描述
	ReportTime  string  `json:"reportTime"`  // 上报时间 yyyy-MM-dd HH:mm:ss
	Timestamp   int64   `json:"timestamp"`   // 上报时间戳（Unix 毫秒）
}

// OnlineInfo 设备在线状态Redis存储的JSON
type OnlineInfo struct {
	DeviceID          string `json:"deviceId"`          // 边缘设备 ID
	CompanyCode       string `json:"companyCode"`       // 最近心跳的公司编码
	CompanyName       string `json:"companyName"`       // 最近心跳的公司名称
	LastHeartbeatTime string `json:"lastHeartbeatTime"` // 最近心跳时间 yyyy-MM-dd HH:mm:ss
	Timestamp         int64  `json:"timestamp"`         // 最近心跳时间戳（Unix 毫秒）
}
