package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"github.com/fengzhencai/MySQLer/backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionService 执行服务
type ExecutionService struct {
	db                *gorm.DB
	cfg               *config.Config
	connectionService *ConnectionService
	crypto            *utils.CryptoService
}

// NewExecutionService 创建执行服务
func NewExecutionService(db *gorm.DB, cfg *config.Config, connectionService *ConnectionService) *ExecutionService {
	return &ExecutionService{
		db:                db,
		cfg:               cfg,
		connectionService: connectionService,
		crypto:            utils.NewCryptoService(cfg.EncryptionKey),
	}
}

// CreateExecutionRequest 创建执行请求
type CreateExecutionRequest struct {
	ConnectionID    string                  `json:"connection_id" binding:"required,uuid4"`
	TableName       string                  `json:"table_name" binding:"required,min=1,max=200"`
	DatabaseName    string                  `json:"database_name" binding:"required,min=1,max=100"`
	DDLType         *models.DDLType         `json:"ddl_type" binding:"required"`
	OriginalDDL     *string                 `json:"original_ddl" binding:"omitempty,max=2000"`
	ExecutionParams *models.ExecutionParams `json:"execution_params"`
}

// PreviewCommandRequest 预览命令请求
type PreviewCommandRequest struct {
	ConnectionID    string                  `json:"connection_id" binding:"required,uuid4"`
	TableName       string                  `json:"table_name" binding:"required,min=1,max=200"`
	DatabaseName    string                  `json:"database_name" binding:"required,min=1,max=100"`
	DDLType         string                  `json:"ddl_type" binding:"required,oneof=fragment custom"` // fragment 或 custom
	OriginalDDL     *string                 `json:"original_ddl" binding:"omitempty,max=2000"`         // 自定义DDL时需要
	ExecutionParams *models.ExecutionParams `json:"execution_params"`
}

// PreviewCommandResponse 预览命令响应
type PreviewCommandResponse struct {
	Command              string                 `json:"command"`
	RiskAnalysis         map[string]interface{} `json:"risk_analysis"`
	TableInfo            *utils.TableInfo       `json:"table_info"`
	EstimatedTime        string                 `json:"estimated_time"`
	RecommendedChunkSize int                    `json:"recommended_chunk_size"`
}

// List 获取执行记录列表（分页与过滤）
func (s *ExecutionService) List(userID string) (interface{}, error) {
	// 读取过滤参数将由handler负责，这里只保留查询逻辑的可复用函数
	// 为了兼容现有handler签名，这里返回interface{}，由调用方传入范围
	return nil, fmt.Errorf("请使用带筛选参数的ListWithFilters方法")
}

// ListWithFilters 获取执行记录列表，支持分页与过滤
func (s *ExecutionService) ListWithFilters(params map[string]interface{}) (records []models.ExecutionRecord, total int64, err error) {
	db := s.db.Model(&models.ExecutionRecord{}).Preload("Connection")

	// 过滤条件
	if v, ok := params["status"].(string); ok && v != "" {
		db = db.Where("status = ?", v)
	}
	if v, ok := params["connection_id"].(string); ok && v != "" {
		db = db.Where("connection_id = ?", v)
	}
	if v, ok := params["start_date"].(string); ok && v != "" {
		db = db.Where("created_at >= ?", v)
	}
	if v, ok := params["end_date"].(string); ok && v != "" {
		db = db.Where("created_at <= ?", v)
	}
	if v, ok := params["keyword"].(string); ok && v != "" {
		like := "%" + v + "%"
		db = db.Where("id LIKE ? OR database_name LIKE ? OR target_table_name LIKE ?", like, like, like)
	}

	// 总数
	if err = db.Count(&total).Error; err != nil {
		return
	}

	// 分页
	page := 1
	size := 20
	if v, ok := params["page"].(int); ok && v > 0 {
		page = v
	}
	if v, ok := params["size"].(int); ok && v > 0 && v <= 200 {
		size = v
	}
	offset := (page - 1) * size

	err = db.Order("created_at DESC").Limit(size).Offset(offset).Find(&records).Error
	return
}

