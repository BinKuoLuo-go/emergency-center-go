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
	objectstoreapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/objectstore"
	snapshotapp "github.com/BinKuoLuo-go/emergency-center-go/internal/application/snapshot"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/config"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/persistence/sqlite"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/infrastructure/port"
	"github.com/BinKuoLuo-go/emergency-center-go/internal/interfaces/http/router"
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
	db, err := sqlite.NewSqlite(cfg.Sqlite)
	if err != nil {
		applog.Fatalf("init mysql: %v", err)
	}

	objectStoreRepo := persistence.NewObjectStoreConfigRepository(db)
	objectStoreService := objectstoreapp.NewService(objectStoreRepo)

	// 对象存储实例提供者：始终取到「当前」配置的存储实现（支持配置热更新）
	storeProvider := func() port.ObjectStore { return objectStoreService.Store() }

	// 快照查询：仓储 + 应用服务
	snapshotRepo := persistence.NewSnapshotRepository(storeProvider)
	snapshotService := snapshotapp.NewService(snapshotRepo, storeProvider)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := gin.New()
	router.Setup(r, router.Deps{
		ObjectStoreService: objectStoreService,
		SnapshotService:    snapshotService,
	})

	// 优雅关闭信号监听
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
