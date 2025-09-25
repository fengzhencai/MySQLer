package services

import (
	"errors"
	"fmt"

	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"github.com/fengzhencai/MySQLer/backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConnectionService 连接服务
type ConnectionService struct {
	db     *gorm.DB
	cfg    *config.Config
	crypto *utils.CryptoService
}

// NewConnectionService 创建连接服务
func NewConnectionService(db *gorm.DB, cfg *config.Config) *ConnectionService {
	return &ConnectionService{
		db:     db,
		cfg:    cfg,
		crypto: utils.NewCryptoService(cfg.EncryptionKey),
	}
}

// CreateConnectionRequest 创建连接请求
type CreateConnectionRequest struct {
	Name           string             `json:"name" binding:"required,min=1,max=100"`
	Environment    models.Environment `json:"environment" binding:"omitempty,oneof=prod test dev"`
	Host           string             `json:"host" binding:"required,hostname_rfc1123|ip"`
	Port           int                `json:"port" binding:"omitempty,min=1,max=65535"`
	Username       string             `json:"username" binding:"required,min=1,max=100"`
	Password       string             `json:"password" binding:"required,min=1,max=200"`
	DatabaseName   string             `json:"database_name" binding:"required,min=1,max=100"`
	Description    *string            `json:"description" binding:"omitempty,max=200"`
	ConnectTimeout int                `json:"connect_timeout" binding:"omitempty,min=1,max=60"`
	Charset        string             `json:"charset" binding:"omitempty,oneof=utf8 utf8mb4"`
	UseSSL         bool               `json:"use_ssl"`
}

// List 获取连接列表
func (s *ConnectionService) List(userID string) ([]models.ConnectionResponse, error) {
	var connections []models.Connection
	err := s.db.Find(&connections).Error
	if err != nil {
		return nil, err
	}

	var responses []models.ConnectionResponse
	for _, conn := range connections {
		responses = append(responses, conn.ToResponse())
	}

	return responses, nil
}

// Create 创建连接
func (s *ConnectionService) Create(req *CreateConnectionRequest, userID string) (*models.ConnectionResponse, error) {
	// 1. 验证请求参数
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. 测试连接
	dbConn := &utils.DatabaseConnection{
		Host:           req.Host,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		DatabaseName:   req.DatabaseName,
		ConnectTimeout: req.ConnectTimeout,
		Charset:        req.Charset,
		UseSSL:         req.UseSSL,
	}

	if err := utils.TestConnection(dbConn); err != nil {
		return nil, fmt.Errorf("连接测试失败: %v", err)
	}

	// 3. 加密密码
	encryptedPassword, err := s.crypto.Encrypt(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 4. 创建连接记录
	connection := &models.Connection{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Environment:    req.Environment,
		Host:           req.Host,
		Port:           req.Port,
		Username:       req.Username,
		Password:       encryptedPassword, // 存储加密后的密码
		DatabaseName:   req.DatabaseName,
		Description:    req.Description,
		ConnectTimeout: req.ConnectTimeout,
		Charset:        req.Charset,
		UseSSL:         req.UseSSL,
		CreatedBy:      userID,
	}

	// 设置默认值
	if connection.Port == 0 {
		connection.Port = 3306
	}
	if connection.ConnectTimeout == 0 {
		connection.ConnectTimeout = 5
	}
	if connection.Charset == "" {
		connection.Charset = "utf8mb4"
	}
	if connection.Environment == "" {
		connection.Environment = models.EnvTest
	}

	// 5. 保存到数据库
	if err := s.db.Create(connection).Error; err != nil {
		return nil, fmt.Errorf("保存连接失败: %v", err)
	}

	// 6. 返回响应
	response := connection.ToResponse()
	return &response, nil
}

// GetByID 根据ID获取连接
func (s *ConnectionService) GetByID(id string) (*models.ConnectionResponse, error) {
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("连接不存在")
		}
		return nil, err
	}

	response := connection.ToResponse()
	return &response, nil
}

// Update 更新连接
func (s *ConnectionService) Update(id string, req *CreateConnectionRequest, userID string) (*models.ConnectionResponse, error) {
	// 1. 验证请求参数
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. 查找现有连接
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("连接不存在")
		}
		return nil, err
	}

	// 3. 如果密码有变化，测试新连接
	if req.Password != "" {
		dbConn := &utils.DatabaseConnection{
			Host:           req.Host,
			Port:           req.Port,
			Username:       req.Username,
			Password:       req.Password,
			DatabaseName:   req.DatabaseName,
			ConnectTimeout: req.ConnectTimeout,
			Charset:        req.Charset,
			UseSSL:         req.UseSSL,
		}

		if err := utils.TestConnection(dbConn); err != nil {
			return nil, fmt.Errorf("连接测试失败: %v", err)
		}

		// 加密新密码
		encryptedPassword, err := s.crypto.Encrypt(req.Password)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %v", err)
		}
		connection.Password = encryptedPassword
	}

	// 4. 更新其他字段
	connection.Name = req.Name
	connection.Environment = req.Environment
	connection.Host = req.Host
	connection.Port = req.Port
	connection.Username = req.Username
	connection.DatabaseName = req.DatabaseName
	connection.Description = req.Description
	connection.ConnectTimeout = req.ConnectTimeout
	connection.Charset = req.Charset
	connection.UseSSL = req.UseSSL

	// 设置默认值
	if connection.Port == 0 {
		connection.Port = 3306
	}
	if connection.ConnectTimeout == 0 {
		connection.ConnectTimeout = 5
	}
	if connection.Charset == "" {
		connection.Charset = "utf8mb4"
	}
	if connection.Environment == "" {
		connection.Environment = models.EnvTest
	}

	// 5. 保存更新
	if err := s.db.Save(&connection).Error; err != nil {
		return nil, fmt.Errorf("更新连接失败: %v", err)
	}

	// 6. 返回响应
	response := connection.ToResponse()
	return &response, nil
}

