package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AuditAction 审计动作类型
type AuditAction string

const (
	// 用户认证相关
	ActionLogin  AuditAction = "login"
	ActionLogout AuditAction = "logout"

	// 连接管理相关
	ActionConnectionCreate AuditAction = "connection_create"
	ActionConnectionUpdate AuditAction = "connection_update"
	ActionConnectionDelete AuditAction = "connection_delete"
	ActionConnectionTest   AuditAction = "connection_test"

	// DDL执行相关
	ActionExecutionCreate AuditAction = "execution_create"
	ActionExecutionStart  AuditAction = "execution_start"
	ActionExecutionStop   AuditAction = "execution_stop"
	ActionExecutionCancel AuditAction = "execution_cancel"
	ActionExecutionDelete AuditAction = "execution_delete"
	ActionExecutionRerun  AuditAction = "execution_rerun"

	// 用户管理相关
	ActionUserCreate AuditAction = "user_create"
	ActionUserUpdate AuditAction = "user_update"
	ActionUserDelete AuditAction = "user_delete"

	// 权限管理相关
	ActionPermissionGrant  AuditAction = "permission_grant"
	ActionPermissionRevoke AuditAction = "permission_revoke"

	// 系统管理相关
	ActionSystemConfig AuditAction = "system_config"
	ActionSystemBackup AuditAction = "system_backup"
)

// AuditStatus 审计状态
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailed  AuditStatus = "failed"
	AuditStatusBlocked AuditStatus = "blocked" // 被权限系统阻止
)

// AuditLog 操作审计日志模型
type AuditLog struct {
	ID              int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID          *string          `json:"user_id" gorm:"type:varchar(36);index"`
	Username        *string          `json:"username" gorm:"type:varchar(50)"`
	Action          string           `json:"action" gorm:"type:varchar(100);not null;index"`
	ResourceType    *string          `json:"resource_type" gorm:"type:varchar(50)"`
	ResourceID      *string          `json:"resource_id" gorm:"type:varchar(36)"`
	RequestData     *json.RawMessage `json:"request_data" gorm:"type:json"`
	ResponseData    *json.RawMessage `json:"response_data" gorm:"type:json"`
	IPAddress       *string          `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent       *string          `json:"user_agent" gorm:"type:text"`
	ExecutionTimeMs *int             `json:"execution_time_ms"`
	Status          AuditStatus      `json:"status" gorm:"type:varchar(20);default:'success';index"`
	ErrorMsg        *string          `json:"error_msg" gorm:"type:text"`
	RiskLevel       string           `json:"risk_level" gorm:"type:varchar(20);default:'low';index"` // low, medium, high
	Environment     *string          `json:"environment" gorm:"type:varchar(20);index"`              // prod, test, dev
	SessionID       *string          `json:"session_id" gorm:"type:varchar(36);index"`               // 会话ID
	RequestID       *string          `json:"request_id" gorm:"type:varchar(36);index"`               // 请求ID
	CreatedAt       time.Time        `json:"created_at" gorm:"index"`
	DeletedAt       gorm.DeletedAt   `json:"-" gorm:"index"`
}

// TableName 返回表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogRequest 审计日志请求
type AuditLogRequest struct {
	UserID       *string          `json:"user_id"`
	Username     *string          `json:"username"`
	Action       string           `json:"action"`
	ResourceType *string          `json:"resource_type"`
	ResourceID   *string          `json:"resource_id"`
	RequestData  *json.RawMessage `json:"request_data"`
	ResponseData *json.RawMessage `json:"response_data"`
	IPAddress    *string          `json:"ip_address"`
	UserAgent    *string          `json:"user_agent"`
}

// ToAuditLog 转换为审计日志模型
func (r *AuditLogRequest) ToAuditLog() *AuditLog {
	return &AuditLog{
		UserID:       r.UserID,
		Username:     r.Username,
		Action:       r.Action,
		ResourceType: r.ResourceType,
		ResourceID:   r.ResourceID,
		RequestData:  r.RequestData,
		ResponseData: r.ResponseData,
		IPAddress:    r.IPAddress,
		UserAgent:    r.UserAgent,
		Status:       AuditStatusSuccess,
		RiskLevel:    "low",
		CreatedAt:    time.Now(),
	}
}

// GetRiskLevel 根据操作类型和环境自动确定风险等级
func GetRiskLevel(action string, environment string, resource string) string {
	// 生产环境操作风险较高
	if environment == "prod" {
		switch action {
		case string(ActionExecutionStart), string(ActionConnectionDelete), string(ActionUserDelete):
			return "high"
		case string(ActionConnectionCreate), string(ActionConnectionUpdate), string(ActionExecutionCreate):
			return "medium"
		default:
			return "low"
		}
	}

	// 危险操作
	dangerousActions := []string{
		string(ActionExecutionStart),
		string(ActionConnectionDelete),
		string(ActionUserDelete),
		string(ActionSystemConfig),
	}

	for _, dangerousAction := range dangerousActions {
		if action == dangerousAction {
			return "medium"
		}
	}

	return "low"
}

// IsHighRisk 判断是否为高风险操作
func (a *AuditLog) IsHighRisk() bool {
	return a.RiskLevel == "high"
}

// IsFailed 判断操作是否失败
func (a *AuditLog) IsFailed() bool {
	return a.Status == AuditStatusFailed || a.Status == AuditStatusBlocked
}

// SetError 设置错误信息
func (a *AuditLog) SetError(err error) {
	if err != nil {
		a.Status = AuditStatusFailed
		errMsg := err.Error()
		a.ErrorMsg = &errMsg
	}
}

// SetBlocked 设置被阻止状态
func (a *AuditLog) SetBlocked(reason string) {
	a.Status = AuditStatusBlocked
	a.ErrorMsg = &reason
}
