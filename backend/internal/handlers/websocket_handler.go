package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fengzhencai/MySQLer/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	upgrader        websocket.Upgrader
	executionEngine *services.ExecutionEngine

	// 连接管理
	connections map[string]*WebSocketConnection
	mutex       sync.RWMutex
}

// WebSocketConnection WebSocket连接
type WebSocketConnection struct {
	ID          string
	Conn        *websocket.Conn
	UserID      string
	ExecutionID string
	SendCh      chan []byte
	CloseCh     chan bool
}

// WebSocketMessage WebSocket消息
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// ExecutionProgressMessage 执行进度消息
type ExecutionProgressMessage struct {
	ExecutionID   string  `json:"execution_id"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`
	ProcessedRows int64   `json:"processed_rows"`
	TotalRows     int64   `json:"total_rows"`
	CurrentSpeed  float64 `json:"current_speed"`
	CurrentStage  string  `json:"current_stage"`
	LogLine       string  `json:"log_line,omitempty"`
	Timestamp     string  `json:"timestamp"`
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(executionEngine *services.ExecutionEngine) *WebSocketHandler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 生产环境中应该检查Origin
			return true
		},
	}

	return &WebSocketHandler{
		upgrader:        upgrader,
		executionEngine: executionEngine,
		connections:     make(map[string]*WebSocketConnection),
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 获取用户信息
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取执行ID（可选）
	executionID := c.Query("execution_id")

	// 升级HTTP连接为WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// 创建WebSocket连接
	wsConn := &WebSocketConnection{
		ID:          fmt.Sprintf("%s-%d", userID, len(h.connections)),
		Conn:        conn,
		UserID:      userID,
		ExecutionID: executionID,
		SendCh:      make(chan []byte, 256),
		CloseCh:     make(chan bool),
	}

	// 注册连接
	h.registerConnection(wsConn)

	// 启动读写协程
	go h.handleConnection(wsConn)
}

// registerConnection 注册WebSocket连接
func (h *WebSocketHandler) registerConnection(conn *WebSocketConnection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.connections[conn.ID] = conn
	log.Printf("WebSocket连接已注册: %s (用户: %s)", conn.ID, conn.UserID)
}

// unregisterConnection 注销WebSocket连接
func (h *WebSocketHandler) unregisterConnection(connID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if conn, exists := h.connections[connID]; exists {
		close(conn.SendCh)
		delete(h.connections, connID)
		log.Printf("WebSocket连接已注销: %s", connID)
	}
}

// handleConnection 处理WebSocket连接
func (h *WebSocketHandler) handleConnection(wsConn *WebSocketConnection) {
	defer func() {
		wsConn.Conn.Close()
		h.unregisterConnection(wsConn.ID)
	}()

	// 启动写协程
	go h.writeMessages(wsConn)

	// 发送连接成功消息
	h.sendMessage(wsConn, &WebSocketMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"connection_id": wsConn.ID,
			"message":       "WebSocket连接已建立",
		},
	})

	// 如果指定了执行ID，发送当前状态
	if wsConn.ExecutionID != "" {
		h.sendExecutionStatus(wsConn, wsConn.ExecutionID)
	}

	// 读取客户端消息
	for {
		var msg WebSocketMessage
		err := wsConn.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 处理客户端消息
		h.handleClientMessage(wsConn, &msg)
	}
}

// writeMessages 写消息协程
func (h *WebSocketHandler) writeMessages(wsConn *WebSocketConnection) {
	ticker := &time.Ticker{}
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-wsConn.SendCh:
			if !ok {
				wsConn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			wsConn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := wsConn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-wsConn.CloseCh:
			return
		}
	}
}

// handleClientMessage 处理客户端消息
func (h *WebSocketHandler) handleClientMessage(wsConn *WebSocketConnection, msg *WebSocketMessage) {
	switch msg.Type {
	case "subscribe_execution":
		// 订阅执行进度
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if executionID, ok := data["execution_id"].(string); ok {
				wsConn.ExecutionID = executionID
				h.sendExecutionStatus(wsConn, executionID)
			}
		}

	case "unsubscribe_execution":
		// 取消订阅
		wsConn.ExecutionID = ""

	case "ping":
		// 心跳检测
		h.sendMessage(wsConn, &WebSocketMessage{
			Type: "pong",
			Data: map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
		})

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// sendMessage 发送消息
func (h *WebSocketHandler) sendMessage(wsConn *WebSocketConnection, msg *WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Message marshal error: %v", err)
		return
	}

	select {
	case wsConn.SendCh <- data:
	default:
		// 通道已满，关闭连接
		close(wsConn.CloseCh)
	}
}

// sendExecutionStatus 发送执行状态
func (h *WebSocketHandler) sendExecutionStatus(wsConn *WebSocketConnection, executionID string) {
	// 获取执行状态
	status, err := h.executionEngine.GetTaskStatus(executionID)
	if err != nil {
		h.sendMessage(wsConn, &WebSocketMessage{
			Type: "error",
			Data: map[string]interface{}{
				"message": fmt.Sprintf("获取执行状态失败: %v", err),
			},
		})
		return
	}

	// 发送状态消息
	h.sendMessage(wsConn, &WebSocketMessage{
		Type: "execution_progress",
		Data: &ExecutionProgressMessage{
			ExecutionID:  executionID,
			Status:       string(status.Status),
			Progress:     status.Progress,
			CurrentSpeed: status.Speed,
			CurrentStage: status.CurrentStage,
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// BroadcastExecutionProgress 广播执行进度（给所有相关连接）
func (h *WebSocketHandler) BroadcastExecutionProgress(executionID string, progress interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	message := &WebSocketMessage{
		Type: "execution_progress",
		Data: progress,
	}

	for _, conn := range h.connections {
		if conn.ExecutionID == executionID {
			h.sendMessage(conn, message)
		}
	}
}

// BroadcastExecutionLog 广播执行日志
func (h *WebSocketHandler) BroadcastExecutionLog(executionID string, logLine string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	message := &WebSocketMessage{
		Type: "execution_log",
		Data: map[string]interface{}{
			"execution_id": executionID,
			"log_line":     logLine,
			"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	for _, conn := range h.connections {
		if conn.ExecutionID == executionID {
			h.sendMessage(conn, message)
		}
	}
}

// GetActiveConnections 获取活跃连接数
func (h *WebSocketHandler) GetActiveConnections() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.connections)
}

// CloseAllConnections 关闭所有连接
func (h *WebSocketHandler) CloseAllConnections() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, conn := range h.connections {
		close(conn.CloseCh)
		conn.Conn.Close()
	}
	h.connections = make(map[string]*WebSocketConnection)
}
