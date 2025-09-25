package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"gorm.io/gorm"
)

// SafetyService 安全防护服务
type SafetyService struct {
	db                *gorm.DB
	permissionService *PermissionService
	auditService      *AuditService
}

// NewSafetyService 创建安全防护服务
func NewSafetyService(db *gorm.DB, permissionService *PermissionService, auditService *AuditService) *SafetyService {
	return &SafetyService{
		db:                db,
		permissionService: permissionService,
		auditService:      auditService,
	}
}

// DangerousOperationRequest 危险操作请求
type DangerousOperationRequest struct {
	UserID       string         `json:"user_id"`
	ConnectionID string         `json:"connection_id"`
	TableName    string         `json:"table_name"`
	DatabaseName string         `json:"database_name"`
	DDLType      models.DDLType `json:"ddl_type"`
	OriginalDDL  string         `json:"original_ddl"`
	Reason       string         `json:"reason"`    // 操作原因
	TicketID     *string        `json:"ticket_id"` // 工单ID
}

// SafetyCheckResult 安全检查结果
type SafetyCheckResult struct {
	IsSafe            bool     `json:"is_safe"`
	RiskLevel         string   `json:"risk_level"` // low, medium, high, critical
	Warnings          []string `json:"warnings"`
	Blocks            []string `json:"blocks"`             // 阻止原因
	Suggestions       []string `json:"suggestions"`        // 建议
	RequiredApprovals int      `json:"required_approvals"` // 需要的审批数量
}

// CheckDangerousOperation 检查危险操作
func (s *SafetyService) CheckDangerousOperation(req *DangerousOperationRequest) (*SafetyCheckResult, error) {
	result := &SafetyCheckResult{
		IsSafe:            true,
		RiskLevel:         "low",
		Warnings:          []string{},
		Blocks:            []string{},
		Suggestions:       []string{},
		RequiredApprovals: 0,
	}

	// 1. 获取连接信息
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", req.ConnectionID).Error
	if err != nil {
		return nil, fmt.Errorf("连接不存在")
	}

	// 2. 基础权限检查
	hasPermission, err := s.permissionService.HasPermission(req.UserID, models.PermissionExecutionExecute)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		result.IsSafe = false
		result.Blocks = append(result.Blocks, "用户没有DDL执行权限")
		return result, nil
	}

	// 3. 环境风险评估
	s.checkEnvironmentRisk(&connection, result)

	// 4. DDL类型风险评估
	s.checkDDLTypeRisk(req.DDLType, req.OriginalDDL, result)

	// 5. 表结构风险评估
	s.checkTableRisk(req.ConnectionID, req.DatabaseName, req.TableName, result)

	// 6. 时间窗口检查
	s.checkTimeWindow(result)

	// 7. 操作频率检查
	s.checkOperationFrequency(req.UserID, result)

	// 8. 工单检查（生产环境高风险操作）
	if connection.Environment == "prod" && result.RiskLevel == "high" {
		s.checkTicketRequirement(req.TicketID, result)
	}

	// 9. 最终风险等级决策
	s.finalizeRiskAssessment(result)

	return result, nil
}

// 环境风险检查
func (s *SafetyService) checkEnvironmentRisk(connection *models.Connection, result *SafetyCheckResult) {
	switch connection.Environment {
	case "prod":
		result.RiskLevel = "high"
		result.Warnings = append(result.Warnings, "这是生产环境操作，风险极高")
		result.Suggestions = append(result.Suggestions, "建议在测试环境验证后再执行")
		result.RequiredApprovals = 2
	case "test":
		if result.RiskLevel == "low" {
			result.RiskLevel = "medium"
		}
		result.Warnings = append(result.Warnings, "测试环境操作请谨慎")
	case "dev":
		result.Suggestions = append(result.Suggestions, "开发环境操作相对安全")
	}
}

