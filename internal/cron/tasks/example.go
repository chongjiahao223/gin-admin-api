package tasks

import (
	"gin-api/internal/config"

	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

type ExampleTask struct {
	logger *zap.Logger
}

func NewExampleTask(i do.Injector) *ExampleTask {
	return &ExampleTask{
		logger: do.MustInvoke[*config.LoggerService](i).Logger,
	}
}
func (t *ExampleTask) Run() {
	t.logger.Info("开始执行 Example 定时任务")

	t.logger.Info("Example 任务执行成功")
}
