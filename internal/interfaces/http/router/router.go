/**
@Time : 2026/09/16 14:31
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package router

import (
	alarmapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/alarm"
	devicestatusapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/devicestatus"
	objectstoreapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/objectstore"
	snapshotapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/snapshot"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/interfaces/http/handler"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	ObjectStoreService  *objectstoreapp.Service
	SnapshotService     *snapshotapp.Service
	AlarmService        *alarmapp.Service
	DeviceStatusService *devicestatusapp.Service
}

func Setup(r *gin.Engine, deps Deps) {
	r.Use(gin.Recovery())              // 使用 Gin 自带的恢复中间件，捕获 panic 防止程序崩溃
	r.Use(middleware.CORSMiddleware()) // 跨域中间件

	objectStoreHandler := handler.NewObjectStoreHandler(deps.ObjectStoreService)
	snapshotHandler := handler.NewSnapshotHandler(deps.SnapshotService)
	alarmHandler := handler.NewAlarmHandler(deps.AlarmService)
	deviceStatusHandler := handler.NewDeviceStatusHandler(deps.DeviceStatusService)
	apiGroup := r.Group("")
	{
		serverAPI := apiGroup.Group("/server")
		{
			serverAPI.GET("/object_store_config", objectStoreHandler.Get)
			serverAPI.POST("/object_store_config/save", objectStoreHandler.Save)
			serverAPI.GET("/object_store_config/health", objectStoreHandler.Health)
		}
		snapshotAPI := apiGroup.Group("/snapshot")
		{
			// 快照 按deviceId 日期文件夹分页查询图片
			snapshotAPI.GET("/snapshots", snapshotHandler.List)

			// 图片流式代理 /server/file/{objectKey}，由中心平台转发图片内容
			snapshotAPI.GET("/file/*key", snapshotHandler.Image)
		}
		alarmAPI := apiGroup.Group("/alarm")
		{
			// 告警记录分页查询报警、销警事件
			alarmAPI.GET("/list", alarmHandler.List)
		}
		deviceStatusAPI := apiGroup.Group("/device-status")
		{
			// 设备最新资源状态查询 Redis
			deviceStatusAPI.GET("/latest", deviceStatusHandler.Latest)
		}
		deviceAPI := apiGroup.Group("/device")
		{
			// 设备在线状态查询 Redis
			deviceAPI.GET("/online", deviceStatusHandler.Online)
			// 设备列表查询 Redis
			deviceAPI.GET("/list", deviceStatusHandler.List)
		}

	}

}
