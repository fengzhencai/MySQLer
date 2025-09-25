package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fengzhencai/MySQLer/backend/internal/config"
	"github.com/fengzhencai/MySQLer/backend/internal/models"
	"github.com/fengzhencai/MySQLer/backend/internal/utils"
	"gorm.io/gorm"
)

// ExecutionEngine 执行引擎
type ExecutionEngine struct {
	db            *gorm.DB
	cfg           *config.Config
	dockerService *utils.DockerService
	crypto        *utils.CryptoService

	// 执行队列管理
	runningTasks  map[string]*ExecutionTask
	maxConcurrent int
	mutex         sync.RWMutex

	// 新增：任务队列与worker
	queue       chan string
	workerGroup sync.WaitGroup

	// 上下文管理
	ctx    context.Context
	cancel context.CancelFunc

	// WebSocket回调
	logBroadcaster      func(string, string)      // executionID, logLine
	progressBroadcaster func(string, interface{}) // executionID, progress
}

// ExecutionTask 执行任务
type ExecutionTask struct {
	ID          string                  `json:"id"`
	Record      *models.ExecutionRecord `json:"record"`
	Context     context.Context         `json:"-"`
	Cancel      context.CancelFunc      `json:"-"`
	ContainerID string                  `json:"container_id"`
	StartTime   time.Time               `json:"start_time"`
	LogCallback func(string)            `json:"-"`

	// 状态管理
	Status       models.ExecutionStatus `json:"status"`
	Progress     float64                `json:"progress"`
	CurrentStage string                 `json:"current_stage"`
	Speed        float64                `json:"speed"`

	mutex sync.RWMutex
}