// DDL类型风险检查
func (s *SafetyService) checkDDLTypeRisk(ddlType models.DDLType, originalDDL string, result *SafetyCheckResult) {
	ddlUpper := strings.ToUpper(originalDDL)

	switch ddlType {
	case "drop_column":
		result.RiskLevel = "high"
		result.Warnings = append(result.Warnings, "删除列操作不可逆，数据将永久丢失")
		result.Suggestions = append(result.Suggestions, "建议先备份相关数据")

	case "drop_index":
		result.RiskLevel = "medium"
		result.Warnings = append(result.Warnings, "删除索引可能影响查询性能")
		result.Suggestions = append(result.Suggestions, "请确认该索引不被重要查询使用")

	case "modify_column":
		if strings.Contains(ddlUpper, "NOT NULL") {
			result.RiskLevel = "medium"
			result.Warnings = append(result.Warnings, "将列修改为NOT NULL可能导致现有NULL数据报错")
		}
		if strings.Contains(ddlUpper, "DROP DEFAULT") {
			result.Warnings = append(result.Warnings, "删除默认值可能影响新插入的数据")
		}

	case "add_column":
		if strings.Contains(ddlUpper, "NOT NULL") && !strings.Contains(ddlUpper, "DEFAULT") {
			result.Warnings = append(result.Warnings, "添加非空列但无默认值，可能导致现有数据报错")
		}

	case "fragment":
		result.Warnings = append(result.Warnings, "表重建操作会锁表，请在业务低峰期执行")
		result.Suggestions = append(result.Suggestions, "建议监控表大小，预估执行时间")
	}

	// 检查特殊关键词
	dangerousKeywords := []string{
		"TRUNCATE", "DROP TABLE", "DROP DATABASE",
		"ALTER TABLE", "RENAME TO",
	}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(ddlUpper, keyword) {
			if keyword == "TRUNCATE" || keyword == "DROP TABLE" || keyword == "DROP DATABASE" {
				result.RiskLevel = "critical"
				result.Blocks = append(result.Blocks, fmt.Sprintf("检测到极危险操作: %s", keyword))
			} else {
				result.RiskLevel = "high"
				result.Warnings = append(result.Warnings, fmt.Sprintf("检测到高风险操作: %s", keyword))
			}
		}
	}
}

// 表结构风险检查
func (s *SafetyService) checkTableRisk(connectionID, databaseName, tableName string, result *SafetyCheckResult) {
	// 检查表大小和行数（这里简化处理，实际应该查询information_schema）
	// 大表操作风险更高

	// 检查是否为系统表
	systemTables := []string{
		"mysql", "information_schema", "performance_schema", "sys",
	}

	for _, sysTable := range systemTables {
		if strings.EqualFold(databaseName, sysTable) {
			result.RiskLevel = "critical"
			result.Blocks = append(result.Blocks, "禁止操作系统数据库")
			return
		}
	}

	// 检查重要业务表（这里可以配置化）
	importantTables := []string{
		"users", "orders", "payments", "accounts",
	}

	for _, importantTable := range importantTables {
		if strings.EqualFold(tableName, importantTable) {
			if result.RiskLevel == "low" || result.RiskLevel == "medium" {
				result.RiskLevel = "high"
			}
			result.Warnings = append(result.Warnings, "这是重要业务表，操作需谨慎")
			result.RequiredApprovals++
		}
	}
}

// 时间窗口检查
func (s *SafetyService) checkTimeWindow(result *SafetyCheckResult) {
	now := time.Now()
	hour := now.Hour()

	// 工作时间（9-18点）风险较高
	if hour >= 9 && hour <= 18 {
		result.Warnings = append(result.Warnings, "当前为工作时间，建议在业务低峰期执行")
	}

	// 周末相对安全
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		result.Suggestions = append(result.Suggestions, "周末执行相对安全")
	}
}

