package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// responseWriter 包装gin.ResponseWriter以捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditMiddleware 审计日志中间件
func AuditMiddleware(auditService *services.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过不需要审计的接口
		if shouldSkipAudit(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 生成请求ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// 记录开始时间
		startTime := time.Now()

		// 读取请求体
		var requestData json.RawMessage
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// 恢复请求体
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// 过滤敏感数据
				filteredBody := filterSensitiveData(bodyBytes)
				requestData = json.RawMessage(filteredBody)
			}
		}

		// 包装响应写入器以捕获响应数据
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 执行请求
		c.Next()

		// 计算执行时间
		duration := time.Since(startTime)
		executionTimeMs := int(duration.Milliseconds())

		// 创建审计日志
		auditLog := &models.AuditLog{
			RequestID:       &requestID,
			Action:          getActionFromRequest(c.Request.Method, c.Request.URL.Path),
			ResourceType:    getResourceTypeFromPath(c.Request.URL.Path),
			RequestData:     &requestData,
			IPAddress:       getClientIP(c.Request),
			UserAgent:       getUserAgent(c.Request),
			ExecutionTimeMs: &executionTimeMs,
			CreatedAt:       startTime,
		}

		// 设置用户信息
		if userID, exists := c.Get("user_id"); exists {
			userIDStr := userID.(string)
			auditLog.UserID = &userIDStr
		}
		if username, exists := c.Get("username"); exists {
			usernameStr := username.(string)
			auditLog.Username = &usernameStr
		}

		// 设置资源ID（从URL参数中提取）
		if resourceID := extractResourceID(c.Request.URL.Path, c.Params); resourceID != "" {
			auditLog.ResourceID = &resourceID
		}

		// 设置环境信息（从连接信息中获取）
		if env := getEnvironmentFromContext(c); env != "" {
			auditLog.Environment = &env
		}

		// 设置会话ID
		if sessionID, exists := c.Get("session_id"); exists {
			sessionIDStr := sessionID.(string)
			auditLog.SessionID = &sessionIDStr
		}

		// 处理响应数据和状态
		statusCode := c.Writer.Status()
		if statusCode >= 200 && statusCode < 300 {
			auditLog.Status = models.AuditStatusSuccess
		} else if statusCode == 403 {
			auditLog.Status = models.AuditStatusBlocked
		} else {
			auditLog.Status = models.AuditStatusFailed
		}

		// 捕获响应数据（限制大小）
		if blw.body.Len() > 0 && blw.body.Len() < 10240 { // 限制10KB
			responseData := json.RawMessage(blw.body.Bytes())
			auditLog.ResponseData = &responseData
		}

		// 设置风险等级
		auditLog.RiskLevel = models.GetRiskLevel(
			auditLog.Action,
			getStringPtr(auditLog.Environment),
			getStringPtr(auditLog.ResourceType),
		)

		// 处理错误信息
		if errors, exists := c.Get("errors"); exists {
			if errorList, ok := errors.([]error); ok && len(errorList) > 0 {
				errorMsg := errorList[0].Error()
				auditLog.ErrorMsg = &errorMsg
			}
		}

		// 异步保存审计日志
		go func() {
			if err := auditService.CreateAuditLog(auditLog); err != nil {
				// 日志保存失败时的处理（可以写入文件或发送到外部系统）
				// 这里简单打印错误，实际项目中应该有更robust的错误处理
				c.Error(err)
			}
		}()
	}
}

// AuditAction 手动记录审计日志
func AuditAction(auditService *services.AuditService, c *gin.Context, action string, resourceType, resourceID *string) {
	auditLog := &models.AuditLog{
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    getClientIP(c.Request),
		UserAgent:    getUserAgent(c.Request),
		Status:       models.AuditStatusSuccess,
		RiskLevel:    "low",
		CreatedAt:    time.Now(),
	}

	// 设置用户信息
	if userID, exists := c.Get("user_id"); exists {
		userIDStr := userID.(string)
		auditLog.UserID = &userIDStr
	}
	if username, exists := c.Get("username"); exists {
		usernameStr := username.(string)
		auditLog.Username = &usernameStr
	}

	// 设置请求ID
	if requestID, exists := c.Get("request_id"); exists {
		requestIDStr := requestID.(string)
		auditLog.RequestID = &requestIDStr
	}

	// 异步保存
	go auditService.CreateAuditLog(auditLog)
}

