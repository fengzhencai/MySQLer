package middleware

import (
	"net/http"
	"strings"

	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(permissionService *services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过登录等不需要权限的接口
		if shouldSkipPermissionCheck(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未登录",
			})
			c.Abort()
			return
		}

		// 根据请求路径和方法确定所需权限
		requiredPermission := getRequiredPermission(c.Request.Method, c.Request.URL.Path)
		if requiredPermission == "" {
			// 如果没有明确的权限要求，允许通过
			c.Next()
			return
		}

		// 检查权限
		hasPermission, err := permissionService.HasPermission(userID.(string), requiredPermission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限检查失败",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":                403,
				"message":             "权限不足",
				"required_permission": requiredPermission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 要求特定权限的中间件
func RequirePermission(permissionService *services.PermissionService, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未登录",
			})
			c.Abort()
			return
		}

		hasPermission, err := permissionService.HasPermission(userID.(string), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限检查失败",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":                403,
				"message":             "权限不足",
				"required_permission": permission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}


// RequireExecutionPermission DDL执行权限检查中间件
func RequireExecutionPermission(permissionService *services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未登录",
			})
			c.Abort()
			return
		}

		// 从请求中获取执行参数
		var requestBody struct {
			ConnectionID string         `json:"connection_id"`
			TableName    string         `json:"table_name"`
			DDLType      models.DDLType `json:"ddl_type"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 检查执行权限
		err := permissionService.CheckExecutionPermission(
			userID.(string),
			requestBody.ConnectionID,
			requestBody.TableName,
			requestBody.DDLType,
		)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "执行权限不足",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// shouldSkipPermissionCheck 判断是否跳过权限检查
func shouldSkipPermissionCheck(path string) bool {
	skipPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/api/v1/health",
		"/api/v1/ping",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

// getRequiredPermission 根据请求路径和方法获取所需权限
func getRequiredPermission(method, path string) string {
	// 规范化路径，移除查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// 权限映射表
	permissionMap := map[string]map[string]string{
		// 连接管理
		"GET /api/v1/connections":         {"GET": models.PermissionConnectionView},
		"POST /api/v1/connections":        {"POST": models.PermissionConnectionCreate},
		"PUT /api/v1/connections/*":       {"PUT": models.PermissionConnectionUpdate},
		"DELETE /api/v1/connections/*":    {"DELETE": models.PermissionConnectionDelete},
		"POST /api/v1/connections/*/test": {"POST": models.PermissionConnectionTest},

		// DDL执行
		"GET /api/v1/executions":               {"GET": models.PermissionExecutionView},
		"POST /api/v1/executions":              {"POST": models.PermissionExecutionCreate},
		"POST /api/v1/executions/*/start":      {"POST": models.PermissionExecutionExecute},
		"POST /api/v1/executions/*/stop":       {"POST": models.PermissionExecutionStop},
		"POST /api/v1/executions/*/cancel":     {"POST": models.PermissionExecutionCancel},
		"DELETE /api/v1/executions/*":          {"DELETE": models.PermissionExecutionDelete},
		"POST /api/v1/executions/*/re-execute": {"POST": models.PermissionExecutionRerun},

		// 用户管理
		"GET /api/v1/users":      {"GET": models.PermissionUserView},
		"POST /api/v1/users":     {"POST": models.PermissionUserCreate},
		"PUT /api/v1/users/*":    {"PUT": models.PermissionUserUpdate},
		"DELETE /api/v1/users/*": {"DELETE": models.PermissionUserDelete},

		// 系统管理
		"GET /api/v1/system/config": {"GET": models.PermissionSystemConfig},
		"PUT /api/v1/system/config": {"PUT": models.PermissionSystemConfig},
		"GET /api/v1/system/logs":   {"GET": models.PermissionSystemLogs},
		"GET /api/v1/system/stats":  {"GET": models.PermissionSystemStats},

		// 审计日志
		"GET /api/v1/audit": {"GET": models.PermissionAuditView},
	}

	// 尝试精确匹配
	key := method + " " + path
	if methodMap, exists := permissionMap[key]; exists {
		if permission, exists := methodMap[method]; exists {
			return permission
		}
	}

	// 尝试通配符匹配
	for pattern, methodMap := range permissionMap {
		if matchPattern(key, pattern) {
			if permission, exists := methodMap[method]; exists {
				return permission
			}
		}
	}

	return ""
}

// matchPattern 匹配路径模式（支持*通配符）
func matchPattern(path, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return path == pattern
	}

	// 简单的通配符匹配
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return false
	}

	prefix := parts[0]
	suffix := parts[1]

	return strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) && len(path) >= len(prefix)+len(suffix)
}
