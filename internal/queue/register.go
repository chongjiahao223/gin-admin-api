package queue

import (
	"gin-api/internal/queue/tasks"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func RegisterHandlers(mux *asynq.ServeMux, logger *zap.Logger) {
	// 所有任务类型集中注册
	handlers := []struct {
		Type string
		Func asynq.HandlerFunc
	}{
		{tasks.TypeExample, tasks.NewExampleTask(logger).ProcessExample},
	}

	for _, t := range handlers {
		mux.HandleFunc(t.Type, t.Func)
	}

	logger.Info("Asynq 处理器注册完成", zap.Int("handler_count", len(handlers)))
}
