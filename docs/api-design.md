# MySQLer API接口设计文档

## 1. API设计规范

### 1.1 RESTful设计原则
- 使用HTTP动词表示操作类型（GET, POST, PUT, DELETE）
- URL路径表示资源，避免动词
- 使用复数形式表示资源集合
- 使用HTTP状态码表示操作结果

### 1.2 统一响应格式

#### 成功响应格式
```json
{
    "code": 200,
    "message": "success",
    "data": {
        // 具体数据内容
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-123456789"
}
```

#### 错误响应格式
```json
{
    "code": 400,
    "message": "参数验证失败",
    "error": {
        "type": "ValidationError",
        "details": [
            {
                "field": "host",
                "message": "主机地址不能为空"
            }
        ]
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-123456789"
}
```

### 1.3 状态码定义

| HTTP状态码 | 业务码 | 说明 | 示例场景 |
|-----------|--------|------|----------|
| 200 | 200 | 操作成功 | 查询、更新成功 |
| 201 | 201 | 创建成功 | 新建连接、提交执行任务 |
| 400 | 400 | 请求参数错误 | 参数格式错误、必填参数缺失 |
| 401 | 401 | 未认证 | Token无效、未登录 |
| 403 | 403 | 权限不足 | 无操作权限 |
| 404 | 404 | 资源不存在 | 连接ID不存在 |
| 409 | 409 | 资源冲突 | 连接名称重复 |
| 422 | 422 | 业务逻辑错误 | 连接测试失败 |
| 500 | 500 | 服务器内部错误 | 系统异常 |
| 503 | 503 | 服务不可用 | Docker服务异常 |

### 1.4 自定义业务错误码

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 1001 | 数据库连接失败 | 检查连接配置 |
| 1002 | DDL语法错误 | 修改DDL语句 |
| 1003 | 执行超时 | 调整超时配置或分批执行 |
| 1004 | Docker容器启动失败 | 检查Docker服务状态 |
| 1005 | 目标表不存在 | 确认表名是否正确 |
| 1006 | 权限不足 | 检查数据库用户权限 |
| 1007 | 表正在执行DDL | 等待当前操作完成 |
| 1008 | 存储空间不足 | 清理磁盘空间 |

## 2. 认证授权API

### 2.1 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "username": "admin",
    "password": "password123"
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "登录成功",
    "data": {
        "user": {
            "id": "user-123",
            "username": "admin",
            "display_name": "管理员",
            "role": "admin",
            "permissions": ["execute_ddl", "manage_connections", "view_history"]
        },
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_at": "2024-01-16T10:30:00Z"
    }
}
```

### 2.2 用户登出
```http
POST /api/v1/auth/logout
Authorization: Bearer {token}
```

### 2.3 获取用户信息
```http
GET /api/v1/auth/profile
Authorization: Bearer {token}
```

### 2.4 刷新Token
```http
POST /api/v1/auth/refresh
Authorization: Bearer {token}
```

## 3. 连接管理API

### 3.1 获取连接列表
```http
GET /api/v1/connections?environment=prod&page=1&size=20
Authorization: Bearer {token}
```

**查询参数**:
- `environment`: 环境筛选 (prod/test/dev)
- `keyword`: 关键词搜索
- `page`: 页码 (默认1)
- `size`: 每页数量 (默认20)

**响应示例**:
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "list": [
            {
                "id": "conn-123",
                "name": "生产环境-报表库",
                "environment": "prod",
                "host": "prod-db.example.com",
                "port": 3306,
                "database_name": "report_db",
                "description": "生产环境报表数据库",
                "created_at": "2024-01-15T10:30:00Z",
                "created_by": "admin",
                "status": "active"
            }
        ],
        "pagination": {
            "page": 1,
            "size": 20,
            "total": 5,
            "pages": 1
        }
    }
}
```

### 3.2 创建连接
```http
POST /api/v1/connections
Authorization: Bearer {token}
Content-Type: application/json

{
    "name": "测试环境-用户库",
    "environment": "test",
    "host": "test-db.example.com",
    "port": 3306,
    "username": "test_user",
    "password": "test_password",
    "database_name": "user_db",
    "description": "测试环境用户数据库",
    "connect_timeout": 10,
    "charset": "utf8mb4",
    "use_ssl": false
}
```

### 3.3 更新连接
```http
PUT /api/v1/connections/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
    "name": "更新后的连接名称",
    "description": "更新描述信息"
}
```

### 3.4 删除连接
```http
DELETE /api/v1/connections/{id}
Authorization: Bearer {token}
```

### 3.5 测试连接
```http
POST /api/v1/connections/{id}/test
Authorization: Bearer {token}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "连接测试成功",
    "data": {
        "status": "success",
        "response_time": 150,
        "server_version": "8.0.32",
        "charset": "utf8mb4"
    }
}
```

### 3.6 获取数据库列表
```http
GET /api/v1/connections/{id}/databases
Authorization: Bearer {token}
```

### 3.7 获取表列表
```http
GET /api/v1/connections/{id}/databases/{database}/tables?keyword=user
Authorization: Bearer {token}
```