// Delete 删除连接
func (s *ConnectionService) Delete(id string) error {
	// 1. 检查连接是否存在
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("连接不存在")
		}
		return err
	}

	// 2. 检查是否有关联的执行记录
	var count int64
	s.db.Model(&models.ExecutionRecord{}).Where("connection_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("该连接存在关联的执行记录，无法删除")
	}

	// 3. 软删除连接
	if err := s.db.Delete(&connection).Error; err != nil {
		return fmt.Errorf("删除连接失败: %v", err)
	}

	return nil
}

// TestConnection 测试连接
func (s *ConnectionService) TestConnection(id string) (map[string]interface{}, error) {
	// 1. 获取连接信息
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", id).Error
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

	// 3. 构建连接配置
	dbConn := &utils.DatabaseConnection{
		Host:           connection.Host,
		Port:           connection.Port,
		Username:       connection.Username,
		Password:       password,
		DatabaseName:   connection.DatabaseName,
		ConnectTimeout: connection.ConnectTimeout,
		Charset:        connection.Charset,
		UseSSL:         connection.UseSSL,
	}

	// 4. 测试连接并获取信息
	if err := utils.TestConnection(dbConn); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// 5. 获取数据库信息
	info, err := utils.GetDatabaseInfo(dbConn)
	if err != nil {
		return map[string]interface{}{
			"success": true,
			"message": "连接成功，但获取数据库信息失败",
			"error":   err.Error(),
		}, nil
	}

	info["success"] = true
	info["message"] = "连接测试成功"
	return info, nil
}

// TestConnectionByParams 基于传入参数测试连接（不落库）
func (s *ConnectionService) TestConnectionByParams(req *CreateConnectionRequest) (map[string]interface{}, error) {
	// 1. 复用字段校验，确保必要参数完整
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 2. 构建连接配置
	dbConn := &utils.DatabaseConnection{
		Host:           req.Host,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		DatabaseName:   req.DatabaseName,
		ConnectTimeout: req.ConnectTimeout,
		Charset:        req.Charset,
		UseSSL:         req.UseSSL,
	}

	// 3. 测试连接
	if err := utils.TestConnection(dbConn); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// 4. 获取数据库信息
	info, err := utils.GetDatabaseInfo(dbConn)
	if err != nil {
		return map[string]interface{}{
			"success": true,
			"message": "连接成功，但获取数据库信息失败",
			"error":   err.Error(),
		}, nil
	}

	info["success"] = true
	info["message"] = "连接测试成功"
	return info, nil
}

// GetDatabases 获取连接的数据库列表
func (s *ConnectionService) GetDatabases(id string) ([]string, error) {
	// 1. 获取连接信息
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", id).Error
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

	// 3. 构建连接配置
	dbConn := &utils.DatabaseConnection{
		Host:           connection.Host,
		Port:           connection.Port,
		Username:       connection.Username,
		Password:       password,
		DatabaseName:   connection.DatabaseName,
		ConnectTimeout: connection.ConnectTimeout,
		Charset:        connection.Charset,
		UseSSL:         connection.UseSSL,
	}

	// 4. 获取数据库列表
	return utils.GetDatabaseList(dbConn)
}

// GetTables 获取指定数据库的表列表
func (s *ConnectionService) GetTables(id string, database string) ([]map[string]interface{}, error) {
	// 1. 获取连接信息
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", id).Error
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

	// 3. 构建连接配置
	dbConn := &utils.DatabaseConnection{
		Host:           connection.Host,
		Port:           connection.Port,
		Username:       connection.Username,
		Password:       password,
		DatabaseName:   database, // 使用指定的数据库
		ConnectTimeout: connection.ConnectTimeout,
		Charset:        connection.Charset,
		UseSSL:         connection.UseSSL,
	}

	// 4. 获取表列表
	return utils.GetTableList(dbConn, database)
}

// GetTableSchema 获取指定表的列与索引结构
func (s *ConnectionService) GetTableSchema(id string, database string, table string) (map[string]interface{}, error) {
	// 1. 获取连接信息
	var connection models.Connection
	err := s.db.First(&connection, "id = ?", id).Error
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

	// 3. 构建连接配置
	dbConn := &utils.DatabaseConnection{
		Host:           connection.Host,
		Port:           connection.Port,
		Username:       connection.Username,
		Password:       password,
		DatabaseName:   database,
		ConnectTimeout: connection.ConnectTimeout,
		Charset:        connection.Charset,
		UseSSL:         connection.UseSSL,
	}

	// 4. 获取表结构
	return utils.GetTableSchema(dbConn, database, table)
}

// validateCreateRequest 验证创建请求
func (s *ConnectionService) validateCreateRequest(req *CreateConnectionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("连接名称不能为空")
	}
	if req.Host == "" {
		return fmt.Errorf("主机地址不能为空")
	}
	if req.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if req.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if req.DatabaseName == "" {
		return fmt.Errorf("数据库名不能为空")
	}
	if req.Port < 1 || req.Port > 65535 {
		req.Port = 3306 // 设置默认端口
	}
	if req.ConnectTimeout < 1 || req.ConnectTimeout > 60 {
		req.ConnectTimeout = 5 // 设置默认超时
	}
	if req.Charset == "" {
		req.Charset = "utf8mb4"
	}

	return nil
}
