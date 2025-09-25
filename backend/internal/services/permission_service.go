package services

import (
	"fmt"
	"time"

	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionService 权限管理服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// HasPermission 检查用户是否有指定权限
func (s *PermissionService) HasPermission(userID, permissionKey string) (bool, error) {
	user, err := s.getUserWithPermissions(userID)
	if err != nil {
		return false, err
	}

	if !user.IsActive {
		return false, nil
	}

	resource, action := models.ParsePermissionKey(permissionKey)

	// 1. 检查用户特殊权限（优先级最高）
	userPermission, err := s.getUserPermission(userID, resource, action)
	if err == nil && userPermission != nil {
		if userPermission.IsExpired() {
			// 过期权限自动删除
			s.db.Delete(userPermission)
		} else {
			return userPermission.IsGranted, nil
		}
	}

	// 2. 检查角色权限
	return s.hasRolePermission(user.Role, resource, action)
}

// GrantUserPermission 授予用户特殊权限
func (s *PermissionService) GrantUserPermission(userID, permissionKey string, expiresAt *time.Time, grantedBy string) error {
	resource, action := models.ParsePermissionKey(permissionKey)

	// 检查权限是否存在
	permission, err := s.getPermissionByKey(resource, action)
	if err != nil {
		return fmt.Errorf("权限不存在: %v", err)
	}

	userPermission := &models.UserPermission{
		ID:           uuid.New().String(),
		UserID:       userID,
		PermissionID: permission.ID,
		IsGranted:    true,
		ExpiresAt:    expiresAt,
		CreatedBy:    grantedBy,
	}

	return s.db.Create(userPermission).Error
}

// RevokeUserPermission 撤销用户特殊权限
func (s *PermissionService) RevokeUserPermission(userID, permissionKey string, revokedBy string) error {
	resource, action := models.ParsePermissionKey(permissionKey)

	permission, err := s.getPermissionByKey(resource, action)
	if err != nil {
		return fmt.Errorf("权限不存在: %v", err)
	}

	// 创建拒绝权限记录
	userPermission := &models.UserPermission{
		ID:           uuid.New().String(),
		UserID:       userID,
		PermissionID: permission.ID,
		IsGranted:    false,
		CreatedBy:    revokedBy,
	}

	return s.db.Create(userPermission).Error
}

// GetUserPermissions 获取用户所有权限
func (s *PermissionService) GetUserPermissions(userID string) ([]string, error) {
	user, err := s.getUserWithPermissions(userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return []string{}, nil
	}

	// 获取角色默认权限
	rolePermissions := models.GetDefaultPermissions()[user.Role]
	permissionMap := make(map[string]bool)
	for _, perm := range rolePermissions {
		permissionMap[perm] = true
	}

	// 获取用户特殊权限
	var userPermissions []models.UserPermission
	err = s.db.Preload("Permission").
		Where("user_id = ? AND (expires_at IS NULL OR expires_at > ?)", userID, time.Now()).
		Find(&userPermissions).Error
	if err != nil {
		return nil, err
	}

	// 应用用户特殊权限（覆盖角色权限）
	for _, up := range userPermissions {
		key := models.PermissionKey(up.Permission.Resource, up.Permission.Action)
		if up.IsGranted {
			permissionMap[key] = true
		} else {
			delete(permissionMap, key)
		}
	}

	// 转换为权限列表
	permissions := make([]string, 0, len(permissionMap))
	for perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CheckExecutionPermission 检查DDL执行权限（包含环境和表检查）
func (s *PermissionService) CheckExecutionPermission(userID, connectionID, tableName string, ddlType models.DDLType) error {
	// 1. 基本执行权限检查
	hasPermission, err := s.HasPermission(userID, models.PermissionExecutionExecute)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("用户没有DDL执行权限")
	}

	// 2. 获取连接信息检查环境
	var connection models.Connection
	err = s.db.First(&connection, "id = ?", connectionID).Error
	if err != nil {
		return fmt.Errorf("连接不存在")
	}

	// 3. 生产环境额外检查
	if connection.Environment == "prod" {
		hasDangerousPermission, err := s.HasPermission(userID, models.PermissionDangerousOperations)
		if err != nil {
			return err
		}
		if !hasDangerousPermission {
			return fmt.Errorf("生产环境操作需要危险操作权限")
		}
	}

	// 4. 危险DDL类型检查
	dangerousDDLTypes := []models.DDLType{"drop_column", "drop_index"}
	for _, dangerousType := range dangerousDDLTypes {
		if ddlType == dangerousType {
			hasDangerousPermission, err := s.HasPermission(userID, models.PermissionDangerousOperations)
			if err != nil {
				return err
			}
			if !hasDangerousPermission {
				return fmt.Errorf("危险DDL操作需要特殊权限")
			}
		}
	}

	return nil
}

// InitializeDefaultPermissions 初始化默认权限数据
func (s *PermissionService) InitializeDefaultPermissions() error {
	// 1. 创建权限记录
	permissions := models.GetAllPermissions()
	for i := range permissions {
		permissions[i].ID = uuid.New().String()
		permissions[i].IsActive = true
	}

	// 使用事务创建权限
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 清空现有权限（开发环境）
		if err := tx.Where("1 = 1").Delete(&models.Permission{}).Error; err != nil {
			return err
		}

		// 创建权限记录
		if err := tx.Create(&permissions).Error; err != nil {
			return err
		}

		// 2. 创建角色权限关联
		defaultPermissions := models.GetDefaultPermissions()
		for role, permKeys := range defaultPermissions {
			for _, permKey := range permKeys {
				resource, action := models.ParsePermissionKey(permKey)

				var permission models.Permission
				if err := tx.Where("resource = ? AND action = ?", resource, action).First(&permission).Error; err != nil {
					continue
				}

				rolePermission := &models.RolePermission{
					ID:           uuid.New().String(),
					Role:         role,
					PermissionID: permission.ID,
				}

				if err := tx.Create(rolePermission).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// 私有方法

func (s *PermissionService) getUserWithPermissions(userID string) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, "id = ?", userID).Error
	return &user, err
}

func (s *PermissionService) getUserPermission(userID, resource, action string) (*models.UserPermission, error) {
	var userPermission models.UserPermission
	err := s.db.Preload("Permission").
		Joins("JOIN permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ? AND permissions.resource = ? AND permissions.action = ?", userID, resource, action).
		Where("user_permissions.expires_at IS NULL OR user_permissions.expires_at > ?", time.Now()).
		Order("user_permissions.created_at DESC").
		First(&userPermission).Error

	if err != nil {
		return nil, err
	}
	return &userPermission, nil
}

func (s *PermissionService) hasRolePermission(role models.UserRole, resource, action string) (bool, error) {
	var count int64
	err := s.db.Model(&models.RolePermission{}).
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role = ? AND permissions.resource = ? AND permissions.action = ?", role, resource, action).
		Where("permissions.is_active = ?", true).
		Count(&count).Error

	return count > 0, err
}

func (s *PermissionService) getPermissionByKey(resource, action string) (*models.Permission, error) {
	var permission models.Permission
	err := s.db.Where("resource = ? AND action = ? AND is_active = ?", resource, action, true).First(&permission).Error
	return &permission, err
}