// 操作频率检查
func (s *SafetyService) checkOperationFrequency(userID string, result *SafetyCheckResult) {
	// 检查用户最近1小时的操作次数
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	var count int64
	s.db.Model(&models.AuditLog{}).
		Where("user_id = ? AND action LIKE '%execution%' AND created_at > ?", userID, oneHourAgo).
		Count(&count)

	if count > 10 {
		result.Warnings = append(result.Warnings, "您在过去1小时内执行了大量操作，请注意操作频率")
	}

	// 检查最近的失败操作
	var failedCount int64
	s.db.Model(&models.AuditLog{}).
		Where("user_id = ? AND status = ? AND created_at > ?", userID, models.StatusFailed, oneHourAgo).
		Count(&failedCount)

	if failedCount > 3 {
		result.Warnings = append(result.Warnings, "您在过去1小时内有多次失败操作，建议检查操作方式")
	}
}

// 工单检查
func (s *SafetyService) checkTicketRequirement(ticketID *string, result *SafetyCheckResult) {
	if ticketID == nil || *ticketID == "" {
		result.Blocks = append(result.Blocks, "生产环境高风险操作需要提供工单号")
		result.IsSafe = false
		return
	}

	// TODO: 实际项目中应该集成工单系统API验证工单状态
	result.Suggestions = append(result.Suggestions, "已提供工单号，请确保工单已审批通过")
}

// 最终风险评估
func (s *SafetyService) finalizeRiskAssessment(result *SafetyCheckResult) {
	// 如果有阻止原因，则不安全
	if len(result.Blocks) > 0 {
		result.IsSafe = false
		result.RiskLevel = "critical"
		return
	}

	// 根据警告数量调整风险等级
	if len(result.Warnings) > 3 {
		if result.RiskLevel == "low" {
			result.RiskLevel = "medium"
		} else if result.RiskLevel == "medium" {
			result.RiskLevel = "high"
		}
	}

	// 设置审批要求
	switch result.RiskLevel {
	case "critical":
		result.RequiredApprovals = 3
		result.IsSafe = false
	case "high":
		if result.RequiredApprovals < 2 {
			result.RequiredApprovals = 2
		}
	case "medium":
		if result.RequiredApprovals < 1 {
			result.RequiredApprovals = 1
		}
	}
}

// CreateSafetyCheck 创建安全检查记录
func (s *SafetyService) CreateSafetyCheck(userID string, req *DangerousOperationRequest, result *SafetyCheckResult) error {
	// 记录安全检查的审计日志
	auditLog := &models.AuditLog{
		UserID:       &userID,
		Action:       "safety_check",
		ResourceType: stringPtr("execution"),
		Status:       models.AuditStatusSuccess,
		RiskLevel:    result.RiskLevel,
		CreatedAt:    time.Now(),
	}

	if !result.IsSafe {
		auditLog.Status = models.AuditStatusBlocked
		if len(result.Blocks) > 0 {
			errorMsg := strings.Join(result.Blocks, "; ")
			auditLog.ErrorMsg = &errorMsg
		}
	}

	return s.auditService.CreateAuditLog(auditLog)
}

// GetDangerousOperationStats 获取危险操作统计
func (s *SafetyService) GetDangerousOperationStats(days int) (map[string]interface{}, error) {
	startTime := time.Now().AddDate(0, 0, -days)

	stats := make(map[string]interface{})

	// 按风险等级统计
	var riskStats []struct {
		RiskLevel string `json:"risk_level"`
		Count     int64  `json:"count"`
	}

	err := s.db.Model(&models.AuditLog{}).
		Select("risk_level, COUNT(*) as count").
		Where("action LIKE '%execution%' AND created_at > ?", startTime).
		Group("risk_level").
		Scan(&riskStats).Error

	if err != nil {
		return nil, err
	}

	stats["risk_distribution"] = riskStats

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	err = s.db.Model(&models.AuditLog{}).
		Select("status, COUNT(*) as count").
		Where("action LIKE '%execution%' AND created_at > ?", startTime).
		Group("status").
		Scan(&statusStats).Error

	if err != nil {
		return nil, err
	}

	stats["status_distribution"] = statusStats

	// 被阻止的操作数量
	var blockedCount int64
	s.db.Model(&models.AuditLog{}).
		Where("status = ? AND created_at > ?", models.AuditStatusBlocked, startTime).
		Count(&blockedCount)

	stats["blocked_operations"] = blockedCount

	return stats, nil
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