// PreviewCommand 预览pt命令
func (s *ExecutionService) PreviewCommand(req *PreviewCommandRequest) (*PreviewCommandResponse, error) {
	// 1. 获取连接信息
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", req.ConnectionID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("连接不存在")
		}
		return nil, err
	}

	// 2. 解密密码
	password, err := s.crypto.Decrypt(connection.Password)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败: %v", err)
	}

	// 3. 构建数据库连接配置
	dbConn := &utils.DatabaseConnection{
		Host:           connection.Host,
		Port:           connection.Port,
		Username:       connection.Username,
		Password:       password,
		DatabaseName:   req.DatabaseName,
		ConnectTimeout: connection.ConnectTimeout,
		Charset:        connection.Charset,
		UseSSL:         connection.UseSSL,
	}

	// 4. 获取表信息
	tables, err := utils.GetTableList(dbConn, req.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("获取表信息失败: %v", err)
	}

	var tableInfo *utils.TableInfo
	for _, table := range tables {
		if tableName, ok := table["table_name"].(string); ok && tableName == req.TableName {
			tableInfo = &utils.TableInfo{
				Database: req.DatabaseName,
				Table:    req.TableName,
			}

			if engine, ok := table["engine"].(string); ok {
				tableInfo.Engine = engine
			}
			if rows, ok := table["table_rows"].(int64); ok {
				tableInfo.Rows = rows
			}
			if size, ok := table["data_length"].(int64); ok {
				tableInfo.Size = size
			}
			break
		}
	}

	if tableInfo == nil {
		return nil, fmt.Errorf("表 %s 不存在", req.TableName)
	}

	// 5. 构建PT命令
	builder := utils.NewPTCommandBuilder(dbConn, tableInfo)

	// 设置执行参数
	if req.ExecutionParams != nil {
		ptOptions := &utils.PTOptions{
			ChunkSize:    req.ExecutionParams.ChunkSize,
			MaxLoad:      req.ExecutionParams.MaxLoad,
			CriticalLoad: req.ExecutionParams.CriticalLoad,
			Charset:      req.ExecutionParams.Charset,
			Execute:      true,
			Print:        true,
			Statistics:   true,
			DropOldTable: true,
		}
		// 将锁等待超时映射到 --set-vars
		if req.ExecutionParams.LockWaitTimeout > 0 {
			ptOptions.SetVars = fmt.Sprintf("lock_wait_timeout=%d", req.ExecutionParams.LockWaitTimeout)
		}
		builder.SetOptions(ptOptions)
	}

	var riskAnalysis map[string]interface{}

	// 6. 根据DDL类型构建命令
	var command string
	switch req.DDLType {
	case "fragment":
		command, err = builder.BuildFragmentCommand()
		if err != nil {
			return nil, fmt.Errorf("构建碎片整理命令失败: %v", err)
		}
	case "custom":
		if req.OriginalDDL == nil || *req.OriginalDDL == "" {
			return nil, fmt.Errorf("自定义DDL时原始DDL语句不能为空")
		}
		command, err = builder.BuildCustomDDLCommand(*req.OriginalDDL)
		if err != nil {
			return nil, fmt.Errorf("构建自定义DDL命令失败: %v", err)
		}
	default:
		return nil, fmt.Errorf("不支持的DDL类型: %s", req.DDLType)
	}

	// 7. 风险分析
	riskAnalysis = builder.AnalyzeDDLRisk()

	// 8. 预览命令（隐藏密码）
	previewCommand, err := builder.PreviewCommand()
	if err != nil {
		return nil, fmt.Errorf("生成预览命令失败: %v", err)
	}

	// 在这里使用command变量进行一些逻辑检查
	if command == "" {
		return nil, fmt.Errorf("生成的命令为空")
	}

	// 9. 获取推荐的chunk-size
	recommendedChunkSize := builder.GetRecommendedChunkSize()

	return &PreviewCommandResponse{
		Command:              previewCommand,
		RiskAnalysis:         riskAnalysis,
		TableInfo:            tableInfo,
		EstimatedTime:        riskAnalysis["estimated_time"].(string),
		RecommendedChunkSize: recommendedChunkSize,
	}, nil
}

