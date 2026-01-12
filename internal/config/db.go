package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBService struct {
	DB *gorm.DB
}

func NewDB(i do.Injector) (*DBService, error) {
	cfg := do.MustInvoke[*Config](i)
	l := do.MustInvoke[*LoggerService](i).Logger
	l.Info("数据库连接")

	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.Charset,
		cfg.Database.ParseTime,
		cfg.Database.Loc,
	)

	// GORM 配置
	gormConfig := &gorm.Config{
		// 开发环境：详细日志；生产环境：静默或仅错误
		Logger: logger.New(
			zap.NewStdLog(l), // 使用 zap 作为底层日志
			logger.Config{
				SlowThreshold:             time.Second, // 慢 SQL 阈值
				LogLevel:                  logger.Info, // 开发用 Info，生产可改为 Warn
				IgnoreRecordNotFoundError: true,        // 忽略 ErrRecordNotFound
				Colorful:                  cfg.App.Env != "production",
			},
		),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		l.Error("数据库连接失败", zap.Error(err), zap.String("dsn", maskPassword(dsn)))
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 获取底层 *sql.DB 用于连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库实例失败: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.ConnMaxIdleTime) * time.Second)

	// 可选：连接测试
	if err = sqlDB.Ping(); err != nil {
		l.Error("数据库 Ping 失败", zap.Error(err))
		return nil, fmt.Errorf("数据库 Ping 失败: %w", err)
	}

	l.Info("数据库连接成功",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.DBName),
		zap.Int("max_open", cfg.Database.MaxOpenConns),
		zap.Int("max_idle", cfg.Database.MaxIdleConns),
	)

	return &DBService{DB: db}, nil
}

// Shutdown 优雅关闭数据库连接
func (s *DBService) Shutdown() error {
	if s.DB == nil {
		fmt.Println("数据库未初始化，跳过关闭")
		return nil
	}

	fmt.Println("正在关闭数据库连接...")

	sqlDB, err := s.DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库实例失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	fmt.Println(" ✅ 数据库连接已关闭")
	return nil
}

// maskPassword 隐藏 DSN 中的密码（日志安全）
func maskPassword(dsn string) string {
	if i := strings.Index(dsn, ":@tcp"); i != -1 {
		start := strings.Index(dsn, "://")
		if start == -1 {
			start = 0
		}
		return dsn[:start+3] + "****" + dsn[i:]
	}
	return dsn
}