## 4. DDL执行API

### 4.1 提交执行任务
```http
POST /api/v1/executions
Authorization: Bearer {token}
Content-Type: application/json

{
    "connection_id": "conn-123",
    "database_name": "test_db",
    "table_name": "users",
    "ddl_type": "add_column",
    "ddl_statement": "ADD COLUMN phone VARCHAR(20) NULL COMMENT '手机号' AFTER email",
    "execution_params": {
        "chunk_size": 8000,
        "max_load": "Threads_connected:1000",
        "critical_load": "Threads_running:500",
        "charset": "utf8mb4",
        "lock_wait_timeout": 100,
        "dry_run": false
    },
    "description": "用户表添加手机号字段"
}
```

**响应示例**:
```json
{
    "code": 201,
    "message": "执行任务创建成功",
    "data": {
        "execution_id": "exec-456",
        "status": "pending",
        "generated_command": "pt-online-schema-change --user=*** --password=*** --host=test-db.example.com --alter \"ADD COLUMN phone VARCHAR(20) NULL COMMENT '手机号' AFTER email\" D=test_db,t=users --chunk-size=8000 --print --execute --statistics",
        "created_at": "2024-01-15T10:30:00Z",
        "estimated_duration": "5-10分钟"
    }
}
```

### 4.2 获取执行历史
```http
GET /api/v1/executions?status=completed&database=test_db&start_date=2024-01-01&end_date=2024-01-15&page=1&size=20
Authorization: Bearer {token}
```

**查询参数**:
- `status`: 执行状态筛选
- `database`: 数据库筛选
- `table`: 表名筛选
- `ddl_type`: DDL类型筛选
- `start_date`: 开始日期
- `end_date`: 结束日期
- `created_by`: 执行人筛选
- `keyword`: 关键词搜索

### 4.3 获取执行详情
```http
GET /api/v1/executions/{id}
Authorization: Bearer {token}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "exec-456",
        "connection": {
            "id": "conn-123",
            "name": "测试环境-用户库",
            "host": "test-db.example.com"
        },
        "database_name": "test_db",
        "table_name": "users",
        "ddl_type": "add_column",
        "ddl_statement": "ADD COLUMN phone VARCHAR(20) NULL COMMENT '手机号' AFTER email",
        "status": "completed",
        "start_time": "2024-01-15T10:30:00Z",
        "end_time": "2024-01-15T10:38:32Z",
        "duration_seconds": 512,
        "processed_rows": 1500000,
        "total_rows": 1500000,
        "avg_speed": 2929.69,
        "generated_command": "pt-online-schema-change ...",
        "execution_params": {
            "chunk_size": 8000,
            "max_load": "Threads_connected:1000"
        },
        "container_id": "container-789",
        "created_by": "admin",
        "description": "用户表添加手机号字段"
    }
}
```

### 4.4 停止执行
```http
POST /api/v1/executions/{id}/stop
Authorization: Bearer {token}
```

### 4.5 重新执行
```http
POST /api/v1/executions/{id}/retry
Authorization: Bearer {token}
Content-Type: application/json

{
    "execution_params": {
        "chunk_size": 5000
    }
}
```

### 4.6 获取执行日志
```http
GET /api/v1/executions/{id}/logs?offset=0&limit=1000&level=info
Authorization: Bearer {token}
```

**查询参数**:
- `offset`: 日志偏移量
- `limit`: 返回条数
- `level`: 日志级别筛选
- `keyword`: 关键词搜索

## 5. 工具辅助API

### 5.1 DDL语句解析
```http
POST /api/v1/tools/ddl/parse
Authorization: Bearer {token}
Content-Type: application/json

{
    "ddl_statement": "ADD COLUMN phone VARCHAR(20) NULL COMMENT '手机号' AFTER email"
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "DDL解析成功",
    "data": {
        "ddl_type": "add_column",
        "is_valid": true,
        "parsed_info": {
            "operation": "ADD COLUMN",
            "column_name": "phone",
            "data_type": "VARCHAR(20)",
            "nullable": true,
            "comment": "手机号",
            "position": "AFTER email"
        },
        "warnings": [],
        "suggestions": [
            "建议为新增列设置默认值"
        ]
    }
}
```

### 5.2 命令预览
```http
POST /api/v1/tools/ddl/preview
Authorization: Bearer {token}
Content-Type: application/json

{
    "connection_id": "conn-123",
    "database_name": "test_db",
    "table_name": "users",
    "ddl_statement": "ADD COLUMN phone VARCHAR(20) NULL",
    "execution_params": {
        "chunk_size": 8000,
        "dry_run": true
    }
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "命令生成成功",
    "data": {
        "generated_command": "pt-online-schema-change --user=*** --password=*** --host=test-db.example.com --alter \"ADD COLUMN phone VARCHAR(20) NULL\" D=test_db,t=users --chunk-size=8000 --dry-run --print",
        "estimated_info": {
            "table_size": "2.5GB",
            "row_count": 1500000,
            "estimated_duration": "8-12分钟",
            "disk_usage": "需要额外5GB磁盘空间"
        },
        "risk_assessment": {
            "level": "low",
            "warnings": [
                "表数据量较大，建议在业务低峰期执行"
            ]
        }
    }
}
```

