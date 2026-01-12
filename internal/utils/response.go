package utils

import (
	"gin-api/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": types.CodeSuccess,
		"msg":  "success",
		"data": data,
	})
	c.Abort() // 防止后续代码继续执行
}
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
	})
	c.Abort()
}
func FailWithStatus(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, gin.H{
		"code": code,
		"msg":  msg,
	})
	c.Abort()
}
