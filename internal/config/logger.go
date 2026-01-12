package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logLevelMap = map[string]zapcore.Level{
	"debug":   zapcore.DebugLevel,
	"info":    zapcore.InfoLevel,
	"warn":    zapcore.WarnLevel,
	"warning": zapcore.WarnLevel,
	"error":   zapcore.ErrorLevel,
	"fatal":   zapcore.FatalLevel,
}

type LoggerService struct {
	Logger       *zap.Logger
	lumberWriter *lumberjack.Logger // 关键：保存 lumberjack 实例
}

func NewLogger(i do.Injector) (*LoggerService, error) {
	cfg := do.MustInvoke[*Config](i)

	// 创建 lumberjack writer（关键：保存引用）
	lumberWriter := &lumberjack.Logger{
		Filename:   getLogFilePath(cfg),
		MaxSize:    cfg.Log.FileMaxSize,    // MB，0 表示不限制
		MaxBackups: cfg.Log.FileMaxBackups, // 0 表示不限制备份数
		MaxAge:     cfg.Log.FileMaxAge,     // 天，0 表示不限制年龄
		Compress:   cfg.Log.Compress,
		LocalTime:  true, // 按本地时间切割（实现按天拆分）
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(lumberWriter.Filename), 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
	}

	writeSyncer := zapcore.AddSync(lumberWriter)

	// 获取编码器（统一配置）
	encoder := getEncoder(cfg)

	// 动态解析日志级别
	level := parseLogLevel(cfg.Log.Level)

	// 核心 Core
	var core zapcore.Core

	if cfg.App.Env == "production" {
		// 生产环境：仅写入文件（JSON 格式，高性能）
		core = zapcore.NewCore(encoder, writeSyncer, level)
	} else {
		// 开发环境：同时输出到控制台（彩色）和文件
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		consoleWriter := zapcore.Lock(os.Stdout) // gin.DefaultWriter 可能并发不安全，用 os.Stdout 更稳

		core = zapcore.NewTee(
			zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel),          // 文件：JSON
			zapcore.NewCore(consoleEncoder, consoleWriter, zapcore.DebugLevel), // 控制台：彩色可读
		)
	}
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	return &LoggerService{
		Logger:       logger,
		lumberWriter: lumberWriter,
	}, nil
}

// getEncoder 获取编码器（根据配置动态选择）
func getEncoder(cfg *Config) zapcore.Encoder {
	if cfg.App.Env == "production" || cfg.Log.Format == "json" {
		// 生产或强制 JSON：高性能结构化
		config := zap.NewProductionEncoderConfig()
		config.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncodeLevel = zapcore.CapitalLevelEncoder
		config.EncodeCaller = zapcore.ShortCallerEncoder
		return zapcore.NewJSONEncoder(config)
	}

	// 开发环境默认：控制台彩色
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

// getLogFilePath 构建完整日志路径
func getLogFilePath(cfg *Config) string {
	// 默认路径
	appName := cfg.App.Name
	if appName == "" {
		appName = "app"
	}
	path := cfg.Log.Path
	if path == "" {
		path = "./logs"
	}
	return filepath.Join(path, appName+"/"+appName+".log")
}

func parseLogLevel(levelStr string) zapcore.Level {
	// 统一转小写，支持大小写不敏感
	if level, ok := logLevelMap[strings.ToLower(strings.TrimSpace(levelStr))]; ok {
		return level
	}
	fmt.Printf("警告: 无效的日志级别 '%s'，使用默认 info\n", levelStr)
	return zapcore.InfoLevel
}

func (s *LoggerService) Shutdown() error {
	fmt.Println("正在关闭日志文件...")
	// 1. 刷 zap 缓冲区
	_ = s.Logger.Sync() // 直接忽略错误

	// 2. 关闭 lumberjack 文件句柄（关键！）
	if s.lumberWriter != nil {
		_ = s.lumberWriter.Close()
	}
	fmt.Println(" ✅ 日志文件已关闭")
	return nil
}
