/**
@Time : 2026/09/16 15:20
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package handler

import (
	"context"
	objectstoreapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/objectstore"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/model"
	"github.com/BinKuoLuo-go/emergency-center-go/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

type ObjectStoreHandler struct {
	svc *objectstoreapp.Service
}

func NewObjectStoreHandler(svc *objectstoreapp.Service) *ObjectStoreHandler {
	return &ObjectStoreHandler{svc: svc}
}

func (h *ObjectStoreHandler) Get(c *gin.Context) {
	row, err := h.svc.GetOrEmpty()
	if err != nil {
		response.Error(c, response.CodeError, err.Error())
		return
	}
	response.OK(c, row)
}

func (h *ObjectStoreHandler) Save(c *gin.Context) {
	var body model.ObjectStoreConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, response.CodeBadReq, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Save(&body); err != nil {
		response.Error(c, response.CodeError, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "保存成功", "provider": h.svc.Store().Provider()})
}

func (h *ObjectStoreHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	err := h.svc.Health(ctx)
	if err != nil {
		response.OK(c, gin.H{
			"ok":       false,
			"provider": h.svc.Store().Provider(),
			"error":    err.Error(),
		})
		return
	}
	response.OK(c, gin.H{"ok": true, "provider": h.svc.Store().Provider()})
}
