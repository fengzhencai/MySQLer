package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql" // MySQL驱动
)

// DatabaseConnection 数据库连接配置
type DatabaseConnection struct {
	Host           string
	Port           int
	Username       string
	Password       string
	DatabaseName   string
	ConnectTimeout int
	Charset        string
	UseSSL         bool
}

// TestConnection 测试数据库连接
func TestConnection(conn *DatabaseConnection) error {
	if conn == nil {
		return fmt.Errorf("连接配置不能为空")
	}

	// 构建DSN
	dsn := buildDSN(conn)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("创建数据库连接失败: %v", err)
	}
	defer db.Close()

	// 设置连接超时
	if conn.ConnectTimeout > 0 {
		db.SetConnMaxLifetime(time.Duration(conn.ConnectTimeout) * time.Second)
	}

	// 测试连接
	ctx, cancel := createTimeoutContext(conn.ConnectTimeout)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	return nil
}

// GetDatabaseInfo 获取数据库基本信息
func GetDatabaseInfo(conn *DatabaseConnection) (map[string]interface{}, error) {
	dsn := buildDSN(conn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %v", err)
	}
	defer db.Close()

	ctx, cancel := createTimeoutContext(conn.ConnectTimeout)
	defer cancel()

	// 获取MySQL版本
	var version string
	err = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("获取数据库版本失败: %v", err)
	}

	// 获取当前数据库名
	var currentDB string
	err = db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&currentDB)
	if err != nil {
		return nil, fmt.Errorf("获取当前数据库名失败: %v", err)
	}

	// 获取字符集
	var charset string
	err = db.QueryRowContext(ctx, "SELECT @@character_set_database").Scan(&charset)
	if err != nil {
		charset = "unknown"
	}

	return map[string]interface{}{
		"version":     version,
		"database":    currentDB,
		"charset":     charset,
		"server_time": time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// GetDatabaseList 获取数据库列表
func GetDatabaseList(conn *DatabaseConnection) ([]string, error) {
	dsn := buildDSN(conn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %v", err)
	}
	defer db.Close()

	ctx, cancel := createTimeoutContext(conn.ConnectTimeout)
	defer cancel()

	// 查询数据库列表（排除系统数据库）
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
		ORDER BY schema_name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询数据库列表失败: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		databases = append(databases, dbName)
	}

	return databases, nil
}

// GetTableList 获取指定数据库的表列表
func GetTableList(conn *DatabaseConnection, database string) ([]map[string]interface{}, error) {
	dsn := buildDSN(conn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %v", err)
	}
	defer db.Close()

	ctx, cancel := createTimeoutContext(conn.ConnectTimeout)
	defer cancel()

	// 查询表信息
	query := `
		SELECT 
			table_name,
			table_rows,
			data_length,
			index_length,
			table_comment,
			engine,
			create_time
		FROM information_schema.tables 
		WHERE table_schema = ? AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := db.QueryContext(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %v", err)
	}
	defer rows.Close()

	var tables []map[string]interface{}
	for rows.Next() {
		var tableName, tableComment, engine sql.NullString
		var tableRows, dataLength, indexLength sql.NullInt64
		var createTime sql.NullTime

		err := rows.Scan(&tableName, &tableRows, &dataLength, &indexLength,
			&tableComment, &engine, &createTime)
		if err != nil {
			continue
		}

		table := map[string]interface{}{
			"table_name":    tableName.String,
			"table_rows":    tableRows.Int64,
			"data_length":   dataLength.Int64,
			"index_length":  indexLength.Int64,
			"table_comment": tableComment.String,
			"engine":        engine.String,
		}

		if createTime.Valid {
			table["create_time"] = createTime.Time.Format("2006-01-02 15:04:05")
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// GetTableSchema 获取指定表的列与索引结构
func GetTableSchema(conn *DatabaseConnection, database string, table string) (map[string]interface{}, error) {
	dsn := buildDSN(conn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %v", err)
	}
	defer db.Close()

	ctx, cancel := createTimeoutContext(conn.ConnectTimeout)
	defer cancel()

	// 列信息
	colQuery := `
        SELECT 
            column_name,
            ordinal_position,
            column_type,
            is_nullable,
            column_default,
            column_key,
            extra,
            column_comment
        FROM information_schema.columns
        WHERE table_schema = ? AND table_name = ?
        ORDER BY ordinal_position
    `

	colRows, err := db.QueryContext(ctx, colQuery, database, table)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败: %v", err)
	}
	defer colRows.Close()

	var columns []map[string]interface{}
	for colRows.Next() {
		var (
			columnName    sql.NullString
			ordinal       sql.NullInt64
			columnType    sql.NullString
			isNullable    sql.NullString
			columnDefault sql.NullString
			columnKey     sql.NullString
			extra         sql.NullString
			columnComment sql.NullString
		)

		if err := colRows.Scan(&columnName, &ordinal, &columnType, &isNullable, &columnDefault, &columnKey, &extra, &columnComment); err != nil {
			continue
		}

		columns = append(columns, map[string]interface{}{
			"column_name":      columnName.String,
			"ordinal_position": ordinal.Int64,
			"column_type":      columnType.String,
			"is_nullable":      isNullable.String,
			"column_default":   columnDefault.String,
			"column_key":       columnKey.String,
			"extra":            extra.String,
			"column_comment":   columnComment.String,
		})
	}

	// 索引信息
	idxQuery := `
        SELECT 
            index_name,
            non_unique,
            column_name,
            seq_in_index
        FROM information_schema.statistics
        WHERE table_schema = ? AND table_name = ?
        ORDER BY index_name, seq_in_index
    `

	idxRows, err := db.QueryContext(ctx, idxQuery, database, table)
	if err != nil {
		return nil, fmt.Errorf("查询索引信息失败: %v", err)
	}
	defer idxRows.Close()

	var indexes []map[string]interface{}
	for idxRows.Next() {
		var (
			indexName  sql.NullString
			nonUnique  sql.NullInt64
			columnName sql.NullString
			seqInIndex sql.NullInt64
		)

		if err := idxRows.Scan(&indexName, &nonUnique, &columnName, &seqInIndex); err != nil {
			continue
		}

		indexes = append(indexes, map[string]interface{}{
			"index_name":   indexName.String,
			"non_unique":   nonUnique.Int64,
			"column_name":  columnName.String,
			"seq_in_index": seqInIndex.Int64,
		})
	}

	return map[string]interface{}{
		"columns": columns,
		"indexes": indexes,
	}, nil
}

// buildDSN 构建MySQL DSN
func buildDSN(conn *DatabaseConnection) string {
	config := mysql.Config{
		User:   conn.Username,
		Passwd: conn.Password,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%d", conn.Host, conn.Port),
		DBName: conn.DatabaseName,
		Params: map[string]string{
			"charset":   conn.Charset,
			"parseTime": "true",
			"loc":       "Local",
		},
		Timeout:              time.Duration(conn.ConnectTimeout) * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		AllowNativePasswords: true,
	}

	if conn.UseSSL {
		config.TLSConfig = "true"
	}

	return config.FormatDSN()
}

// createTimeoutContext 创建带超时的context
func createTimeoutContext(timeoutSeconds int) (context.Context, context.CancelFunc) {
	timeout := 10 * time.Second // 默认10秒
	if timeoutSeconds > 0 {
		timeout = time.Duration(timeoutSeconds) * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}
