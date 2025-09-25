package handlers

import (
	"net/http"
	"strconv"

	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// AuditHandler 审计处理器
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler 创建审计处理器
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// List 获取审计日志列表
func (h *AuditHandler) List(c *gin.Context) {
	logs, err := h.auditService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get audit logs",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    logs,
	})
}

// GetByID 根据ID获取审计日志
func (h *AuditHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid ID",
			"data":    nil,
		})
		return
	}

	log, err := h.auditService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Audit log not found",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    log,
	})
}
