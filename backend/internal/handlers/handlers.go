package handlers

import (
	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/middleware"
	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(router *gin.Engine, services *services.Services, cfg *config.Config) {
	// API版本分组
	v1 := router.Group("/api/v1")

	// 认证相关路由（无需认证）
	authHandler := NewAuthHandler(services.Auth, services.Audit)
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// MVP最小接口（无鉴权）
	mvp := v1.Group("/mvp")
	{
		mvph := NewMVPHandler(services.MVP)
		mvp.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"pong": true}) })
		mvp.POST("/preview", mvph.Preview)
		mvp.POST("/execute", mvph.Execute)
	}

	// 需要认证的路由
	authenticated := v1.Group("")
	authenticated.Use(middleware.AuthMiddleware(cfg))
	{
		// 用户相关
		authenticated.GET("/auth/profile", authHandler.GetProfile)

		// 连接管理
		connectionHandler := NewConnectionHandler(services.Connection, services.Audit)
		connectionGroup := authenticated.Group("/connections")
		{
			connectionGroup.GET("", connectionHandler.List)
			connectionGroup.POST("", connectionHandler.Create)
			connectionGroup.GET("/:id", connectionHandler.GetByID)
			connectionGroup.PUT("/:id", connectionHandler.Update)
			connectionGroup.DELETE("/:id", connectionHandler.Delete)
			connectionGroup.POST("/:id/test", connectionHandler.Test)
			// 基于参数的测试连接（不依赖已保存的连接）
			connectionGroup.POST("/test", connectionHandler.TestByParams)
		}

		// DDL执行
		executionHandler := NewExecutionHandler(services.Execution, services.ExecutionEngine, services.Audit)
		executionGroup := authenticated.Group("/executions")
		{
			executionGroup.GET("", executionHandler.List)
			executionGroup.POST("", executionHandler.Create)
			executionGroup.GET("/:id", executionHandler.GetByID)
			executionGroup.POST("/:id/stop", executionHandler.Stop)
			executionGroup.POST("/:id/retry", executionHandler.Retry)
			executionGroup.GET("/:id/logs", executionHandler.GetLogs)
			executionGroup.POST("/preview", executionHandler.PreviewCommand)
			executionGroup.POST("/:id/start", executionHandler.StartExecution)
			executionGroup.GET("/:id/status", executionHandler.GetExecutionStatus)
			executionGroup.GET("/running", executionHandler.GetRunningTasks)
		}

		// 工具类接口
		toolsGroup := authenticated.Group("/tools")
		{
			toolsGroup.GET("/connections/:id/databases", connectionHandler.GetDatabases)
			toolsGroup.GET("/connections/:id/databases/:database/tables", connectionHandler.GetTables)
			toolsGroup.GET("/connections/:id/databases/:database/tables/:table/schema", connectionHandler.GetTableSchema)
		}

		// 管理员路由
		adminGroup := authenticated.Group("/admin")
		adminGroup.Use(middleware.RequireRole("admin"))
		{
			// 用户管理
			userHandler := NewUserHandler(services.User, services.Audit)
			userGroup := adminGroup.Group("/users")
			{
				userGroup.GET("", userHandler.List)
				userGroup.POST("", userHandler.Create)
				userGroup.GET("/:id", userHandler.GetByID)
				userGroup.PUT("/:id", userHandler.Update)
				userGroup.DELETE("/:id", userHandler.Delete)
			}

			// 审计日志
			auditHandler := NewAuditHandler(services.Audit)
			auditGroup := adminGroup.Group("/audit-logs")
			{
				auditGroup.GET("", auditHandler.List)
				auditGroup.GET("/:id", auditHandler.GetByID)
			}
		}
	}

	// WebSocket路由（独立鉴权：支持query参数token）
	wsHandler := NewWebSocketHandler(services.ExecutionEngine)
	wsGroup := v1.Group("/ws")
	wsGroup.Use(middleware.WebSocketAuthMiddleware(cfg))
	{
		wsGroup.GET("/execution", wsHandler.HandleWebSocket)
	}

	// 设置WebSocket广播器
	services.ExecutionEngine.SetBroadcasters(
		wsHandler.BroadcastExecutionLog,
		func(executionID string, progress interface{}) {
			if progressMsg, ok := progress.(map[string]interface{}); ok {
				// 创建进度消息并广播
				progressData := map[string]interface{}{
					"execution_id":  progressMsg["execution_id"],
					"status":        progressMsg["status"],
					"progress":      progressMsg["progress"],
					"current_speed": progressMsg["current_speed"],
					"current_stage": progressMsg["current_stage"],
					"timestamp":     progressMsg["timestamp"],
				}
				wsHandler.BroadcastExecutionProgress(executionID, progressData)
			}
		},
	)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
