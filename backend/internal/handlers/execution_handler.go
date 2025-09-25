package handlers

import (
	"net/http"

	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ExecutionHandler 执行处理器
type ExecutionHandler struct {
	executionService *services.ExecutionService
	executionEngine  *services.ExecutionEngine
	auditService     *services.AuditService
}

// NewExecutionHandler 创建执行处理器
func NewExecutionHandler(executionService *services.ExecutionService, executionEngine *services.ExecutionEngine, auditService *services.AuditService) *ExecutionHandler {
	return &ExecutionHandler{
		executionService: executionService,
		executionEngine:  executionEngine,
		auditService:     auditService,
	}
}

// List 获取执行记录列表
func (h *ExecutionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	records, err := h.executionService.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get execution records",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    records,
	})
}

// Create 创建执行记录
func (h *ExecutionHandler) Create(c *gin.Context) {
	var req services.CreateExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"data":    nil,
		})
		return
	}

	userID := c.GetString("user_id")

	record, err := h.executionService.Create(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create execution",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Execution created successfully",
		"data":    record,
	})
}

// GetByID 根据ID获取执行记录
func (h *ExecutionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	record, err := h.executionService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Execution record not found",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    record,
	})
}

// Stop 停止执行
func (h *ExecutionHandler) Stop(c *gin.Context) {
	id := c.Param("id")

	err := h.executionService.Stop(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to stop execution",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Execution stopped successfully",
		"data":    nil,
	})
}

// Retry 重试执行
func (h *ExecutionHandler) Retry(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	err := h.executionService.Retry(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to retry execution",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Execution retried successfully",
		"data":    nil,
	})
}

// GetLogs 获取执行日志
func (h *ExecutionHandler) GetLogs(c *gin.Context) {
	id := c.Param("id")

	logs, err := h.executionService.GetLogs(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get execution logs",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    gin.H{"logs": logs},
	})
}

// PreviewCommand 预览PT命令
func (h *ExecutionHandler) PreviewCommand(c *gin.Context) {
	var req services.PreviewCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"data":    nil,
		})
		return
	}

	response, err := h.executionService.PreviewCommand(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    response,
	})
}

// StartExecution 启动执行
func (h *ExecutionHandler) StartExecution(c *gin.Context) {
	id := c.Param("id")

	// TODO: 实现WebSocket日志推送
	logCallback := func(logLine string) {
		// 这里后续会通过WebSocket推送到前端
		// 暂时可以记录到数据库或缓存
	}

	err := h.executionEngine.StartExecution(id, logCallback)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Execution started successfully",
		"data":    nil,
	})
}

// GetExecutionStatus 获取执行状态
func (h *ExecutionHandler) GetExecutionStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := h.executionEngine.GetTaskStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    status,
	})
}

// GetRunningTasks 获取所有运行中的任务
func (h *ExecutionHandler) GetRunningTasks(c *gin.Context) {
	tasks := h.executionEngine.GetRunningTasks()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    tasks,
	})
}
