package services

import (
	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{
		db:  db,
		cfg: cfg,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username    string          `json:"username" binding:"required"`
	Password    string          `json:"password" binding:"required"`
	DisplayName *string         `json:"display_name"`
	Email       *string         `json:"email"`
	Role        models.UserRole `json:"role"`
}

// List 获取用户列表
func (s *UserService) List() ([]models.User, error) {
	var users []models.User
	err := s.db.Find(&users).Error
	return users, err
}

// Create 创建用户
func (s *UserService) Create(req *CreateUserRequest) (*models.User, error) {
	// TODO: 实现用户创建逻辑
	return nil, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(id string) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, "id = ?", id).Error
	return &user, err
}

// Update 更新用户
func (s *UserService) Update(id string, req *CreateUserRequest) (*models.User, error) {
	// TODO: 实现用户更新逻辑
	return nil, nil
}

// Delete 删除用户
func (s *UserService) Delete(id string) error {
	return s.db.Delete(&models.User{}, "id = ?", id).Error
}
