/**
@Time : 2026/09/16 14:31
@Author: FangYao( 方少、)
@Description: 快照查询 HTTP 接口
@Email: fy20030315@163.com
*/

package handler

import (
	"context"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	snapshotapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/snapshot"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/snapshot"
	"github.com/BinKuoLuo-go/emergency-center-go/pkg/response"

	"github.com/gin-gonic/gin"
)

type SnapshotHandler struct {
	svc *snapshotapp.Service
}

func NewSnapshotHandler(svc *snapshotapp.Service) *SnapshotHandler {
	return &SnapshotHandler{svc: svc}
}

// List 按设备 + 日期文件夹分页查询快照。
// GET /server/snapshots?deviceId=edge-01&date=20260916&company=某某公司&page=1&pageSize=20
func (h *SnapshotHandler) List(c *gin.Context) {
	deviceID := c.Query("deviceId")
	date := c.Query("date")
	if deviceID == "" || date == "" {
		response.Error(c, response.CodeBadReq, "deviceId 和 date 不能为空")
		return
	}

	company := c.Query("company")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := h.svc.ListByDeviceAndDate(ctx, snapshot.Query{
		DeviceID: deviceID,
		Date:     date,
		Company:  company,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, response.CodeError, err.Error())
		return
	}
	response.OK(c, result)
}

// Image 图片流式代理：由中心平台转发对象内容，避免预签名 URL 过长且过期。
// GET /server/file/{objectKey}
func (h *SnapshotHandler) Image(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	if key == "" {
		response.Error(c, response.CodeBadReq, "key 不能为空")
		return
	}

	rc, err := h.svc.OpenImage(c.Request.Context(), key)
	if err != nil {
		response.Error(c, response.CodeError, "读取图片失败: "+err.Error())
		return
	}
	defer rc.Close()

	c.Header("Content-Type", contentTypeByKey(key))
	c.Header("Cache-Control", "public, max-age=3600")
	_, _ = io.Copy(c.Writer, rc)
}

func contentTypeByKey(key string) string {
	switch strings.ToLower(filepath.Ext(key)) {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	default:
		return "image/jpeg"
	}
}
