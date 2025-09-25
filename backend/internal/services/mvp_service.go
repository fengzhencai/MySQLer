package services

import (
	"fmt"
	"strings"

	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/utils"
)

// MVPService 最小可用执行服务（无持久化、无鉴权、一次性执行）
type MVPService struct {
	cfg           *config.Config
	dockerService *utils.DockerService
}

func NewMVPService(cfg *config.Config) *MVPService {
	ds, _ := utils.NewDockerService()
	return &MVPService{cfg: cfg, dockerService: ds}
}

// MVPParams 额外执行参数
type MVPParams struct {
	ChunkSize int    `json:"chunk_size"`
	DryRun    bool   `json:"dry_run"`
	Charset   string `json:"charset"`
}

// MVPRequest 预览/执行请求
type MVPRequest struct {
	Host         string     `json:"host" binding:"required"`
	Port         int        `json:"port" binding:"required"`
	Username     string     `json:"username" binding:"required"`
	Password     string     `json:"password" binding:"required"`
	Database     string     `json:"database" binding:"required"`
	Table        string     `json:"table" binding:"required"`
	DDLStatement string     `json:"ddl_statement" binding:"required"`
	Params       *MVPParams `json:"params"`
}

// MVPPreviewResponse 预览响应
type MVPPreviewResponse struct {
	GeneratedCommand string `json:"generated_command"`
}

// MVPExecuteResponse 执行响应
type MVPExecuteResponse struct {
	GeneratedCommand string `json:"generated_command"`
	ExitCode         int    `json:"exit_code"`
	Stdout           string `json:"stdout"`
	Stderr           string `json:"stderr"`
}

// BuildCommands 根据请求构建预览与执行命令（预览隐藏密码，执行使用$MYSQL_PWD）
func (s *MVPService) BuildCommands(req *MVPRequest, forceDryRun bool) (previewCmd string, execCmd string, password string, err error) {
	// 容器内访问宿主MySQL: 本地地址需要替换
	host := req.Host
	if strings.EqualFold(host, "localhost") || host == "127.0.0.1" {
		host = "host.docker.internal"
	}

	// 构建连接与表信息（不强依赖真实表信息）
	dbConn := &utils.DatabaseConnection{
		Host:           host,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		DatabaseName:   req.Database,
		ConnectTimeout: 5,
		Charset:        "utf8mb4",
		UseSSL:         false,
	}
	tableInfo := &utils.TableInfo{Database: req.Database, Table: req.Table}

	builder := utils.NewPTCommandBuilder(dbConn, tableInfo)

	// 设置选项
	opts := &utils.PTOptions{
		ChunkSize:    0,
		MaxLoad:      "",
		CriticalLoad: "",
		Charset:      "",
		Execute:      true,
		Print:        true,
		Statistics:   true,
		DropOldTable: true,
		DryRun:       false,
	}
	if req.Params != nil {
		if req.Params.ChunkSize > 0 {
			opts.ChunkSize = req.Params.ChunkSize
		}
		if req.Params.Charset != "" {
			opts.Charset = req.Params.Charset
		}
		opts.DryRun = req.Params.DryRun
	}
	if forceDryRun {
		opts.DryRun = true
	}
	builder.SetOptions(opts)

	// 先生成实际命令（包含真实密码）
	realCmd, err := builder.BuildCustomDDLCommand(req.DDLStatement)
	if err != nil {
		return
	}

	// 再生成预览命令（隐藏密码）
	previewCmd, err = builder.PreviewCommand()
	if err != nil {
		return
	}

	// 执行命令：将真实密码替换为环境变量引用
	password = req.Password
	needle := fmt.Sprintf("--password=%s", password)
	execCmd = strings.ReplaceAll(realCmd, needle, "--password=$MYSQL_PWD")

	return
}

// Execute 在容器中执行命令
func (s *MVPService) Execute(req *MVPRequest) (*MVPExecuteResponse, error) {
	preview, execCmd, pwd, err := s.BuildCommands(req, false)
	if err != nil {
		return nil, err
	}

	cfg := &utils.PTContainerConfig{
		CPULimit:    2.0,
		MemoryLimit: 2 * 1024 * 1024 * 1024,
		NetworkMode: "bridge",
		AutoRemove:  true,
		WorkingDir:  "/tmp",
		Environment: map[string]string{
			"MYSQL_PWD": pwd,
		},
	}

	result, err := s.dockerService.ExecutePTCommand(execCmd, cfg)
	if err != nil {
		return nil, err
	}

	return &MVPExecuteResponse{
		GeneratedCommand: preview,
		ExitCode:         result.ExitCode,
		Stdout:           result.Output,
		Stderr:           result.Error,
	}, nil
}