### 5.3 获取表结构信息
```http
GET /api/v1/tools/tables/{connection_id}/{database}/{table}/schema
Authorization: Bearer {token}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "table_name": "users",
        "engine": "InnoDB",
        "charset": "utf8mb4",
        "collation": "utf8mb4_unicode_ci",
        "row_count": 1500000,
        "data_length": 2684354560,
        "index_length": 524288000,
        "auto_increment": 1500001,
        "columns": [
            {
                "name": "id",
                "type": "bigint",
                "nullable": false,
                "key": "PRI",
                "default": null,
                "extra": "auto_increment",
                "comment": "主键ID"
            },
            {
                "name": "username",
                "type": "varchar(50)",
                "nullable": false,
                "key": "UNI",
                "default": null,
                "extra": "",
                "comment": "用户名"
            }
        ],
        "indexes": [
            {
                "name": "PRIMARY",
                "type": "PRIMARY",
                "columns": ["id"]
            },
            {
                "name": "uk_username",
                "type": "UNIQUE",
                "columns": ["username"]
            }
        ]
    }
}
```

## 6. 系统管理API

### 6.1 获取系统状态
```http
GET /api/v1/system/status
Authorization: Bearer {token}
```

### 6.2 获取执行统计
```http
GET /api/v1/system/statistics?period=7d
Authorization: Bearer {token}
```

### 6.3 获取用户列表（管理员）
```http
GET /api/v1/system/users
Authorization: Bearer {token}
```

### 6.4 创建用户（管理员）
```http
POST /api/v1/system/users
Authorization: Bearer {token}
Content-Type: application/json

{
    "username": "operator1",
    "password": "password123",
    "display_name": "操作员1",
    "email": "operator1@example.com",
    "role": "operator"
}
```

## 7. WebSocket实时通信

### 7.1 连接建立
```javascript
const ws = new WebSocket('wss://example.com/ws/execution/{execution_id}?token={jwt_token}');
```

### 7.2 消息格式

#### 执行进度推送
```json
{
    "type": "progress",
    "data": {
        "execution_id": "exec-456",
        "status": "running",
        "progress": 75.5,
        "processed_rows": 1125000,
        "total_rows": 1500000,
        "current_speed": 3500,
        "current_stage": "copying_data",
        "estimated_remaining": "2分30秒"
    }
}
```

#### 日志推送
```json
{
    "type": "log",
    "data": {
        "execution_id": "exec-456",
        "timestamp": "2024-01-15T10:35:21Z",
        "level": "info",
        "message": "Copied 1125000/1500000 rows (75%)"
    }
}
```

#### 状态变更推送
```json
{
    "type": "status_change",
    "data": {
        "execution_id": "exec-456",
        "old_status": "running",
        "new_status": "completed",
        "timestamp": "2024-01-15T10:38:32Z"
    }
}
```

#### 错误推送
```json
{
    "type": "error",
    "data": {
        "execution_id": "exec-456",
        "error_code": 1002,
        "error_message": "DDL语法错误",
        "details": "Unknown column 'invalid_column' in 'field list'",
        "timestamp": "2024-01-15T10:32:15Z"
    }
}
```

## 8. API限流与安全

### 8.1 请求限流
- **全局限流**: 1000 requests/minute per IP
- **用户限流**: 100 requests/minute per user
- **执行限流**: 最多10个并发执行任务

### 8.2 安全头设置
```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

### 8.3 敏感信息处理
- 密码字段响应时使用 `***` 替换
- 日志中自动脱敏敏感信息
- API响应中不包含完整数据库连接信息

## 9. API版本控制

### 9.1 版本策略
- URL路径版本控制: `/api/v1/`, `/api/v2/`
- 向后兼容性保证
- 弃用API提前通知机制

### 9.2 版本变更记录
| 版本 | 发布时间 | 主要变更 |
|------|----------|----------|
| v1.0 | 2024-01-15 | 初始版本发布 |

## 10. 错误处理最佳实践

### 10.1 客户端错误处理
```javascript
// 统一错误处理示例
axios.interceptors.response.use(
    response => response,
    error => {
        const { code, message } = error.response.data;
        
        switch (code) {
            case 401:
                // 跳转登录页
                break;
            case 403:
                // 显示权限不足提示
                break;
            default:
                // 显示通用错误提示
                showErrorMessage(message);
        }
        
        return Promise.reject(error);
    }
);
```

### 10.2 服务端错误日志
```go
// 错误日志记录示例
logger.WithFields(logrus.Fields{
    "request_id": requestID,
    "user_id":    userID,
    "method":     r.Method,
    "path":       r.URL.Path,
    "error":      err.Error(),
}).Error("API request failed")
```

---

**文档版本**: v1.0  
**创建时间**: 2024-01-15  
**维护人员**: 后端开发团队