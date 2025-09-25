package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// DockerService 基于 Docker CLI 的轻量封装（避免引入重量级 SDK）
type DockerService struct{}

// PTContainerConfig PT容器配置
type PTContainerConfig struct {
	Command     string            `json:"command"`
	Environment map[string]string `json:"environment"`
	WorkingDir  string            `json:"working_dir"`
	CPULimit    float64           `json:"cpu_limit"`
	MemoryLimit int64             `json:"memory_limit"`
	NetworkMode string            `json:"network_mode"`
	AutoRemove  bool              `json:"auto_remove"`
}

// PTContainerResult 容器执行结果
type PTContainerResult struct {
	ContainerID string        `json:"container_id"`
	ExitCode    int           `json:"exit_code"`
	Output      string        `json:"output"`
	Error       string        `json:"error"`
	Duration    time.Duration `json:"duration"`
}

// PTContainerStatus 容器状态
type PTContainerStatus struct {
	ContainerID string     `json:"container_id"`
	Status      string     `json:"status"`
	IsRunning   bool       `json:"is_running"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

// NewDockerService 创建Docker服务
func NewDockerService() (*DockerService, error) { return &DockerService{}, nil }

func dockerBinary() string {
	if runtime.GOOS == "windows" {
		return "docker.exe"
	}
	return "docker"
}

// CreatePTContainer 使用 docker create 返回容器ID
func (d *DockerService) CreatePTContainer(config *PTContainerConfig) (string, error) {
	if config == nil || config.Command == "" {
		return "", fmt.Errorf("invalid container config")
	}

	// 使用 percona/percona-toolkit 镜像执行 pt-online-schema-change
	args := []string{"create"}

	// 资源限制
	if config.CPULimit > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%.2f", config.CPULimit))
	}
	if config.MemoryLimit > 0 {
		args = append(args, "--memory", fmt.Sprintf("%d", config.MemoryLimit))
	}
	if config.AutoRemove {
		args = append(args, "--rm")
	}
	if config.WorkingDir != "" {
		args = append(args, "-w", config.WorkingDir)
	}
	if config.NetworkMode != "" {
		args = append(args, "--network", config.NetworkMode)
	}
	// 环境变量
	for k, v := range config.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// 镜像与命令
	image := "percona/percona-toolkit:latest"
	// 通过 sh -lc 执行以支持多参数命令行
	args = append(args, image, "sh", "-lc", config.Command)

	cmd := exec.Command(dockerBinary(), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker create failed: %v, output: %s", err, string(out))
	}

	containerID := strings.TrimSpace(string(out))
	if containerID == "" {
		return "", fmt.Errorf("empty container id from docker create")
	}
	return containerID, nil
}

// StartContainer 启动容器
func (d *DockerService) StartContainer(containerID string) error {
	cmd := exec.Command(dockerBinary(), "start", containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker start failed: %v, output: %s", err, string(out))
	}
	return nil
}

// StopContainer 停止容器
func (d *DockerService) StopContainer(containerID string, timeout int) error {
	t := fmt.Sprintf("%d", timeout)
	cmd := exec.Command(dockerBinary(), "stop", "-t", t, containerID)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker stop failed: %v, output: %s", err, string(out))
	}
	return nil
}

// RemoveContainer 删除容器
func (d *DockerService) RemoveContainer(containerID string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerID)
	cmd := exec.Command(dockerBinary(), args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker rm failed: %v, output: %s", err, string(out))
	}
	return nil
}

// WaitContainer 等待容器完成
func (d *DockerService) WaitContainer(containerID string) (*PTContainerResult, error) {
	start := time.Now()
	waitCmd := exec.Command(dockerBinary(), "wait", containerID)
	out, err := waitCmd.CombinedOutput()
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("docker wait failed: %v, output: %s", err, string(out))
	}
	// docker wait 输出退出码
	outStr := strings.TrimSpace(string(out))
	exitCode := 1
	fmt.Sscanf(outStr, "%d", &exitCode)

	// 获取日志
	logsCmd := exec.Command(dockerBinary(), "logs", containerID)
	logsOut, _ := logsCmd.CombinedOutput()

	return &PTContainerResult{
		ContainerID: containerID,
		ExitCode:    exitCode,
		Output:      string(logsOut),
		Error:       "",
		Duration:    duration,
	}, nil
}

// StreamContainerLogs 使用 docker logs -f 实时读取日志
func (d *DockerService) StreamContainerLogs(containerID string, callback func(string)) error {
	cmd := exec.Command(dockerBinary(), "logs", "-f", containerID)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// 合并 stdout/stderr 行读取
	readPipe := func(rdr *bufio.Reader) {
		for {
			line, err := rdr.ReadString('\n')
			if line != "" {
				callback(strings.TrimRight(line, "\r\n"))
			}
			if err != nil {
				break
			}
		}
	}

	go readPipe(bufio.NewReader(stdout))
	go readPipe(bufio.NewReader(stderr))

	// 不等待结束，由外部 WaitContainer 负责
	return nil
}

// GetContainerStatus 获取容器状态（简化：查询 docker inspect 状态）
func (d *DockerService) GetContainerStatus(containerID string) (*PTContainerStatus, error) {
	// 为简化，这里仅返回运行中占位
	return &PTContainerStatus{ContainerID: containerID, Status: "running", IsRunning: true, StartTime: time.Now()}, nil
}

// ExecutePTCommand 直接使用 docker run 执行命令（一次性）
func (d *DockerService) ExecutePTCommand(command string, config *PTContainerConfig) (*PTContainerResult, error) {
	if config == nil {
		config = &PTContainerConfig{}
	}
	start := time.Now()

	args := []string{"run"}
	if config.AutoRemove {
		args = append(args, "--rm")
	}
	if config.CPULimit > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%.2f", config.CPULimit))
	}
	if config.MemoryLimit > 0 {
		args = append(args, "--memory", fmt.Sprintf("%d", config.MemoryLimit))
	}
	if config.WorkingDir != "" {
		args = append(args, "-w", config.WorkingDir)
	}
	if config.NetworkMode != "" {
		args = append(args, "--network", config.NetworkMode)
	}
	for k, v := range config.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	image := "percona/percona-toolkit:latest"
	args = append(args, image, "sh", "-lc", command)

	cmd := exec.Command(dockerBinary(), args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return &PTContainerResult{
		ContainerID: "",
		ExitCode:    exitCode,
		Output:      outBuf.String(),
		Error:       errBuf.String(),
		Duration:    duration,
	}, nil
}

// GetDockerInfo 简化占位
func (d *DockerService) GetDockerInfo() (map[string]interface{}, error) {
	return map[string]interface{}{"driver": "cli"}, nil
}

// Close 占位
func (d *DockerService) Close() error { return nil }
