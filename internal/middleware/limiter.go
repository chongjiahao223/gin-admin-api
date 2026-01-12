package middleware

import (
	"gin-api/internal/types"
	"gin-api/internal/utils"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// GlobalRateLimiter 全局令牌桶限流（所有请求共享）
func GlobalRateLimiter(r rate.Limit, b int) gin.HandlerFunc {
	limiter := rate.NewLimiter(r, b)
	return func(c *gin.Context) {
		if err := limiter.Wait(c.Request.Context()); err != nil {
			// 超过限流，返回 429
			utils.FailWithStatus(c, http.StatusTooManyRequests, types.CodeRateLimited, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPRateLimiter 按 IP 独立限流（推荐生产使用）
type IPRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
	cleanup  *time.Ticker // 可选：定期清理过期 IP
}

// NewIPRateLimiter 创建按 IP 限流器
func NewIPRateLimiter(r rate.Limit, b int, cleanupInterval time.Duration) *IPRateLimiter {
	lim := &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}

	// 可选：定期清理长时间不活跃的 IP 限流器（防止内存泄漏）
	if cleanupInterval > 0 {
		lim.cleanup = time.NewTicker(cleanupInterval)
		go lim.cleanupLoop(cleanupInterval * 2) // 清理超过 2 倍间隔未访问的 IP
	}

	return lim
}

// GetLimiter 获取或创建对应 IP 的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.limiters[ip]
	i.mu.RUnlock()

	if exists {
		return limiter
	}

	// 不存在则创建
	i.mu.Lock()
	defer i.mu.Unlock()

	// 双检查（Double-Checked Locking）
	if limiter, exists := i.limiters[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(i.r, i.b)
	i.limiters[ip] = limiter
	return limiter
}

// Limit 返回限流中间件
func (i *IPRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := i.GetLimiter(ip)

		// 使用 WaitN(1) 而不是 Allow()，更精确，支持上下文取消
		ctx := c.Request.Context()
		if err := limiter.WaitN(ctx, 1); err != nil {
			utils.FailWithStatus(c, http.StatusTooManyRequests, types.CodeRateLimited, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanupLoop 定期清理长时间未访问的 IP 限流器（防止内存无限增长）
func (i *IPRateLimiter) cleanupLoop(expire time.Duration) {
	for range i.cleanup.C {
		i.mu.Lock()
		for ip, limiter := range i.limiters {
			// 判断是否长时间未使用（通过 Reserve().DelayFrom(now)）
			if r := limiter.Reserve(); r.Delay() > expire {
				delete(i.limiters, ip)
				r.Cancel() // 取消预留
			}
		}
		i.mu.Unlock()
	}
}

// Stop 停止清理协程（优雅关闭时调用）
func (i *IPRateLimiter) Stop() {
	if i.cleanup != nil {
		i.cleanup.Stop()
	}
}
