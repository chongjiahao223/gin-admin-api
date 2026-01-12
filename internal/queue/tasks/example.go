package tasks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const TypeExample = "example"

// ExamplePayload 任务参数
type ExamplePayload struct {
	Name string `json:"name"`
}

func NewExamplePayload(name string) ([]byte, error) {
	p := &ExamplePayload{Name: name}
	return json.Marshal(p)
}

// ExampleTask 任务结构体（注入 logger）
type ExampleTask struct {
	logger *zap.Logger
}

// NewExampleTask 通过 DI 容器创建任务实例（注入 logger）
func NewExampleTask(l *zap.Logger) *ExampleTask {
	return &ExampleTask{
		logger: l, // ← 通过 DI 获取自定义 logger
	}
}
func (t *ExampleTask) ProcessExample(ctx context.Context, task *asynq.Task) error {

	var p ExamplePayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return err
	}

	t.logger.Info("开始执行 "+TypeExample,
		zap.String("name", p.Name),
		zap.String("task_id", task.Type()),
	)

	// 模拟耗时操作（实际替换成你的业务逻辑）
	time.Sleep(10 * time.Second)

	t.logger.Info("执行完成 "+TypeExample, zap.String("name", p.Name))
	return nil
}
