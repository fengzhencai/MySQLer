package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole 用户角色类型
type UserRole string

const (
	RoleAdmin    UserRole = "admin"    // 管理员
	RoleOperator UserRole = "operator" // 操作员
	RoleViewer   UserRole = "viewer"   // 查看者
)

// User 用户模型
type User struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	Username     string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	PasswordHash string         `json:"-" gorm:"type:varchar(255);not null"`
	DisplayName  *string        `json:"display_name" gorm:"type:varchar(100)"`
	Email        *string        `json:"email" gorm:"type:varchar(255)"`
	Role         UserRole       `json:"role" gorm:"type:enum('admin','operator','viewer');default:'viewer'"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 返回表名
func (User) TableName() string {
	return "users"
}

// HasPermission 检查用户是否有指定权限
func (u *User) HasPermission(action string) bool {
	switch u.Role {
	case RoleAdmin:
		return true // 管理员有所有权限
	case RoleOperator:
		// 操作员可以执行DDL但不能管理用户
		return action != "manage_users"
	case RoleViewer:
		// 查看者只能查看
		return action == "view"
	default:
		return false
	}
}

// IsAdmin 检查是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanExecuteDDL 检查是否可以执行DDL
func (u *User) CanExecuteDDL() bool {
	return u.Role == RoleAdmin || u.Role == RoleOperator
}
