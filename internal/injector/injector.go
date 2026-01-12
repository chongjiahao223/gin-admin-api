package injector

import (
	"gin-api/internal/config"

	"github.com/samber/do/v2"
)

func SetupInjector() do.Injector {
	injector := do.New()
	// 配置层(必须最先注册)
	do.Provide(injector, config.NewConfig)
	do.Provide(injector, config.NewLogger)
	do.Provide(injector, config.NewDB)
	do.Provide(injector, config.NewRedis)
	do.Provide(injector, config.NewQueue)
	return injector
}
