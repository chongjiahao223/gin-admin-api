package router

import (
	"gin-api/internal/config"
	"gin-api/internal/queue/tasks"
	"gin-api/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/samber/do/v2"
)

func SetupRoutes(r *gin.Engine, container do.Injector) {
	// 全局中间件

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		queue := do.MustInvoke[*config.Queue](container).Client
		payload, _ := tasks.NewExamplePayload("cjh")
		_, err := queue.Enqueue(
			asynq.NewTask(tasks.TypeExample, payload),
			asynq.ProcessIn(10*time.Second),
			asynq.Queue("default"),
			asynq.MaxRetry(3),
			asynq.Timeout(30*time.Minute),
		)
		if err != nil {
			utils.Fail(c, 500, "任务入队失败")
			return
		}
		utils.Success(c, nil)
	})

	// API 路由
	api := r.Group("/api")
	ApiRouter(api, container)
}
