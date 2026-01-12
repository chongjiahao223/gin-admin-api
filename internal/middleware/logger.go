package middleware

import (
	"bytes"
	"encoding/json"
	"gin-api/internal/config"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// responseWriter 自定义响应写入器，用于捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LoggerMiddleware(i do.Injector) gin.HandlerFunc {
	logger := do.MustInvoke[*config.LoggerService](i).Logger

	return func(c *gin.Context) {
		start := time.Now()

		// 获取 Trace ID
		traceID := GetTraceID(c)

		// 读取并缓存请求体
		var reqBody any
		var rawReqBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			rawReqBody, _ = io.ReadAll(c.Request.Body)
			// 恢复 Body 供后续 Handler 使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(rawReqBody))

			if json.Valid(rawReqBody) {
				_ = json.Unmarshal(rawReqBody, &reqBody)
			} else {
				reqBody = string(rawReqBody)
			}
		}

		reqLogger := logger.With(zap.String("trace_id", traceID))
		reqLogger.Info("HTTP Request Received",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Any("request_body", reqBody),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)

		blw := &responseWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)

		var respBody any
		respBytes := blw.body.Bytes()
		if json.Valid(respBytes) {
			_ = json.Unmarshal(respBytes, &respBody)
		} else {
			respBody = string(respBytes)
		}

		reqLogger.Info("HTTP Response Sent",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Any("response_body", respBody),
		)
	}
}
