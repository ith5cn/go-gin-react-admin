package middleware

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	systemService "server/service/system"
	loggerInit "server/setup/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// maxOperLogBody 是操作日志记录请求体的上限；超出部分截断，避免大请求撑爆日志表。
const maxOperLogBody = 4 << 10

// sensitiveFieldPattern 匹配 JSON 里的凭证类字段，入库前统一脱敏。
var sensitiveFieldPattern = regexp.MustCompile(`(?i)"(password|oldPassword|newPassword|token|accessToken|refreshToken|secret|authorization)"\s*:\s*"[^"]*"`)

// OperLog 在鉴权后记录写操作，不采集高频且无副作用的查询请求。
func OperLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if !isOperLogMethod(method) {
			c.Next()
			return
		}

		startedAt := time.Now()
		requestData := ""
		contentType := c.GetHeader("Content-Type")
		if c.Request.Body != nil && !strings.Contains(contentType, "multipart/form-data") {
			peeked, err := io.ReadAll(io.LimitReader(c.Request.Body, maxOperLogBody))
			if err == nil {
				c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(peeked), c.Request.Body))
				requestData = sanitizeOperLogData(string(peeked))
			}
		}

		c.Next()

		username := c.GetString("username")
		router := c.FullPath()
		if router == "" {
			router = c.Request.URL.Path
		}
		serviceName := c.HandlerName()
		if idx := strings.LastIndex(serviceName, "."); idx >= 0 {
			serviceName = serviceName[idx+1:]
		}
		ip := c.ClientIP()
		statusCode := c.Writer.Status()
		durationMS := time.Since(startedAt).Milliseconds()

		// 日志不阻塞业务响应；goroutine 内自行兜底，写入失败也不影响主请求。
		go func() {
			defer func() {
				if r := recover(); r != nil {
					loggerInit.Logger.Get().Error("oper log goroutine panic", zap.Any("panic", r))
				}
			}()
			systemService.RecordOperLog(username, method, router, serviceName, ip, requestData, statusCode, durationMS)
		}()
	}
}

func isOperLogMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func sanitizeOperLogData(data string) string {
	return sensitiveFieldPattern.ReplaceAllString(data, `"$1":"***"`)
}
