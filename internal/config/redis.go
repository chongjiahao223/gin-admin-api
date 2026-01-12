package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

type RedisService struct {
	Client *redis.Client
}

func NewRedis(i do.Injector) (*RedisService, error) {
	cfg := do.MustInvoke[*Config](i)
	l := do.MustInvoke[*LoggerService](i).Logger

	// Redis 选项配置
	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,                                  // 支持密码
		DB:           cfg.Redis.DB,                                        // 选择 DB
		PoolSize:     cfg.Redis.PoolSize,                                  // 连接池大小（per CPU core）
		MinIdleConns: cfg.Redis.MinIdleConns,                              // 最小空闲连接
		MaxRetries:   cfg.Redis.MaxRetries,                                // 重试次数（推荐 3）
		DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,  // 连接超时
		ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,  // 读超时
		WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second, // 写超时
		PoolTimeout:  time.Duration(cfg.Redis.PoolTimeout) * time.Second,  // 获取连接超时
	}
	// 创建客户端
	client := redis.NewClient(options)

	// 测试连接（带超时上下文）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		l.Error("Redis 连接失败",
			zap.Error(err),
			zap.String("addr", options.Addr),
			zap.Int("db", options.DB),
		)
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	l.Info("Redis 连接成功",
		zap.String("addr", options.Addr),
		zap.Int("db", options.DB),
		zap.Int("pool_size", options.PoolSize),
		zap.Int("min_idle", options.MinIdleConns),
	)

	return &RedisService{Client: client}, nil
}

// Shutdown 优雅关闭 Redis 连接
func (s *RedisService) Shutdown() error {
	fmt.Println("正在关闭 Redis 连接...")

	if s.Client == nil {
		fmt.Println("Redis 未初始化，跳过关闭")
		return nil
	}

	if err := s.Client.Close(); err != nil {
		return fmt.Errorf("关闭 Redis 连接失败: %w", err)
	}

	fmt.Println(" ✅ Redis 连接已关闭")
	return nil
}
