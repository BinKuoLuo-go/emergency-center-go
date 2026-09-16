/**
@Time : 2026/09/16 14:33
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package persistence

import (
	"context"
	"sort"
	"strings"

	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/snapshot"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
)

// imageExts 视为图片的对象后缀（小写）。
var imageExts = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {},
	".webp": {}, ".gif": {}, ".bmp": {},
}

// 边缘端上传到 MinIO 的对象键结构（见边缘端 pkg/minio 上传逻辑）：
//
//	{deviceID}/composite/{company}/{日期文件夹}/{文件名}
//	例：edge-01/composite/某某公司/20260916/composite_某某公司_1694857600000.jpg
//
// 其中「日期文件夹」为 20060102 格式（如 20260916），且位于 company 之后，
// 因此无法仅靠前缀精确到日期，需要在列出后按日期段二次过滤。
const objectTypeComposite = "composite"

// SnapshotRepository 快照查询仓储，依赖对象存储端口。
// storeProvider 提供「当前」对象存储实例，保证对象存储配置热更新后仍取到最新实现。
type SnapshotRepository struct {
	storeProvider func() port.ObjectStore
}

func NewSnapshotRepository(storeProvider func() port.ObjectStore) *SnapshotRepository {
	return &SnapshotRepository{storeProvider: storeProvider}
}

// ListByDeviceAndDate 按设备 + 日期文件夹分页查询快照。
func (r *SnapshotRepository) ListByDeviceAndDate(ctx context.Context, q snapshot.Query) (*snapshot.Page, error) {
	store := r.storeProvider()
	prefix := buildSnapshotPrefix(q.DeviceID, q.Company)

	objects, err := store.List(ctx, prefix)
	if err != nil {
		return nil, err
	}

	list := make([]snapshot.Snapshot, 0, len(objects))
	for _, o := range objects {
		if !isImageKey(o.Key) {
			continue
		}
		if !matchDateFolder(o.Key, q.Date) {
			continue
		}
		list = append(list, snapshot.Snapshot{
			ID:          o.Key, // 以对象 key 作为唯一标识
			DeviceID:    q.DeviceID,
			ParkName:    parseCompany(o.Key),
			StoragePath: o.Key,
			CaptureTime: o.LastModified.Format("2006-01-02 15:04:05"),
			UpdateTime:  o.LastModified.UnixMilli(),
			Size:        o.Size,
		})
	}

	// 按最后修改时间倒序（新图在前）
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].UpdateTime > list[j].UpdateTime
	})

	total := int64(len(list))
	page, pageSize := normalizePage(q.Page, q.PageSize)
	start := (page - 1) * pageSize
	if start > int(total) {
		start = int(total)
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}

	return &snapshot.Page{
		List:     list[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// buildSnapshotPrefix 构造对象前缀。
// 前缀到类型 + 可选公司层级，日期在其后，需单独过滤：
//
//	deviceID 有值 → "{deviceID}/composite/"
//	company  有值 → "{deviceID}/composite/{company}/"
func buildSnapshotPrefix(deviceID, company string) string {
	deviceID = strings.Trim(deviceID, "/")
	company = strings.Trim(company, "/")

	var b strings.Builder
	if deviceID != "" {
		b.WriteString(deviceID)
		b.WriteByte('/')
	}
	b.WriteString(objectTypeComposite)
	b.WriteByte('/')
	if company != "" {
		b.WriteString(company)
		b.WriteByte('/')
	}
	return b.String()
}

// matchDateFolder 判断对象 key 的日期文件夹段是否等于指定日期。
// key 形如 {deviceID}/composite/{company}/{date}/{file}，日期为倒数第二段。
func matchDateFolder(key, date string) bool {
	if date == "" {
		return true
	}
	parts := strings.Split(strings.Trim(key, "/"), "/")
	if len(parts) < 4 {
		return false
	}
	return parts[len(parts)-2] == date
}

// parseCompany 从对象 key 中解析公司/园区名。
// key 形如 {deviceID}/composite/{company}/{date}/{file}，公司为第 3 段（下标 2）。
func parseCompany(key string) string {
	parts := strings.Split(strings.Trim(key, "/"), "/")
	if len(parts) >= 4 {
		return parts[2]
	}
	return ""
}

func isImageKey(key string) bool {
	lower := strings.ToLower(key)
	for ext := range imageExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}
