package cron

import (
	"gin-api/internal/config"

	"github.com/robfig/cron/v3"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// RegisterTasks 统一注册所有定时任务
func RegisterTasks(c *cron.Cron, i do.Injector) {
	logger := do.MustInvoke[*config.LoggerService](i).Logger
	var err error
	//_, err = c.AddFunc("@every 10s", tasks.NewExampleTask(i).Run)
	if err != nil {
		logger.Fatal("注册 Example 任务失败", zap.Error(err))
	}
}
