package handlers

import (
	"net/http"

	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ConnectionHandler 连接处理器
type ConnectionHandler struct {
	connectionService *services.ConnectionService
	auditService      *services.AuditService
}

// NewConnectionHandler 创建连接处理器
func NewConnectionHandler(connectionService *services.ConnectionService, auditService *services.AuditService) *ConnectionHandler {
	return &ConnectionHandler{
		connectionService: connectionService,
		auditService:      auditService,
	}
}

// List 获取连接列表
func (h *ConnectionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	connections, err := h.connectionService.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get connections",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    connections,
	})
}

// Create 创建连接
func (h *ConnectionHandler) Create(c *gin.Context) {
	var req services.CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"data":    nil,
		})
		return
	}

	userID := c.GetString("user_id")

	connection, err := h.connectionService.Create(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create connection",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Connection created successfully",
		"data":    connection,
	})
}

// GetByID 根据ID获取连接
func (h *ConnectionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	connection, err := h.connectionService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Connection not found",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    connection,
	})
}

// Update 更新连接
func (h *ConnectionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req services.CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"data":    nil,
		})
		return
	}

	userID := c.GetString("user_id")

	connection, err := h.connectionService.Update(id, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update connection",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Connection updated successfully",
		"data":    connection,
	})
}

// Delete 删除连接
func (h *ConnectionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.connectionService.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to delete connection",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Connection deleted successfully",
		"data":    nil,
	})
}

// Test 测试连接
func (h *ConnectionHandler) Test(c *gin.Context) {
	id := c.Param("id")

	result, err := h.connectionService.TestConnection(id)
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
		"data":    result,
	})
}

// TestByParams 基于传入参数测试连接（创建/编辑表单使用）
func (h *ConnectionHandler) TestByParams(c *gin.Context) {
	var req services.CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request parameters",
			"data":    nil,
		})
		return
	}

	result, err := h.connectionService.TestConnectionByParams(&req)
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
		"data":    result,
	})
}

// GetDatabases 获取连接的数据库列表
func (h *ConnectionHandler) GetDatabases(c *gin.Context) {
	id := c.Param("id")

	databases, err := h.connectionService.GetDatabases(id)
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
		"data":    databases,
	})
}

// GetTables 获取指定数据库的表列表
func (h *ConnectionHandler) GetTables(c *gin.Context) {
	id := c.Param("id")
	database := c.Param("database")

	tables, err := h.connectionService.GetTables(id, database)
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
		"data":    tables,
	})
}

// GetTableSchema 获取指定表的列与索引结构
func (h *ConnectionHandler) GetTableSchema(c *gin.Context) {
	id := c.Param("id")
	database := c.Param("database")
	table := c.Param("table")

	schema, err := h.connectionService.GetTableSchema(id, database, table)
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
		"data":    schema,
	})
}
