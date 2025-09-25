package services

import (
	"encoding/json"

	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"gorm.io/gorm"
)

// AuditService 审计服务
type AuditService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAuditService 创建审计服务
func NewAuditService(db *gorm.DB, cfg *config.Config) *AuditService {
	return &AuditService{
		db:  db,
		cfg: cfg,
	}
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

// Log 记录审计日志
func (s *AuditService) Log(req *AuditLogRequest) error {
	auditLog := &models.AuditLog{
		UserID:       req.UserID,
		Username:     req.Username,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		RequestData:  req.RequestData,
		ResponseData: req.ResponseData,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
	}

	return s.db.Create(auditLog).Error
}

// List 获取审计日志列表
func (s *AuditService) List() ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.Order("created_at desc").Find(&logs).Error
	return logs, err
}

// GetByID 根据ID获取审计日志
func (s *AuditService) GetByID(id int64) (*models.AuditLog, error) {
	var log models.AuditLog
	err := s.db.First(&log, "id = ?", id).Error
	return &log, err
}

// CreateAuditLog 创建审计日志
func (s *AuditService) CreateAuditLog(auditLog *models.AuditLog) error {
	return s.db.Create(auditLog).Error
}