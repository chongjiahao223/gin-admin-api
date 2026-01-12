package cmd

import (
	"context"
	"errors"
	"gin-api/internal/config"
	"gin-api/internal/injector"
	"gin-api/internal/middleware"
	"gin-api/internal/router"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "启动 HTTP API 服务",
	Run: func(cmd *cobra.Command, args []string) {
		apiMain()
	},
}

func apiMain() {
	// 初始化 DI 容器
	container := injector.SetupInjector()
	// 获取核心依赖
	cfg := do.MustInvoke[*config.Config](container)
	loggerService := do.MustInvoke[*config.LoggerService](container)
	logger := loggerService.Logger
	_ = do.MustInvoke[*config.DBService](container)
	_ = do.MustInvoke[*config.RedisService](container)

	// 生产环境切换 Gin 模式
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Gin 引擎
	engine := gin.New()
	// 注册全局中间键
	engine.Use(middleware.RecoveryMiddleware(container))
	engine.Use(middleware.TraceIDMiddleware())
	engine.Use(middleware.LoggerMiddleware(container))
	engine.Use(middleware.GlobalRateLimiter(100, 200)) // 全局限流 100 QPS，突发 200
	// IP 限流 每个 IP 10 QPS，突发 20，30 分钟清理一次
	ipLimiter := middleware.NewIPRateLimiter(10, 20, 30*time.Minute)
	engine.Use(ipLimiter.Limit())
	// 优雅关闭时停止清理协程
	defer ipLimiter.Stop()

	// 注册路由
	router.SetupRoutes(engine, container)
	// 启动服务
	srv := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: engine,
	}

	go func() {
		logger.Info("服务器启动", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 7. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器强制关闭", zap.Error(err))
	}

	// 关闭 DI 容器资源（如 DB、Redis）
	_ = container.Shutdown()

	logger.Info("服务器已退出")
	os.Exit(0)
}
