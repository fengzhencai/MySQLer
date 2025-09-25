package handlers

import (
	"net/http"

	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService  *services.AuthService
	auditService *services.AuditService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *services.AuthService, auditService *services.AuditService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		auditService: auditService,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"data":    nil,
		})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 记录审计日志
	h.auditService.Log(&services.AuditLogRequest{
		Username:     &resp.User.Username,
		Action:       "login",
		ResourceType: StringPtr("auth"),
		IPAddress:    StringPtr(c.ClientIP()),
		UserAgent:    StringPtr(c.GetHeader("User-Agent")),
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Login successful",
		"data":    resp,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取用户信息
	username, _ := c.Get("username")

	// 记录审计日志
	if username != nil {
		usernameStr := username.(string)
		h.auditService.Log(&services.AuditLogRequest{
			Username:     &usernameStr,
			Action:       "logout",
			ResourceType: StringPtr("auth"),
			IPAddress:    StringPtr(c.ClientIP()),
			UserAgent:    StringPtr(c.GetHeader("User-Agent")),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Logout successful",
		"data":    nil,
	})
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	// TODO: 实现获取用户详细信息的逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"user_id": userID,
		},
	})
}

// StringPtr 字符串指针辅助函数
func StringPtr(s string) *string {
	return &s
}
