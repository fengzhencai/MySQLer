package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// PTCommandBuilder pt-online-schema-change命令构建器
type PTCommandBuilder struct {
	ConnectionConfig *DatabaseConnection
	TableInfo        *TableInfo
	AlterStatement   string
	Options          *PTOptions
}

// TableInfo 表信息
type TableInfo struct {
	Database string `json:"database"`
	Table    string `json:"table"`
	Engine   string `json:"engine,omitempty"`
	Rows     int64  `json:"rows,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

// PTOptions pt工具选项
type PTOptions struct {
	ChunkSize      int    `json:"chunk_size"`      // 每次处理的行数
	MaxLoad        string `json:"max_load"`        // 最大负载
	CriticalLoad   string `json:"critical_load"`   // 临界负载
	CheckInterval  int    `json:"check_interval"`  // 检查间隔（秒）
	MaxLag         int    `json:"max_lag"`         // 最大延迟（秒）
	Charset        string `json:"charset"`         // 字符集
	Execute        bool   `json:"execute"`         // 是否执行
	Print          bool   `json:"print"`           // 是否打印SQL
	DryRun         bool   `json:"dry_run"`         // 是否仅仅模拟
	DropOldTable   bool   `json:"drop_old_table"`  // 是否删除旧表
	Statistics     bool   `json:"statistics"`      // 是否显示统计信息
	Progress       string `json:"progress"`        // 进度报告方式
	SetVars        string `json:"set_vars"`        // 设置MySQL变量
	Recursion      int    `json:"recursion"`       // 递归级别
	BinlogPosition string `json:"binlog_position"` // binlog位置
}

// DDLType DDL操作类型
type DDLType string

const (
	DDLTypeFragment     DDLType = "fragment"      // 碎片整理
	DDLTypeAddColumn    DDLType = "add_column"    // 添加列
	DDLTypeModifyColumn DDLType = "modify_column" // 修改列
	DDLTypeDropColumn   DDLType = "drop_column"   // 删除列
	DDLTypeAddIndex     DDLType = "add_index"     // 添加索引
	DDLTypeDropIndex    DDLType = "drop_index"    // 删除索引
	DDLTypeCustom       DDLType = "custom"        // 自定义DDL
)

// NewPTCommandBuilder 创建PT命令构建器
func NewPTCommandBuilder(conn *DatabaseConnection, table *TableInfo) *PTCommandBuilder {
	return &PTCommandBuilder{
		ConnectionConfig: conn,
		TableInfo:        table,
		Options:          getDefaultPTOptions(),
	}
}

// getDefaultPTOptions 获取默认PT选项
func getDefaultPTOptions() *PTOptions {
	return &PTOptions{
		ChunkSize:     1000,
		MaxLoad:       "Threads_running=25",
		CriticalLoad:  "Threads_running=50",
		CheckInterval: 1,
		MaxLag:        1,
		Charset:       "utf8mb4",
		Execute:       true,
		Print:         true,
		DryRun:        false,
		DropOldTable:  true,
		Statistics:    true,
		Progress:      "time,5",
		Recursion:     0,
	}
}

// BuildFragmentCommand 构建碎片整理命令
func (b *PTCommandBuilder) BuildFragmentCommand() (string, error) {
	if b.TableInfo.Database == "" || b.TableInfo.Table == "" {
		return "", fmt.Errorf("数据库名和表名不能为空")
	}

	b.AlterStatement = "ENGINE=INNODB"
	return b.buildCommand()
}

// BuildCustomDDLCommand 构建自定义DDL命令
func (b *PTCommandBuilder) BuildCustomDDLCommand(alterSQL string) (string, error) {
	if alterSQL == "" {
		return "", fmt.Errorf("ALTER语句不能为空")
	}

	// 验证和清理ALTER语句
	cleanSQL, err := b.validateAndCleanAlterSQL(alterSQL)
	if err != nil {
		return "", err
	}

	b.AlterStatement = cleanSQL
	return b.buildCommand()
}

// SetOptions 设置PT选项
func (b *PTCommandBuilder) SetOptions(options *PTOptions) *PTCommandBuilder {
	if options != nil {
		b.Options = options
	}
	return b
}

// buildCommand 构建pt-online-schema-change命令
func (b *PTCommandBuilder) buildCommand() (string, error) {
	var parts []string

	// 基础命令
	parts = append(parts, "pt-online-schema-change")

	// 连接参数
	parts = append(parts, fmt.Sprintf("--host=%s", b.ConnectionConfig.Host))
	parts = append(parts, fmt.Sprintf("--port=%d", b.ConnectionConfig.Port))
	parts = append(parts, fmt.Sprintf("--user=%s", b.ConnectionConfig.Username))
	parts = append(parts, fmt.Sprintf("--password=%s", b.ConnectionConfig.Password))

	// 数据库和表
	parts = append(parts, fmt.Sprintf("D=%s,t=%s", b.TableInfo.Database, b.TableInfo.Table))

	// ALTER语句
	if b.AlterStatement != "" {
		parts = append(parts, fmt.Sprintf("--alter=\"%s\"", b.AlterStatement))
	}

	// PT选项
	if b.Options.ChunkSize > 0 {
		parts = append(parts, fmt.Sprintf("--chunk-size=%d", b.Options.ChunkSize))
	}

	if b.Options.MaxLoad != "" {
		parts = append(parts, fmt.Sprintf("--max-load=%s", b.Options.MaxLoad))
	}

	if b.Options.CriticalLoad != "" {
		parts = append(parts, fmt.Sprintf("--critical-load=%s", b.Options.CriticalLoad))
	}

	if b.Options.CheckInterval > 0 {
		parts = append(parts, fmt.Sprintf("--check-interval=%d", b.Options.CheckInterval))
	}

	if b.Options.MaxLag > 0 {
		parts = append(parts, fmt.Sprintf("--max-lag=%d", b.Options.MaxLag))
	}

	if b.Options.Charset != "" {
		parts = append(parts, fmt.Sprintf("--charset=%s", b.Options.Charset))
	}

	if b.Options.Progress != "" {
		parts = append(parts, fmt.Sprintf("--progress=%s", b.Options.Progress))
	}

	if b.Options.SetVars != "" {
		parts = append(parts, fmt.Sprintf("--set-vars=%s", b.Options.SetVars))
	}

	if b.Options.BinlogPosition != "" {
		parts = append(parts, fmt.Sprintf("--check-slave-lag=%s", b.Options.BinlogPosition))
	}

	// 执行选项
	if b.Options.Print {
		parts = append(parts, "--print")
	}

	if b.Options.Execute && !b.Options.DryRun {
		parts = append(parts, "--execute")
	}

	if b.Options.DryRun {
		parts = append(parts, "--dry-run")
	}

	if b.Options.DropOldTable {
		parts = append(parts, "--drop-old-table")
	}

	if b.Options.Statistics {
		parts = append(parts, "--statistics")
	}

	// 拼接命令
	command := strings.Join(parts, " \\\n  ")

	return command, nil
}

// validateAndCleanAlterSQL 验证和清理ALTER语句
func (b *PTCommandBuilder) validateAndCleanAlterSQL(alterSQL string) (string, error) {
	// 移除前后空白字符
	sql := strings.TrimSpace(alterSQL)

	// 移除可能的分号
	sql = strings.TrimSuffix(sql, ";")

	// 检查是否以ALTER TABLE开头，如果是则移除
	if strings.HasPrefix(strings.ToUpper(sql), "ALTER TABLE") {
		// 找到表名后的部分
		re := regexp.MustCompile(`(?i)^\s*ALTER\s+TABLE\s+(\S+)\s+(.+)$`)
		matches := re.FindStringSubmatch(sql)
		if len(matches) >= 3 {
			sql = matches[2]
		}
	}

	// 检查是否包含危险操作
	if err := b.checkDangerousOperations(sql); err != nil {
		return "", err
	}

	// 验证ALTER语句语法
	if err := b.validateAlterSyntax(sql); err != nil {
		return "", err
	}

	return sql, nil
}

// checkDangerousOperations 检查危险操作
func (b *PTCommandBuilder) checkDangerousOperations(sql string) error {
	upperSQL := strings.ToUpper(sql)

	// 检查DROP TABLE操作
	if strings.Contains(upperSQL, "DROP TABLE") {
		return fmt.Errorf("不允许使用DROP TABLE操作")
	}

	// 检查TRUNCATE操作
	if strings.Contains(upperSQL, "TRUNCATE") {
		return fmt.Errorf("不允许使用TRUNCATE操作")
	}

	// 检查DELETE操作
	if strings.Contains(upperSQL, "DELETE") {
		return fmt.Errorf("不允许使用DELETE操作")
	}

	return nil
}

// validateAlterSyntax 验证ALTER语句语法
func (b *PTCommandBuilder) validateAlterSyntax(sql string) error {
	upperSQL := strings.ToUpper(sql)

	// 检查基本的ALTER语法
	validOperations := []string{
		"ADD COLUMN", "ADD INDEX", "ADD KEY", "ADD UNIQUE",
		"DROP COLUMN", "DROP INDEX", "DROP KEY",
		"MODIFY COLUMN", "CHANGE COLUMN", "ALTER COLUMN",
		"ENGINE=", "AUTO_INCREMENT=", "COMMENT=",
		"ADD CONSTRAINT", "DROP CONSTRAINT",
		"ADD PRIMARY KEY", "DROP PRIMARY KEY",
	}

	hasValidOperation := false
	for _, op := range validOperations {
		if strings.Contains(upperSQL, op) {
			hasValidOperation = true
			break
		}
	}

	if !hasValidOperation {
		return fmt.Errorf("不支持的ALTER操作类型")
	}

	return nil
}

// AnalyzeDDLRisk 分析DDL操作风险
func (b *PTCommandBuilder) AnalyzeDDLRisk() map[string]interface{} {
	risk := map[string]interface{}{
		"level":          "low",
		"warnings":       []string{},
		"suggestions":    []string{},
		"estimated_time": "unknown",
	}

	upperSQL := strings.ToUpper(b.AlterStatement)

	// 检查高风险操作
	if strings.Contains(upperSQL, "DROP COLUMN") {
		risk["level"] = "high"
		risk["warnings"] = append(risk["warnings"].([]string), "删除列操作不可逆，请确保数据已备份")
	}

	if strings.Contains(upperSQL, "DROP INDEX") || strings.Contains(upperSQL, "DROP KEY") {
		risk["level"] = "medium"
		risk["warnings"] = append(risk["warnings"].([]string), "删除索引可能影响查询性能")
	}

	// 根据表大小评估时间
	if b.TableInfo.Rows > 0 {
		estimatedTime := b.estimateExecutionTime()
		risk["estimated_time"] = estimatedTime

		if b.TableInfo.Rows > 10000000 { // 超过1000万行
			risk["level"] = "high"
			risk["warnings"] = append(risk["warnings"].([]string), "大表操作，执行时间较长，建议在低峰期执行")
		} else if b.TableInfo.Rows > 1000000 { // 超过100万行
			if risk["level"] == "low" {
				risk["level"] = "medium"
			}
			risk["suggestions"] = append(risk["suggestions"].([]string), "建议适当调整chunk-size参数")
		}
	}

	// 生产环境警告
	if b.ConnectionConfig.Host != "" && strings.Contains(strings.ToLower(b.ConnectionConfig.Host), "prod") {
		risk["warnings"] = append(risk["warnings"].([]string), "生产环境操作，请谨慎执行")
	}

	return risk
}

// estimateExecutionTime 估算执行时间
func (b *PTCommandBuilder) estimateExecutionTime() string {
	if b.TableInfo.Rows <= 0 {
		return "unknown"
	}

	// 简单估算：假设每秒处理1000行
	seconds := b.TableInfo.Rows / 1000

	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%d分钟", seconds/60)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
}

// GetRecommendedChunkSize 获取推荐的chunk-size
func (b *PTCommandBuilder) GetRecommendedChunkSize() int {
	if b.TableInfo.Rows <= 0 {
		return 1000 // 默认值
	}

	// 根据表大小推荐chunk-size
	if b.TableInfo.Rows < 100000 { // 10万行以下
		return 1000
	} else if b.TableInfo.Rows < 1000000 { // 100万行以下
		return 2000
	} else if b.TableInfo.Rows < 10000000 { // 1000万行以下
		return 5000
	} else { // 1000万行以上
		return 8000
	}
}

// PreviewCommand 预览命令（隐藏密码）
func (b *PTCommandBuilder) PreviewCommand() (string, error) {
	// 临时保存原密码
	originalPassword := b.ConnectionConfig.Password

	// 用星号替换密码
	b.ConnectionConfig.Password = "***"

	// 构建命令
	command, err := b.buildCommand()

	// 恢复原密码
	b.ConnectionConfig.Password = originalPassword

	return command, err
}
