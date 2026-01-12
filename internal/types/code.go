package types

// 业务错误码
const (
	CodeSuccess      int = 0    // 成功
	CodeInvalidParam int = 1001 // 参数错误
	CodeUnauthorized int = 1002 // 未登录
	CodeForbidden    int = 1003 // 权限不足
	CodeNotFound     int = 1004 // 资源不存在
	CodeExist        int = 1005 // 已存在
	CodeRateLimited  int = 1006 // 限流
	CodeServerError  int = 5000 // 服务器内部错误
)

// CodeMsg 错误码文本映射
var CodeMsg = map[int]string{
	CodeSuccess:      "success",
	CodeInvalidParam: "参数错误",
	CodeUnauthorized: "请先登录",
	CodeForbidden:    "没有权限",
	CodeNotFound:     "不存在",
	CodeExist:        "已存在",
	CodeServerError:  "服务器内部错误",
}

// GetCodeMsg 获取错误消息（默认返回 code 字符串）
func GetCodeMsg(code int) string {
	if msg, ok := CodeMsg[code]; ok {
		return msg
	}
	return "未知错误"
}
