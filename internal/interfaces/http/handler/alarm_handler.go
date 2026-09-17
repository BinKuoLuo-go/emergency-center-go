/**
@Time : 2026/09/17 15:47
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package handler

import (
	"context"
	alarmapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/alarm"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/domain/alarm"
	"github.com/BinKuoLuo-go/emergency-center-go/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type AlarmHandler struct {
	svc *alarmapp.Service
}

func NewAlarmHandler(svc *alarmapp.Service) *AlarmHandler {
	return &AlarmHandler{svc: svc}
}

// List 分页查询告警记录。
func (h *AlarmHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := h.svc.List(ctx, alarm.Query{
		CompanyCode: c.Query("companyCode"),
		DeviceID:    c.Query("deviceId"),
		AlarmStatus: c.Query("alarmStatus"),
		StartTime:   c.Query("startTime"),
		EndTime:     c.Query("endTime"),
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		response.Error(c, response.CodeError, err.Error())
		return
	}
	response.OK(c, result)
}
