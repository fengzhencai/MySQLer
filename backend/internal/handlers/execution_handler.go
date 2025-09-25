package handlers

import (
	"net/http"
	"strconv"

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
	// 读取查询参数
	page := 1
	size := 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			size = n
		}
	}

	params := map[string]interface{}{
		"status":        c.Query("status"),
		"connection_id": c.Query("connection_id"),
		"start_date":    c.Query("start_date"),
		"end_date":      c.Query("end_date"),
		"keyword":       c.Query("keyword"),
		"page":          page,
		"size":          size,
	}

	records, total, err := h.executionService.ListWithFilters(params)
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
		"data": gin.H{
			"records": records,
			"total":   total,
		},
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
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

	// 先尝试通过引擎停止运行中的任务
	if err := h.executionEngine.StopExecution(id); err != nil {
		// 引擎未在运行或停止失败则继续走服务层兜底
	}

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
