package router

import (
	"gin-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

func ApiRouter(r *gin.RouterGroup, container do.Injector) {
	r.GET("/health", func(c *gin.Context) {
		utils.Success(c, "api/health")
	})
}
