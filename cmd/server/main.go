/**
@Time : 2026/09/16 15:23
@Author: FangYao( 方少、)
@Description:
@Email: fy20030315@163.com
*/

package main

import (
	"context"
	"flag"
	"fmt"
	alarmapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/alarm"
	devicestatusapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/devicestatus"
	objectstoreapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/objectstore"
	snapshotapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/snapshot"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/config"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/mysql"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/redis"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/interfaces/http/router"
	mqttsub "github.com/BinKuoLuo-go/emergency-center-go/internal/interfaces/mqtt"
	applog "github.com/BinKuoLuo-go/emergency-center-go/pkg/log"
	"github.com/gin-gonic/gin"
	"os"
	"os/signal"
	"syscall"
)

const version = "1.0.0"

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		applog.Fatalf("加载配置文件: %v", err)
	}

	if err := applog.Init(cfg.Log); err != nil {
		applog.Fatalf("初始化日志: %v", err)
	}

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := mysql.NewMysqlDB(cfg.Mysql)
	if err != nil {
		applog.Fatalf("init mysql: %v", err)
	}

	// redis
	var redisClient *redis.Client
	redisClient, err = redis.NewClient(cfg.Redis)
	if err != nil {
		applog.Warn("Redis 无法使用", "err", err)
	} else {
		defer redisClient.Close()
	}
	// 对象存储相关 minio
	objectStoreRepo := persistence.NewObjectStoreConfigRepository(db)
	objectStoreService := objectstoreapp.NewService(objectStoreRepo)

	// 当前配置的存储实现,配置热更新
	storeProvider := func() port.ObjectStore { return objectStoreService.Store() }

	// 快照相关
	snapshotRepo := persistence.NewSnapshotRepository(storeProvider)
	snapshotService := snapshotapp.NewService(snapshotRepo, storeProvider)

	// 告警相关
	alarmRepo := persistence.NewAlarmRepository(db)
	alarmService := alarmapp.NewService(alarmRepo)

	// 设备状态相关
	deviceStatusService := devicestatusapp.NewService(redisClient)

	// MQTT订阅器
	alarmSubscriber := mqttsub.NewSubscriber(cfg.Mqtt, alarmService, deviceStatusService)
	if err := alarmSubscriber.Start(); err != nil {
		applog.Warn("MQTT 订阅器启动失败", "err", err)
	} else {
		defer alarmSubscriber.Stop()
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := gin.New()
	router.Setup(r, router.Deps{
		ObjectStoreService:  objectStoreService,
		SnapshotService:     snapshotService,
		AlarmService:        alarmService,
		DeviceStatusService: deviceStatusService,
	})

	// 关闭信号监听
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel() // 触发上下文取消，关闭后台任务
	}()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	applog.Info("server start",
		"version", version,
		"http", addr,
	)
	if err := r.Run(addr); err != nil {
		applog.Fatalf("server exit: %v", err)
	}

}