// NewExecutionEngine 创建执行引擎
func NewExecutionEngine(db *gorm.DB, cfg *config.Config) (*ExecutionEngine, error) {
	dockerService, err := utils.NewDockerService()
	if err != nil {
		return nil, fmt.Errorf("创建Docker服务失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	engine := &ExecutionEngine{
		db:            db,
		cfg:           cfg,
		dockerService: dockerService,
		crypto:        utils.NewCryptoService(cfg.EncryptionKey),
		runningTasks:  make(map[string]*ExecutionTask),
		maxConcurrent: 10, // 最大并发执行数
		queue:         make(chan string, 100),
		ctx:           ctx,
		cancel:        cancel,
	}

	// 启动固定数量的worker
	for i := 0; i < engine.maxConcurrent; i++ {
		engine.workerGroup.Add(1)
		go engine.worker()
	}

	return engine, nil
}

// SetBroadcasters 设置WebSocket广播器
func (e *ExecutionEngine) SetBroadcasters(logBroadcaster func(string, string), progressBroadcaster func(string, interface{})) {
	e.logBroadcaster = logBroadcaster
	e.progressBroadcaster = progressBroadcaster
}

// StartExecution 将任务加入队列
func (e *ExecutionEngine) StartExecution(recordID string, logCallback func(string)) error {
	select {
	case e.queue <- recordID:
		_ = logCallback // 如需回调映射可扩展
		return nil
	default:
		return fmt.Errorf("执行队列已满，请稍后重试")
	}
}

// StopExecution 停止执行任务
func (e *ExecutionEngine) StopExecution(recordID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	task, exists := e.runningTasks[recordID]
	if !exists {
		return fmt.Errorf("任务未在执行中")
	}

	// 取消任务上下文
	task.Cancel()

	// 停止Docker容器
	if task.ContainerID != "" {
		if err := e.dockerService.StopContainer(task.ContainerID, 10); err != nil {
			// 记录错误但不返回，继续清理
			fmt.Printf("停止容器失败: %v\n", err)
		}
	}

	// 更新任务状态
	task.mutex.Lock()
	task.Status = models.StatusCancelled
	task.mutex.Unlock()

	// 更新数据库状态
	now := time.Now()
	task.Record.Status = models.StatusCancelled
	task.Record.EndTime = &now
	duration := int(now.Sub(*task.Record.StartTime).Seconds())
	task.Record.DurationSeconds = &duration

	if err := e.db.Save(task.Record).Error; err != nil {
		return fmt.Errorf("更新执行状态失败: %v", err)
	}

	// 从运行任务列表中删除
	delete(e.runningTasks, recordID)

	return nil
}

// GetTaskStatus 获取任务状态
func (e *ExecutionEngine) GetTaskStatus(recordID string) (*ExecutionTask, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	task, exists := e.runningTasks[recordID]
	if !exists {
		return nil, fmt.Errorf("任务未在执行中")
	}

	// 返回任务副本，避免并发问题
	task.mutex.RLock()
	defer task.mutex.RUnlock()

	return &ExecutionTask{
		ID:           task.ID,
		Status:       task.Status,
		Progress:     task.Progress,
		CurrentStage: task.CurrentStage,
		Speed:        task.Speed,
		StartTime:    task.StartTime,
		ContainerID:  task.ContainerID,
	}, nil
}

// GetRunningTasks 获取所有运行中的任务
func (e *ExecutionEngine) GetRunningTasks() []*ExecutionTask {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	var tasks []*ExecutionTask
	for _, task := range e.runningTasks {
		task.mutex.RLock()
		taskCopy := &ExecutionTask{
			ID:           task.ID,
			Status:       task.Status,
			Progress:     task.Progress,
			CurrentStage: task.CurrentStage,
			Speed:        task.Speed,
			StartTime:    task.StartTime,
			ContainerID:  task.ContainerID,
		}
		task.mutex.RUnlock()
		tasks = append(tasks, taskCopy)
	}

	return tasks
}

// worker 从队列取任务并执行
func (e *ExecutionEngine) worker() {
	defer e.workerGroup.Done()
	for {
		select {
		case <-e.ctx.Done():
			return
		case recordID, ok := <-e.queue:
			if !ok {
				return
			}

			// 已在运行则跳过
			e.mutex.RLock()
			_, exists := e.runningTasks[recordID]
			e.mutex.RUnlock()
			if exists {
				continue
			}

			// 加载记录
			var record models.ExecutionRecord
			if err := e.db.Preload("Connection").First(&record, "id = ?", recordID).Error; err != nil {
				continue
			}
			if record.Status != models.StatusPending {
				continue
			}

			taskCtx, taskCancel := context.WithCancel(e.ctx)
			task := &ExecutionTask{
				ID:           recordID,
				Record:       &record,
				Context:      taskCtx,
				Cancel:       taskCancel,
				StartTime:    time.Now(),
				Status:       models.StatusRunning,
				CurrentStage: "准备执行",
			}

			// 注册运行任务
			e.mutex.Lock()
			e.runningTasks[recordID] = task
			e.mutex.Unlock()

			// 更新DB状态
			now := time.Now()
			record.Status = models.StatusRunning
			record.StartTime = &now
			if err := e.db.Save(&record).Error; err != nil {
				e.mutex.Lock()
				delete(e.runningTasks, recordID)
				e.mutex.Unlock()
				taskCancel()
				continue
			}

			// 执行
			e.executeTask(task)
		}
	}
}

// executeTask 执行单个任务
func (e *ExecutionEngine) executeTask(task *ExecutionTask) {
	defer func() {
		e.mutex.Lock()
		delete(e.runningTasks, task.ID)
		e.mutex.Unlock()
	}()

	var err error
	defer func() {
		// 更新最终状态
		now := time.Now()
		task.Record.EndTime = &now
		duration := int(now.Sub(*task.Record.StartTime).Seconds())
		task.Record.DurationSeconds = &duration

		if err != nil {
			task.Record.Status = models.StatusFailed
			errorMsg := err.Error()
			task.Record.ErrorMessage = &errorMsg

			task.mutex.Lock()
			task.Status = models.StatusFailed
			task.mutex.Unlock()

			if task.LogCallback != nil {
				task.LogCallback(fmt.Sprintf("执行失败: %v", err))
			}
		} else {
			task.Record.Status = models.StatusCompleted
			task.mutex.Lock()
			task.Status = models.StatusCompleted
			task.Progress = 100.0
			task.mutex.Unlock()

			if task.LogCallback != nil {
				task.LogCallback("执行完成")
			}
		}

		// 保存最终状态
		e.db.Save(task.Record)
	}()

	// 步骤1: 准备执行环境
	e.updateStage(task, "准备执行环境")

	// 解密连接密码
	password, err := e.crypto.Decrypt(task.Record.Connection.Password)
	if err != nil {
		err = fmt.Errorf("密码解密失败: %v", err)
		return
	}

	// 步骤2: 创建Docker容器
	e.updateStage(task, "创建执行容器")

	// 处理容器内访问宿主机 MySQL：将 localhost/127.0.0.1 替换为 host.docker.internal
	cmd := task.Record.GeneratedCommand
	if h := strings.ToLower(task.Record.Connection.Host); h == "localhost" || h == "127.0.0.1" {
		cmd = strings.Replace(cmd, fmt.Sprintf("--host=%s", task.Record.Connection.Host), "--host=host.docker.internal", 1)
	}

	// 将命令中的明文密码替换为环境变量，避免特殊字符导致的解析问题
	if password != "" {
		needle := fmt.Sprintf("--password=%s", password)
		cmd = strings.ReplaceAll(cmd, needle, "--password=$MYSQL_PWD")
	}

	containerConfig := &utils.PTContainerConfig{
		Command:     cmd,
		CPULimit:    2.0,
		MemoryLimit: 2 * 1024 * 1024 * 1024, // 2GB
		NetworkMode: "bridge",
		AutoRemove:  false, // 保留容器以便获取日志
		WorkingDir:  "/tmp",
		Environment: map[string]string{
			"MYSQL_PWD": password, // 通过环境变量传递密码
		},
	}

	containerID, err := e.dockerService.CreatePTContainer(containerConfig)
	if err != nil {
		err = fmt.Errorf("创建容器失败: %v", err)
		return
	}

	task.ContainerID = containerID
	task.Record.ContainerID = &containerID

	// 步骤3: 启动容器
	e.updateStage(task, "启动执行容器")

	if err = e.dockerService.StartContainer(containerID); err != nil {
		err = fmt.Errorf("启动容器失败: %v", err)
		return
	}

	// 步骤4: 监控执行进度
	e.updateStage(task, "正在执行DDL操作")

	// 启动日志监控
	go e.monitorContainerLogs(task)

	// 等待容器完成
	result, waitErr := e.dockerService.WaitContainer(containerID)
	if waitErr != nil {
		err = fmt.Errorf("等待容器完成失败: %v", waitErr)
		return
	}

	// 检查执行结果
	if result.ExitCode != 0 {
		err = fmt.Errorf("PT工具执行失败，退出码: %d, 错误信息: %s", result.ExitCode, result.Error)
		return
	}

	// 保存执行日志
	if result.Output != "" {
		task.Record.ExecutionLogs = &result.Output
	}

	// 清理容器
	e.dockerService.RemoveContainer(containerID, true)
}

// updateStage 更新任务阶段
func (e *ExecutionEngine) updateStage(task *ExecutionTask, stage string) {
	task.mutex.Lock()
	defer task.mutex.Unlock()

	task.CurrentStage = stage

	logLine := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), stage)

	// 本地回调
	if task.LogCallback != nil {
		task.LogCallback(logLine)
	}

	// WebSocket广播
	if e.logBroadcaster != nil {
		e.logBroadcaster(task.ID, logLine)
	}
}

