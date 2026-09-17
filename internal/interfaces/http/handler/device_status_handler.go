/**
@Time : 2026/09/17 15:53
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package handler

import (
	"context"
	devicestatusapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/devicestatus"
	"github.com/BinKuoLuo-go/emergency-center-go/pkg/response"
	"github.com/gin-gonic/gin"
	"time"
)

type DeviceStatusHandler struct {
	svc *devicestatusapp.Service
}

func NewDeviceStatusHandler(svc *devicestatusapp.Service) *DeviceStatusHandler {
	return &DeviceStatusHandler{svc: svc}
}

// Latest 查询设备最新资源状态。
func (h *DeviceStatusHandler) Latest(c *gin.Context) {
	deviceID := c.Query("deviceId")
	if deviceID == "" {
		response.Error(c, response.CodeBadReq, "deviceId 不能为空")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	d, err := h.svc.LatestStatus(ctx, deviceID)
	if err != nil {
		response.Error(c, response.CodeError, err.Error())
		return
	}
	response.OK(c, gin.H{"deviceId": deviceID, "status": d})
}

// List 查询所有已知设备的 deviceId 列表。
func (h *DeviceStatusHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	list, err := h.svc.ListDevices(ctx)
	if err != nil {
		response.Error(c, response.CodeError, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list})
}

// Online 查询设备在线状态。
func (h *DeviceStatusHandler) Online(c *gin.Context) {
	deviceID := c.Query("deviceId")
	if deviceID == "" {
		response.Error(c, response.CodeBadReq, "deviceId 不能为空")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	info, err := h.svc.Online(ctx, deviceID)
	if err != nil {
		response.Error(c, response.CodeError, err.Error())
		return
	}
	response.OK(c, gin.H{
		"deviceId": deviceID,
		"online":   info != nil,
		"info":     info,
	})
}
