package models

import (
	"time"

	"gorm.io/gorm"
)

// Environment 环境类型
type Environment string

const (
	EnvProduction  Environment = "prod" // 生产环境
	EnvTest        Environment = "test" // 测试环境
	EnvDevelopment Environment = "dev"  // 开发环境
)

// Connection 数据库连接模型
type Connection struct {
	ID             string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name           string         `json:"name" gorm:"type:varchar(100);not null;index"`
	Environment    Environment    `json:"environment" gorm:"type:enum('prod','test','dev');default:'test';index"`
	Host           string         `json:"host" gorm:"type:varchar(255);not null"`
	Port           int            `json:"port" gorm:"default:3306"`
	Username       string         `json:"username" gorm:"type:varchar(100);not null"`
	Password       string         `json:"-" gorm:"type:text;not null"` // 加密存储，不在JSON中返回
	DatabaseName   string         `json:"database_name" gorm:"type:varchar(100);not null"`
	Description    *string        `json:"description" gorm:"type:text"`
	ConnectTimeout int            `json:"connect_timeout" gorm:"default:5"`
	Charset        string         `json:"charset" gorm:"type:varchar(20);default:'utf8mb4'"`
	UseSSL         bool           `json:"use_ssl" gorm:"default:false"`
	CreatedBy      string         `json:"created_by" gorm:"type:varchar(100)"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联的执行记录
	ExecutionRecords []ExecutionRecord `json:"execution_records,omitempty" gorm:"foreignKey:ConnectionID"`
}

// TableName 返回表名
func (Connection) TableName() string {
	return "connections"
}

// DSN 生成数据库连接DSN（不包含密码）
func (c *Connection) DSN() string {
	return ""
}

// IsProduction 检查是否为生产环境
func (c *Connection) IsProduction() bool {
	return c.Environment == EnvProduction
}

// ConnectionResponse 连接响应模型（不包含敏感信息）
type ConnectionResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Environment    Environment `json:"environment"`
	Host           string      `json:"host"`
	Port           int         `json:"port"`
	Username       string      `json:"username"`
	DatabaseName   string      `json:"database_name"`
	Description    *string     `json:"description"`
	ConnectTimeout int         `json:"connect_timeout"`
	Charset        string      `json:"charset"`
	UseSSL         bool        `json:"use_ssl"`
	CreatedBy      string      `json:"created_by"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// ToResponse 转换为响应模型
func (c *Connection) ToResponse() ConnectionResponse {
	return ConnectionResponse{
		ID:             c.ID,
		Name:           c.Name,
		Environment:    c.Environment,
		Host:           c.Host,
		Port:           c.Port,
		Username:       c.Username,
		DatabaseName:   c.DatabaseName,
		Description:    c.Description,
		ConnectTimeout: c.ConnectTimeout,
		Charset:        c.Charset,
		UseSSL:         c.UseSSL,
		CreatedBy:      c.CreatedBy,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
