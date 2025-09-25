package handlers

import (
	"net/http"

	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type MVPHandler struct {
	svc *services.MVPService
}

func NewMVPHandler(svc *services.MVPService) *MVPHandler { return &MVPHandler{svc: svc} }

// Preview 生成命令（不执行）
func (h *MVPHandler) Preview(c *gin.Context) {
	var req services.MVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request"})
		return
	}

	preview, _, _, err := h.svc.BuildCommands(&req, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"generated_command": preview}})
}

// Execute 容器内执行
func (h *MVPHandler) Execute(c *gin.Context) {
	var req services.MVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request"})
		return
	}

	resp, err := h.svc.Execute(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{
			"generated_command": resp.GeneratedCommand,
			"exit_code":         resp.ExitCode,
			"stdout":            resp.Stdout,
			"stderr":            resp.Stderr,
		}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}
