package middleware

import (
	"context"
	"gin-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// TraceIDKey 上下文 key（使用私有 type 避免冲突）
type traceIDKey struct{}

// TraceIDMiddleware 生成或读取 X-Trace-ID，并注入上下文
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Header 读取 X-Trace-ID（支持大小写变体）
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = c.GetHeader("X-Trace-Id") // 兼容小写
		}

		// 2. 如果没有，生成一个短 TraceID
		if traceID == "" {
			traceID = utils.GenerateShortTraceID() // 你的工具函数，推荐 16 字符 hex
		}

		// 3. 写入响应 Header（关键！便于网关/前端追踪）
		c.Header("X-Trace-ID", traceID)

		// 4. 注入 Gin Context（方便中间件/Handler 使用）
		c.Set("trace_id", traceID)

		// 5. 注入 Request Context（支持 context.WithValue 传播）
		ctx := context.WithValue(c.Request.Context(), traceIDKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetTraceID 从 context 获取 Trace ID（工具函数，强烈推荐在日志中使用）
func GetTraceID(c *gin.Context) string {
	// 优先从 Gin Context 取
	if id, exists := c.Get("trace_id"); exists {
		if str, ok := id.(string); ok {
			return str
		}
	}

	// 备选从 Request Context 取
	if id := c.Request.Context().Value(traceIDKey{}); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}

	return "unknown"
}
