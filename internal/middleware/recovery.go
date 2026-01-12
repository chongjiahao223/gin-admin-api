package middleware

import (
	"gin-api/internal/config"
	"gin-api/internal/types"
	"gin-api/internal/utils"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// RecoveryMiddleware 全局 Panic 恢复中间件（生产必备）
func RecoveryMiddleware(i do.Injector) gin.HandlerFunc {
	logger := do.MustInvoke[*config.LoggerService](i).Logger

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 获取 trace_id（从上下文）
				traceID := GetTraceID(c)

				// 记录详细错误日志（带堆栈）
				logger.Error("PANIC RECOVERED",
					zap.String("trace_id", traceID),
					zap.Any("error", r),
					zap.String("stack", string(debug.Stack())),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
				)

				// 统一返回服务器错误（不暴露细节）
				utils.FailWithStatus(c, http.StatusInternalServerError, types.CodeServerError, "服务器内部错误")

				// 中止请求
				c.Abort()
			}
		}()

		c.Next()
	}
}
