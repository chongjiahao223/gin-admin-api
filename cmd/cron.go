package cmd

import (
	"fmt"
	"gin-api/internal/config"
	cronR "gin-api/internal/cron"
	"gin-api/internal/injector"
	"gin-api/internal/queue"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/robfig/cron/v3"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "启动 Cron Job 服务",
	Run: func(cmd *cobra.Command, args []string) {
		cronMain()
	},
}

func cronMain() {
	// 初始化 DI 容器
	container := injector.SetupInjector()

	// 获取核心依赖
	cfg := do.MustInvoke[*config.Config](container)
	loggerService := do.MustInvoke[*config.LoggerService](container)
	logger := loggerService.Logger

	// 创建 Cron 调度器（支持秒级任务）
	c := cron.New(
		cron.WithSeconds(), // 支持秒级（如每30秒）
		cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger), // 防止任务重叠执行
			cron.Recover(cron.DefaultLogger),            // 捕获任务 panic
		),
	)

	//  注册所有定时任务
	cronR.RegisterTasks(c, container)

	// 启动 Cron
	c.Start()
	logger.Info("Cron Job 服务已启动，所有定时任务已注册")

	// 启动 Asynq Worker
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Asynq.RedisHost + ":" + strconv.Itoa(cfg.Asynq.RedisPort),
			Password: cfg.Asynq.RedisPassword,
			DB:       cfg.Asynq.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.Asynq.WorkerConcurrency,
			Queues:      cfg.Asynq.Queues,
		},
	)

	mux := asynq.NewServeMux()

	// 注册所有任务处理器（集中管理）
	queue.RegisterHandlers(mux, logger)

	// 启动 Worker（异步）
	go func() {
		if err := srv.Run(mux); err != nil {
			logger.Fatal("Asynq Worker 启动失败", zap.Error(err))
		}
	}()
	logger.Info("Asynq Worker 已启动")

	fmt.Printf("\nredis-db:%d\n", cfg.Asynq.RedisDB)
	fmt.Printf("\nredis-Addr:%s\n", cfg.Asynq.RedisHost+":"+strconv.Itoa(cfg.Asynq.RedisPort))

	if cfg.Asynqmon.Enabled {
		go func() {
			opts := asynqmon.Options{
				RootPath: "/asynqmon",
				RedisConnOpt: asynq.RedisClientOpt{
					Addr: cfg.Asynq.RedisHost + ":" + strconv.Itoa(cfg.Asynq.RedisPort),
					DB:   1,
				},
			}

			http.Handle("/asynqmon/", asynqmon.New(opts))
			addr := ":" + strconv.Itoa(cfg.Asynqmon.HttpAddr)

			logger.Info("Asynqmon Web UI 已启动", zap.String("addr", "http://localhost"+addr+"/asynqmon/"))
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Error("Asynqmon 启动失败", zap.Error(err))
			}
		}()
	}

	// 阻塞主线程，等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭 Cron Job 服务...")

	// 停止 Cron（等待正在运行的任务完成，最多等30秒）
	ctx := c.Stop()
	select {
	case <-ctx.Done():
		logger.Info("所有定时任务已优雅停止")
	case <-time.After(30 * time.Second):
		logger.Warn("任务关闭超时，已强制停止")
	}

	// 关闭 DI 容器资源（DB、Redis 等）
	if err := container.Shutdown(); err != nil {
		logger.Warn("DI 容器关闭出现非致命错误", zap.Error(err))
	}

	// 最后关闭日志
	defer func() {
		if err := loggerService.Shutdown(); err != nil {
			fmt.Printf("日志关闭失败: %v\n", err)
		}
	}()

	logger.Info("Cron Job 服务已安全退出")
	os.Exit(0)
}
