package models

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Permission 权限模型
type Permission struct {
	ID          string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description string         `json:"description" gorm:"type:varchar(255)"`
	Resource    string         `json:"resource" gorm:"type:varchar(50);not null"` // 资源类型：connection, execution, user等
	Action      string         `json:"action" gorm:"type:varchar(50);not null"`   // 操作类型：create, read, update, delete, execute等
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 返回表名
func (Permission) TableName() string {
	return "permissions"
}

// RolePermission 角色权限关联表
type RolePermission struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	Role         UserRole       `json:"role" gorm:"type:enum('admin','operator','viewer');not null"`
	PermissionID string         `json:"permission_id" gorm:"type:varchar(36);not null"`
	Permission   Permission     `json:"permission" gorm:"foreignKey:PermissionID"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 返回表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserPermission 用户特殊权限（覆盖角色权限）
type UserPermission struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID       string         `json:"user_id" gorm:"type:varchar(36);not null"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
	PermissionID string         `json:"permission_id" gorm:"type:varchar(36);not null"`
	Permission   Permission     `json:"permission" gorm:"foreignKey:PermissionID"`
	IsGranted    bool           `json:"is_granted" gorm:"default:true"` // true=授予，false=拒绝
	ExpiresAt    *time.Time     `json:"expires_at"`                     // 权限过期时间
	CreatedBy    string         `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 返回表名
func (UserPermission) TableName() string {
	return "user_permissions"
}

// IsExpired 检查权限是否已过期
func (up *UserPermission) IsExpired() bool {
	return up.ExpiresAt != nil && up.ExpiresAt.Before(time.Now())
}

// PermissionKey 生成权限键
func PermissionKey(resource, action string) string {
	return fmt.Sprintf("%s:%s", resource, action)
}

// ParsePermissionKey 解析权限键
func ParsePermissionKey(key string) (resource, action string) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, ""
}

// 预定义权限常量
const (
	// 连接管理权限
	PermissionConnectionView   = "connection:view"
	PermissionConnectionCreate = "connection:create"
	PermissionConnectionUpdate = "connection:update"
	PermissionConnectionDelete = "connection:delete"
	PermissionConnectionTest   = "connection:test"

	// DDL执行权限
	PermissionExecutionView    = "execution:view"
	PermissionExecutionCreate  = "execution:create"
	PermissionExecutionExecute = "execution:execute"
	PermissionExecutionStop    = "execution:stop"
	PermissionExecutionCancel  = "execution:cancel"
	PermissionExecutionDelete  = "execution:delete"
	PermissionExecutionRerun   = "execution:rerun"

	// 用户管理权限
	PermissionUserView   = "user:view"
	PermissionUserCreate = "user:create"
	PermissionUserUpdate = "user:update"
	PermissionUserDelete = "user:delete"

	// 系统管理权限
	PermissionSystemConfig = "system:config"
	PermissionSystemLogs   = "system:logs"
	PermissionSystemStats  = "system:stats"

	// 审计权限
	PermissionAuditView = "audit:view"

	// 危险操作权限
	PermissionDangerousOperations = "dangerous:execute"
)

// GetDefaultPermissions 获取默认权限配置
func GetDefaultPermissions() map[UserRole][]string {
	return map[UserRole][]string{
		RoleAdmin: {
			// 管理员拥有所有权限
			PermissionConnectionView, PermissionConnectionCreate, PermissionConnectionUpdate, PermissionConnectionDelete, PermissionConnectionTest,
			PermissionExecutionView, PermissionExecutionCreate, PermissionExecutionExecute, PermissionExecutionStop, PermissionExecutionCancel, PermissionExecutionDelete, PermissionExecutionRerun,
			PermissionUserView, PermissionUserCreate, PermissionUserUpdate, PermissionUserDelete,
			PermissionSystemConfig, PermissionSystemLogs, PermissionSystemStats,
			PermissionAuditView,
			PermissionDangerousOperations,
		},
		RoleOperator: {
			// 操作员可以管理连接和执行DDL
			PermissionConnectionView, PermissionConnectionCreate, PermissionConnectionUpdate, PermissionConnectionTest,
			PermissionExecutionView, PermissionExecutionCreate, PermissionExecutionExecute, PermissionExecutionStop, PermissionExecutionCancel, PermissionExecutionRerun,
			PermissionSystemStats,
			PermissionAuditView,
		},
		RoleViewer: {
			// 查看者只能查看
			PermissionConnectionView,
			PermissionExecutionView,
			PermissionSystemStats,
		},
	}
}

// GetAllPermissions 获取所有权限定义
func GetAllPermissions() []Permission {
	return []Permission{
		// 连接管理权限
		{Name: PermissionConnectionView, Description: "查看连接配置", Resource: "connection", Action: "view"},
		{Name: PermissionConnectionCreate, Description: "创建连接配置", Resource: "connection", Action: "create"},
		{Name: PermissionConnectionUpdate, Description: "修改连接配置", Resource: "connection", Action: "update"},
		{Name: PermissionConnectionDelete, Description: "删除连接配置", Resource: "connection", Action: "delete"},
		{Name: PermissionConnectionTest, Description: "测试连接", Resource: "connection", Action: "test"},

		// DDL执行权限
		{Name: PermissionExecutionView, Description: "查看执行记录", Resource: "execution", Action: "view"},
		{Name: PermissionExecutionCreate, Description: "创建执行任务", Resource: "execution", Action: "create"},
		{Name: PermissionExecutionExecute, Description: "执行DDL任务", Resource: "execution", Action: "execute"},
		{Name: PermissionExecutionStop, Description: "停止执行任务", Resource: "execution", Action: "stop"},
		{Name: PermissionExecutionCancel, Description: "取消执行任务", Resource: "execution", Action: "cancel"},
		{Name: PermissionExecutionDelete, Description: "删除执行记录", Resource: "execution", Action: "delete"},
		{Name: PermissionExecutionRerun, Description: "重新执行任务", Resource: "execution", Action: "rerun"},

		// 用户管理权限
		{Name: PermissionUserView, Description: "查看用户信息", Resource: "user", Action: "view"},
		{Name: PermissionUserCreate, Description: "创建用户", Resource: "user", Action: "create"},
		{Name: PermissionUserUpdate, Description: "修改用户信息", Resource: "user", Action: "update"},
		{Name: PermissionUserDelete, Description: "删除用户", Resource: "user", Action: "delete"},

		// 系统管理权限
		{Name: PermissionSystemConfig, Description: "系统配置管理", Resource: "system", Action: "config"},
		{Name: PermissionSystemLogs, Description: "查看系统日志", Resource: "system", Action: "logs"},
		{Name: PermissionSystemStats, Description: "查看系统统计", Resource: "system", Action: "stats"},

		// 审计权限
		{Name: PermissionAuditView, Description: "查看审计日志", Resource: "audit", Action: "view"},

		// 危险操作权限
		{Name: PermissionDangerousOperations, Description: "执行危险操作", Resource: "dangerous", Action: "execute"},
	}
}
