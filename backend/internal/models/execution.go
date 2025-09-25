package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DDLType DDL操作类型
type DDLType string

const (
	DDLFragment     DDLType = "fragment"      // 碎片整理
	DDLAddColumn    DDLType = "add_column"    // 添加列
	DDLModifyColumn DDLType = "modify_column" // 修改列
	DDLDropColumn   DDLType = "drop_column"   // 删除列
	DDLAddIndex     DDLType = "add_index"     // 添加索引
	DDLDropIndex    DDLType = "drop_index"    // 删除索引
	DDLOther        DDLType = "other"         // 其他类型
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"   // 等待执行
	StatusRunning   ExecutionStatus = "running"   // 执行中
	StatusCompleted ExecutionStatus = "completed" // 执行完成
	StatusFailed    ExecutionStatus = "failed"    // 执行失败
	StatusCancelled ExecutionStatus = "cancelled" // 手动取消
)

// ExecutionParams 执行参数
type ExecutionParams struct {
	ChunkSize       int    `json:"chunk_size"`        // 块大小
	MaxLoad         string `json:"max_load"`          // 最大负载
	CriticalLoad    string `json:"critical_load"`     // 临界负载
	Charset         string `json:"charset"`           // 字符集
	LockWaitTimeout int    `json:"lock_wait_timeout"` // 锁等待超时
	OtherParams     string `json:"other_params"`      // 其他参数
}

// ExecutionRecord 执行记录模型
type ExecutionRecord struct {
	ID               string           `json:"id" gorm:"type:varchar(36);primaryKey"`
	ConnectionID     string           `json:"connection_id" gorm:"type:varchar(36);not null;index"`
	TargetTableName  string           `json:"table_name" gorm:"type:varchar(200);not null"`
	DatabaseName     string           `json:"database_name" gorm:"type:varchar(100);not null;index:idx_table"`
	DDLType          *DDLType         `json:"ddl_type" gorm:"type:enum('fragment','add_column','modify_column','drop_column','add_index','drop_index','other')"`
	OriginalDDL      *string          `json:"original_ddl" gorm:"type:text"`
	GeneratedCommand string           `json:"generated_command" gorm:"type:text;not null"`
	ExecutionParams  *ExecutionParams `json:"execution_params" gorm:"type:json"`
	Status           ExecutionStatus  `json:"status" gorm:"type:enum('pending','running','completed','failed','cancelled');default:'pending';index"`
	StartTime        *time.Time       `json:"start_time"`
	EndTime          *time.Time       `json:"end_time"`
	DurationSeconds  *int             `json:"duration_seconds"`
	ProcessedRows    int64            `json:"processed_rows" gorm:"default:0"`
	TotalRows        int64            `json:"total_rows" gorm:"default:0"`
	AvgSpeed         *float64         `json:"avg_speed" gorm:"type:decimal(10,2)"`
	ContainerID      *string          `json:"container_id" gorm:"type:varchar(64)"`
	ExecutionLogs    *string          `json:"execution_logs" gorm:"type:longtext"`
	ErrorMessage     *string          `json:"error_message" gorm:"type:text"`
	CreatedBy        string           `json:"created_by" gorm:"type:varchar(100);index"`
	CreatedAt        time.Time        `json:"created_at" gorm:"index"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `json:"-" gorm:"index"`

	// 关联的连接信息
	Connection Connection `json:"connection,omitempty" gorm:"foreignKey:ConnectionID"`
}

// TableName 返回表名
func (ExecutionRecord) TableName() string {
	return "execution_records"
}

// IsRunning 检查是否正在执行
func (e *ExecutionRecord) IsRunning() bool {
	return e.Status == StatusRunning
}

// IsCompleted 检查是否已完成
func (e *ExecutionRecord) IsCompleted() bool {
	return e.Status == StatusCompleted
}

// IsFailed 检查是否执行失败
func (e *ExecutionRecord) IsFailed() bool {
	return e.Status == StatusFailed
}

// CanCancel 检查是否可以取消
func (e *ExecutionRecord) CanCancel() bool {
	return e.Status == StatusPending || e.Status == StatusRunning
}

// GetProgress 计算执行进度（百分比）
func (e *ExecutionRecord) GetProgress() float64 {
	if e.TotalRows == 0 {
		return 0
	}
	return float64(e.ProcessedRows) / float64(e.TotalRows) * 100
}

// GetDuration 获取执行时长
func (e *ExecutionRecord) GetDuration() time.Duration {
	if e.StartTime == nil {
		return 0
	}

	endTime := time.Now()
	if e.EndTime != nil {
		endTime = *e.EndTime
	}

	return endTime.Sub(*e.StartTime)
}

// ExecutionParams 的 GORM 接口实现
func (p ExecutionParams) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ExecutionParams) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ExecutionParams", value)
	}

	return json.Unmarshal(bytes, p)
}