// updateProgress 更新任务进度
func (e *ExecutionEngine) updateProgress(task *ExecutionTask, progress float64, speed float64) {
	task.mutex.Lock()
	defer task.mutex.Unlock()

	task.Progress = progress
	task.Speed = speed

	// WebSocket广播进度
	if e.progressBroadcaster != nil {
		progressData := map[string]interface{}{
			"execution_id":  task.ID,
			"status":        string(task.Status),
			"progress":      progress,
			"current_speed": speed,
			"current_stage": task.CurrentStage,
			"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		}
		e.progressBroadcaster(task.ID, progressData)
	}
}

// monitorContainerLogs 监控容器日志
func (e *ExecutionEngine) monitorContainerLogs(task *ExecutionTask) {
	err := e.dockerService.StreamContainerLogs(task.ContainerID, func(logLine string) {
		if task.LogCallback != nil {
			task.LogCallback(logLine)
		}

		// WebSocket广播日志
		if e.logBroadcaster != nil {
			e.logBroadcaster(task.ID, logLine)
		}

		// 尝试从日志中解析进度信息
		progress, speed := e.parseProgressFromLog(logLine)
		if progress > 0 {
			e.updateProgress(task, progress, speed)
		}
	})

	if err != nil {
		if task.LogCallback != nil {
			task.LogCallback(fmt.Sprintf("日志监控错误: %v", err))
		}
	}
}

// parseProgressFromLog 从日志中解析进度信息
func (e *ExecutionEngine) parseProgressFromLog(logLine string) (float64, float64) {
	// 常见格式示例：
	// "Copying approximately 5000000 rows"
	// "Copied 3750000/5000000 rows (75%)"
	// "Current copy rate: 5420 rows/sec"

	// 简单解析括号内百分比
	if lp := strings.LastIndex(logLine, "("); lp >= 0 && strings.HasSuffix(logLine, ")") {
		pctStr := strings.TrimSuffix(logLine[lp+1:], ")")
		if strings.HasSuffix(pctStr, "%") {
			pctStr = strings.TrimSuffix(pctStr, "%")
			if p, err := strconv.ParseFloat(strings.TrimSpace(pctStr), 64); err == nil {
				return p, 0
			}
		}
	}

	// 解析速度："Current copy rate: 5420 rows/sec"
	if idx := strings.Index(logLine, "Current copy rate:"); idx >= 0 {
		tail := strings.TrimSpace(logLine[idx+len("Current copy rate:"):])
		fields := strings.Fields(tail)
		if len(fields) > 0 {
			if spd, err := strconv.ParseFloat(fields[0], 64); err == nil {
				return 0, spd
			}
		}
	}

	return 0, 0
}

// Shutdown 关闭执行引擎
func (e *ExecutionEngine) Shutdown() error {
	// 取消所有任务
	e.cancel()

	// 等待所有任务完成或超时
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// 强制停止所有运行中的任务
			e.mutex.Lock()
			for recordID := range e.runningTasks {
				e.StopExecution(recordID)
			}
			e.mutex.Unlock()
			return fmt.Errorf("关闭超时，已强制停止所有任务")

		case <-ticker.C:
			e.mutex.RLock()
			count := len(e.runningTasks)
			e.mutex.RUnlock()

			if count == 0 {
				close(e.queue)
				e.workerGroup.Wait()
				return e.dockerService.Close()
			}
		}
	}
}
