package services

import (
	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"gorm.io/gorm"
)

// Services 服务容器
type Services struct {
	Auth            *AuthService
	Connection      *ConnectionService
	Execution       *ExecutionService
	ExecutionEngine *ExecutionEngine
	User            *UserService
	Audit           *AuditService
	MVP             *MVPService
}

// NewServices 创建服务容器
func NewServices(db *gorm.DB, cfg *config.Config) (*Services, error) {
	connectionService := NewConnectionService(db, cfg)

	// 创建执行引擎
	executionEngine, err := NewExecutionEngine(db, cfg)
	if err != nil {
		return nil, err
	}

	return &Services{
		Auth:            NewAuthService(db, cfg),
		Connection:      connectionService,
		Execution:       NewExecutionService(db, cfg, connectionService),
		ExecutionEngine: executionEngine,
		User:            NewUserService(db, cfg),
		Audit:           NewAuditService(db, cfg),
		MVP:             NewMVPService(cfg),
	}, nil
}
