package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/samber/do/v2"
)

type Queue struct {
	Client *asynq.Client
}

func NewQueue(i do.Injector) (*Queue, error) {
	cfg := do.MustInvoke[*Config](i)

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Asynq.RedisHost + ":" + strconv.Itoa(cfg.Asynq.RedisPort),
		Password: cfg.Asynq.RedisPassword,
		DB:       cfg.Asynq.RedisDB,
	})
	return &Queue{Client: client}, nil
}
func (s *Queue) Shutdown() error {
	fmt.Println("正在关闭 queue 连接...")
	if s.Client == nil {
		fmt.Println("queue 未初始化，跳过关闭")
		return nil
	}
	if err := s.Client.Close(); err != nil {
		return fmt.Errorf("关闭 queue 连接失败: %w", err)
	}

	fmt.Println(" ✅ queue 连接已关闭")
	return nil
}
func (s *Queue) Enqueue(ctx context.Context, taskType string, payload any, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(taskType, payloadBytes)

	// 默认选项
	defaultOpts := []asynq.Option{
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(30 * time.Minute),
	}

	// 合并用户传入的选项
	info, err := s.Client.Enqueue(task, append(defaultOpts, opts...)...)
	return info, err
}
