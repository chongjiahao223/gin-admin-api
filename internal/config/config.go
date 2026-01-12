package config

import (
	"fmt"

	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Asynq    AsynqConfig    `mapstructure:"asynq"`
	Asynqmon AsynqmonConfig `mapstructure:"asynqmon"`
	Log      LogConfig      `mapstructure:"log"`
}
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
	Mode    string `mapstructure:"mode"`
}
type ServerConfig struct {
	Port         int `mapstructure:"port"`
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
	IdleTimeout  int `mapstructure:"idle_timeout"`
}
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parseTime"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"maxIdleConns"`
	MaxOpenConns    int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime int    `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime int    `mapstructure:"connMaxIdleTime"`
}
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	PoolTimeout  int    `mapstructure:"pool_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}
type AsynqConfig struct {
	RedisHost         string         `mapstructure:"redis_host"`
	RedisPort         int            `mapstructure:"redis_port"`
	RedisPassword     string         `mapstructure:"redis_password"`
	RedisDB           int            `mapstructure:"redis_db"`
	WorkerConcurrency int            `mapstructure:"worker_concurrency"`
	Queues            map[string]int `mapstructure:"queues"`
}
type AsynqmonConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	HttpAddr int    `mapstructure:"http_addr"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}
type LogConfig struct {
	Level          string `mapstructure:"level"`
	Format         string `mapstructure:"format"`
	Output         string `mapstructure:"output"`
	Path           string `mapstructure:"path"`
	FileMaxSize    int    `mapstructure:"file_max_size"`
	FileMaxBackups int    `mapstructure:"file_max_backups"`
	FileMaxAge     int    `mapstructure:"file_max_age"`
	Compress       bool   `mapstructure:"compress"`
	LocalTime      bool   `mapstructure:"local_time"`
}

func NewConfig(i do.Injector) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	// 设置默认配置
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析到结构体
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	fmt.Printf("配置文件已加载: %s\n", viper.ConfigFileUsed())

	return &cfg, nil
}
func setDefaults() {
	viper.SetDefault("app.name", "gin-api")
	viper.SetDefault("app.version", "v1.0.0")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 60)
}