// 辅助函数

func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/api/v1/health",
		"/api/v1/ping",
		"/metrics",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	// 跳过静态资源
	if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/assets/") {
		return true
	}

	return false
}

func getActionFromRequest(method, path string) string {
	// 简化的动作映射
	actionMap := map[string]string{
		"POST /api/v1/auth/login":    string(models.ActionLogin),
		"POST /api/v1/auth/logout":   string(models.ActionLogout),
		"POST /api/v1/connections":   string(models.ActionConnectionCreate),
		"PUT /api/v1/connections":    string(models.ActionConnectionUpdate),
		"DELETE /api/v1/connections": string(models.ActionConnectionDelete),
		"POST /api/v1/executions":    string(models.ActionExecutionCreate),
		"POST /api/v1/users":         string(models.ActionUserCreate),
		"PUT /api/v1/users":          string(models.ActionUserUpdate),
		"DELETE /api/v1/users":       string(models.ActionUserDelete),
	}

	key := method + " " + path
	if action, exists := actionMap[key]; exists {
		return action
	}

	// 通配符匹配
	for pattern, action := range actionMap {
		if matchPathPattern(key, pattern) {
			return action
		}
	}

	// 默认动作
	switch method {
	case "GET":
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

func getResourceTypeFromPath(path string) *string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		resourceType := parts[2]
		return &resourceType
	}
	return nil
}

func extractResourceID(path string, params gin.Params) string {
	// 从gin参数中提取ID
	for _, param := range params {
		if param.Key == "id" || strings.HasSuffix(param.Key, "_id") {
			return param.Value
		}
	}

	// 从路径中提取UUID模式的ID
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if isUUID(part) {
			return part
		}
	}

	return ""
}

func getEnvironmentFromContext(c *gin.Context) string {
	// 尝试从上下文中获取环境信息
	if env, exists := c.Get("environment"); exists {
		return env.(string)
	}

	// 从连接信息中获取环境
	if connectionID, exists := c.Get("connection_id"); exists {
		// 这里应该查询数据库获取连接的环境信息
		// 简化处理，返回空字符串
		_ = connectionID
	}

	return ""
}

func getClientIP(r *http.Request) *string {
	// 优先获取真实IP
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// X-Forwarded-For 可能包含多个IP，取第一个
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		}
	}
	if ip == "" {
		ip = r.RemoteAddr
		// 去掉端口号
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
	}

	return &ip
}

func getUserAgent(r *http.Request) *string {
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return nil
	}
	return &userAgent
}

func filterSensitiveData(data []byte) []byte {
	// 过滤敏感数据（如密码、token等）
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return data // 无法解析则返回原数据
	}

	sensitiveFields := []string{
		"password", "passwd", "pwd",
		"token", "access_token", "refresh_token",
		"secret", "key", "private_key",
		"authorization",
	}

	for _, field := range sensitiveFields {
		if _, exists := obj[field]; exists {
			obj[field] = "***"
		}
	}

	filtered, err := json.Marshal(obj)
	if err != nil {
		return data
	}
	return filtered
}

func matchPathPattern(path, pattern string) bool {
	// 简单的路径模式匹配（支持通配符）
	return strings.Contains(pattern, "*") &&
		strings.HasPrefix(path, strings.Split(pattern, "*")[0])
}

func isUUID(s string) bool {
	// 简单的UUID格式检查
	return len(s) == 36 && strings.Count(s, "-") == 4
}

func getStringPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