// Create 创建执行记录
func (s *ExecutionService) Create(req *CreateExecutionRequest, userID string) (*models.ExecutionRecord, error) {
	// 1. 验证连接是否存在
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", req.ConnectionID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("连接不存在")
		}
		return nil, err
	}

	// 2. 解密密码
	password, err := s.crypto.Decrypt(connection.Password)
	if err != nil {
		return nil, fmt.Errorf("密码解密失败: %v", err)
	}

	// 3. 构建数据库连接配置
	dbConn := &utils.DatabaseConnection{
		Host:           connection.Host,
		Port:           connection.Port,
		Username:       connection.Username,
		Password:       password,
		DatabaseName:   req.DatabaseName,
		ConnectTimeout: connection.ConnectTimeout,
		Charset:        connection.Charset,
		UseSSL:         connection.UseSSL,
	}

	// 4. 获取表信息
	tables, err := utils.GetTableList(dbConn, req.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("获取表信息失败: %v", err)
	}

	var tableInfo *utils.TableInfo
	for _, table := range tables {
		if tableName, ok := table["table_name"].(string); ok && tableName == req.TableName {
			tableInfo = &utils.TableInfo{
				Database: req.DatabaseName,
				Table:    req.TableName,
			}

			if engine, ok := table["engine"].(string); ok {
				tableInfo.Engine = engine
			}
			if rows, ok := table["table_rows"].(int64); ok {
				tableInfo.Rows = rows
			}
			if size, ok := table["data_length"].(int64); ok {
				tableInfo.Size = size
			}
			break
		}
	}

	if tableInfo == nil {
		return nil, fmt.Errorf("表 %s 不存在", req.TableName)
	}

	// 5. 构建PT命令
	builder := utils.NewPTCommandBuilder(dbConn, tableInfo)

	// 设置执行参数
	if req.ExecutionParams != nil {
		ptOptions := &utils.PTOptions{
			ChunkSize:    req.ExecutionParams.ChunkSize,
			MaxLoad:      req.ExecutionParams.MaxLoad,
			CriticalLoad: req.ExecutionParams.CriticalLoad,
			Charset:      req.ExecutionParams.Charset,
			Execute:      true,
			Print:        true,
			Statistics:   true,
			DropOldTable: true,
		}
		if req.ExecutionParams.LockWaitTimeout > 0 {
			ptOptions.SetVars = fmt.Sprintf("lock_wait_timeout=%d", req.ExecutionParams.LockWaitTimeout)
		}
		builder.SetOptions(ptOptions)
	}

	var command string

	// 6. 根据DDL类型构建命令
	if req.DDLType != nil {
		switch *req.DDLType {
		case models.DDLFragment:
			command, err = builder.BuildFragmentCommand()
		default:
			if req.OriginalDDL == nil || *req.OriginalDDL == "" {
				return nil, fmt.Errorf("自定义DDL时原始DDL语句不能为空")
			}
			command, err = builder.BuildCustomDDLCommand(*req.OriginalDDL)
		}
	} else {
		return nil, fmt.Errorf("DDL类型不能为空")
	}

	if err != nil {
		return nil, fmt.Errorf("构建PT命令失败: %v", err)
	}

	// 7. 创建执行记录
	record := &models.ExecutionRecord{
		ID:               uuid.New().String(),
		ConnectionID:     req.ConnectionID,
		TargetTableName:  req.TableName,
		DatabaseName:     req.DatabaseName,
		DDLType:          req.DDLType,
		OriginalDDL:      req.OriginalDDL,
		GeneratedCommand: command,
		ExecutionParams:  req.ExecutionParams,
		Status:           models.StatusPending,
		TotalRows:        tableInfo.Rows,
		CreatedBy:        userID,
	}

	// 设置默认执行参数
	if record.ExecutionParams == nil {
		record.ExecutionParams = &models.ExecutionParams{
			ChunkSize:    s.cfg.PTDefaultChunkSize,
			MaxLoad:      s.cfg.PTDefaultMaxLoad,
			CriticalLoad: s.cfg.PTDefaultCriticalLoad,
			Charset:      "utf8mb4",
		}
	}

	// 8. 保存到数据库
	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("创建执行记录失败: %v", err)
	}

	return record, nil
}

// GetByID 根据ID获取执行记录
func (s *ExecutionService) GetByID(id string) (*models.ExecutionRecord, error) {
	var record models.ExecutionRecord
	err := s.db.Preload("Connection").First(&record, "id = ?", id).Error
	return &record, err
}

// Stop 停止执行
func (s *ExecutionService) Stop(id string) error {
	// 调用执行引擎停止任务需通过引擎，服务层更新状态作为兜底
	// 这里仅更新记录状态为cancelled（如果仍处于running）
	var record models.ExecutionRecord
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		return err
	}
	if record.Status == models.StatusRunning {
		now := time.Now()
		record.Status = models.StatusCancelled
		record.EndTime = &now
		duration := int(now.Sub(*record.StartTime).Seconds())
		record.DurationSeconds = &duration
		return s.db.Save(&record).Error
	}
	return fmt.Errorf("当前状态无法停止: %s", record.Status)
}

// Retry 重试执行
func (s *ExecutionService) Retry(id string, userID string) error {
	// 将记录状态重置为pending，并清理运行态字段
	var record models.ExecutionRecord
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		return err
	}
	if record.Status != models.StatusFailed && record.Status != models.StatusCancelled {
		return fmt.Errorf("仅失败或已取消的任务允许重试")
	}
	record.Status = models.StatusPending
	record.StartTime = nil
	record.EndTime = nil
	record.DurationSeconds = nil
	record.ErrorMessage = nil
	return s.db.Save(&record).Error
}

// GetLogs 获取执行日志
func (s *ExecutionService) GetLogs(id string) (string, error) {
	var record models.ExecutionRecord
	err := s.db.Select("execution_logs").First(&record, "id = ?", id).Error
	if err != nil {
		return "", err
	}

	if record.ExecutionLogs != nil {
		return *record.ExecutionLogs, nil
	}

	return "", nil
}
