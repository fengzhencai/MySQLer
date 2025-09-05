# PT-Online-Schema-Change Web管理平台需求文档

## 1. 项目概述

### 1.1 项目背景
为了方便运维人员和开发人员在线执行MySQL数据库的DDL操作，需要开发一个基于Web的pt-online-schema-change工具管理平台。该平台提供可视化界面来执行在线表结构变更，避免传统DDL操作的锁表风险，确保生产环境数据库的稳定性。

### 1.2 项目目标
- **主要目标**：提供友好的Web界面执行pt-online-schema-change命令
- **核心价值**：
  - 降低DDL操作的技术门槛，运维人员无需记忆复杂命令
  - 支持实时查看执行进度和详细日志
  - 提供预设的常用操作模板（如碎片整理）
  - 确保操作安全性和可追溯性
  - 统一管理多个数据库连接信息

### 1.3 目标用户
- **主要用户**：数据库运维工程师
- **次要用户**：后端开发工程师、DBA
- **使用场景**：生产环境在线表结构变更、数据库性能优化

## 2. 技术栈选择

### 2.1 后端技术栈
- **编程语言**: Go 1.21+
- **Web框架**: Gin v1.9+ (最流行且高性能的Go Web框架)
- **数据库**: MySQL 8.0+ (存储执行历史和配置)
- **容器化**: Docker + Docker Compose
- **实时通信**: gorilla/websocket (执行进度实时推送)
- **配置管理**: Viper v1.15+
- **日志框架**: logrus v1.9+
- **参数验证**: go-playground/validator v10.11+
- **数据库ORM**: GORM v1.25+
- **加密**: bcrypt (密码加密)

### 2.2 前端技术栈
- **框架**: Vue 3.3+ + TypeScript 5.0+
- **UI组件库**: Element Plus 2.3+ (功能丰富，文档完善)
- **构建工具**: Vite 4.0+ (快速构建)
- **状态管理**: Pinia 2.1+ (Vue 3推荐状态管理)
- **HTTP客户端**: Axios 1.4+
- **实时通信**: Socket.IO Client 4.7+
- **代码规范**: ESLint + Prettier
- **CSS预处理**: SCSS

### 2.3 基础设施
- **容器运行时**: Docker Engine 20.0+
- **编排工具**: Docker Compose 2.0+
- **基础镜像**: `registry-harbor.yafex.cn/base/percona-server:5.7v6`
- **反向代理**: Nginx (生产环境)

## 3. 核心功能需求

### 3.1 连接管理模块

#### 3.1.1 功能描述
管理和维护数据库连接信息，支持多环境、多数据库的连接配置。

#### 3.1.2 输入字段设计
```
连接配置表单：
┌─────────────────────────────────────────────┐
│ 连接名称: [输入框] (必填, 50字符以内)           │
│ 环境标识: [下拉] 生产/测试/开发                │
│ 主机地址: [输入框] (必填, IP或域名)            │
│ 端口号:   [输入框] (默认3306, 1-65535)        │
│ 用户名:   [输入框] (必填)                     │
│ 密码:     [密码框] (必填)                     │
│ 数据库:   [输入框] (必填)                     │
│ 描述信息: [文本域] (可选, 200字符以内)         │
│                                             │
│ 高级配置: [展开面板]                         │
│ ├─ 连接超时: [5] 秒                          │
│ ├─ 字符集: [utf8mb4]                        │
│ └─ SSL配置: ☐ 启用SSL连接                    │
│                                             │
│ [测试连接] [保存] [取消]                     │
└─────────────────────────────────────────────┘
```

#### 3.1.3 功能要求
- **连接管理**：支持增删改查操作
- **连接测试**：保存前必须测试连接成功
- **安全存储**：密码使用AES加密存储
- **分类管理**：按环境（生产/测试/开发）分类显示
- **导入导出**：支持配置的批量导入和导出
- **权限控制**：敏感连接信息访问权限控制

### 3.2 DDL操作执行模块

#### 3.2.1 功能描述
这是系统的核心模块，负责构建和执行pt-online-schema-change命令。

#### 3.2.2 操作类型设计

##### A. 碎片整理操作
- **触发方式**：一键式操作
- **命令生成**：自动生成 `--alter "ENGINE=INNODB"` 参数
- **应用场景**：定期表维护，优化表存储结构
- **参数配置**：支持自定义chunk-size和负载控制参数

##### B. 自定义DDL操作
支持的DDL类型：
```sql
-- 列操作
ADD COLUMN column_name data_type [NULL|NOT NULL] [DEFAULT value] [COMMENT 'comment'] [AFTER existing_column]
MODIFY COLUMN column_name new_data_type [NULL|NOT NULL] [DEFAULT value] [COMMENT 'comment']
DROP COLUMN column_name
CHANGE COLUMN old_name new_name data_type [attributes]

-- 索引操作  
ADD INDEX index_name (column_list)
ADD UNIQUE INDEX index_name (column_list)
DROP INDEX index_name

-- 约束操作
ADD CONSTRAINT constraint_name PRIMARY KEY (column_list)
ADD CONSTRAINT constraint_name FOREIGN KEY (column_list) REFERENCES table(column_list)
DROP CONSTRAINT constraint_name

-- 表属性修改
ENGINE=InnoDB
AUTO_INCREMENT=value
COMMENT='table comment'
```

#### 3.2.3 操作界面设计
```
DDL执行器界面：
┌─────────────────────────────────────────────┐
│ 📊 DDL执行器                                │
├─────────────────────────────────────────────┤
│ 连接信息:                                   │
│ [下拉选择器: 已保存的连接] [+新建连接]       │
│                                             │
│ 目标表信息:                                 │
│ 数据库: [输入框或下拉]                       │
│ 表名:   [输入框] [📋获取表列表]              │
│                                             │
│ 操作类型:                                   │
│ ○ 碎片整理 (一键优化表结构)                  │
│ ○ 自定义DDL                                │
│                                             │
│ [仅自定义DDL时显示]                         │
│ DDL语句编辑器:                              │
│ ┌─────────────────────────────────────────┐ │
│ │ ADD COLUMN tax_type int(11) NULL        │ │
│ │ COMMENT '1不含税报价，2含税报价'         │ │
│ │ AFTER manufacture_id;                   │ │
│ │                                         │ │
│ │ [语法高亮] [格式化] [验证语法]            │ │
│ └─────────────────────────────────────────┘ │
│                                             │
│ ⚙️ 执行参数配置: [展开/收起]                 │
│ ├─ Chunk Size: [8000] (建议1000-10000)     │
│ ├─ Max Load: [Threads_connected:12000]     │
│ ├─ Critical Load: [Threads_running=12000]  │
│ ├─ Charset: [utf8] [utf8mb4]               │
│ ├─ Lock Wait Timeout: [100]秒              │
│ └─ 其他参数: [--set-vars="..."]            │
│                                             │
│ 🔍 命令预览:                                │
│ ┌─────────────────────────────────────────┐ │
│ │ pt-online-schema-change \               │ │
│ │   --user=*** --password=*** \          │ │
│ │   --host=example.com \                  │ │
│ │   --alter "ADD COLUMN..." \             │ │
│ │   D=database,t=table \                  │ │
│ │   --print --execute --statistics        │ │
│ └─────────────────────────────────────────┘ │
│                                             │
│ [🔗测试连接] [👁预览命令] [▶️执行]           │
└─────────────────────────────────────────────┘
```

#### 3.2.4 安全确认机制
- **危险操作识别**：DROP操作、大表操作自动标记
- **二次确认**：危险操作需要输入确认文本
- **操作权限**：不同用户权限控制
- **操作记录**：所有操作自动记录审计日志

### 3.3 实时监控模块

#### 3.3.1 功能描述
提供pt-online-schema-change执行过程的实时监控和进度跟踪。

#### 3.3.2 监控界面设计
```
执行监控面板：
┌─────────────────────────────────────────────┐
│ 📈 执行状态监控                              │
├─────────────────────────────────────────────┤
│ 基本信息:                                   │
│ ├─ 执行ID: #20240115-001                    │
│ ├─ 目标表: yafex-report.oms_orders         │
│ ├─ 操作类型: 碎片整理                       │
│ ├─ 开始时间: 2024-01-15 10:30:00           │
│ ├─ 执行时长: 00:05:30                      │
│ └─ 当前状态: 🟢 执行中                      │
│                                             │
│ 进度信息:                                   │
│ 总体进度: [████████████░░░░] 75% 完成       │
│ ├─ 已处理行数: 3,750,000 / 5,000,000       │
│ ├─ 处理速度: 5,420 rows/sec                │
│ ├─ 剩余时间: 约 00:03:45                   │
│ └─ 当前阶段: 数据复制中...                  │
│                                             │
│ 执行阶段:                                   │
│ ✅ 1. 创建触发器                            │
│ ✅ 2. 创建新表                              │
│ 🔄 3. 复制数据 (进行中)                     │
│ ⏳ 4. 交换表名                              │
│ ⏳ 5. 清理资源                              │
│                                             │
│ 实时日志: [自动滚动] [暂停] [下载]           │
│ ┌─────────────────────────────────────────┐ │
│ │ [10:30:05] Creating triggers...         │ │
│ │ [10:30:08] Successfully created 3 triggers │
│ │ [10:30:10] Creating new table...        │ │
│ │ [10:32:15] Copying approximately 5000000 rows │
│ │ [10:35:20] Copied 3750000/5000000 rows (75%) │
│ │ [10:35:21] Current copy rate: 5420 rows/sec │
│ │ ▼ 实时更新中...                         │ │
│ └─────────────────────────────────────────┘ │
│                                             │
│ 系统负载:                                   │
│ ├─ 连接数: 145/12000 (1.2%)                │
│ ├─ 运行线程: 8/12000 (0.07%)               │
│ └─ 负载状态: 🟢 正常                        │
│                                             │
│ 操作控制:                                   │
│ [⏸暂停] [⏹停止] [📄导出日志] [🔄刷新]      │
└─────────────────────────────────────────────┘
```

#### 3.3.3 监控功能要求
- **实时更新**：通过WebSocket推送执行状态
- **进度计算**：基于已处理行数计算准确进度
- **性能监控**：显示处理速度、系统负载等指标
- **日志流式显示**：实时显示执行日志，支持搜索和过滤
- **告警机制**：异常情况自动告警
- **操作控制**：支持暂停、停止、重启等操作

### 3.4 历史记录管理模块

#### 3.4.1 功能描述
记录和管理所有DDL执行历史，提供查询、分析和重执行功能。

#### 3.4.2 历史记录界面
```
执行历史管理：
┌─────────────────────────────────────────────┐
│ 📚 执行历史记录                              │
├─────────────────────────────────────────────┤
│ 筛选条件:                                   │
│ 时间范围: [2024-01-01] 至 [2024-01-15]     │
│ 执行状态: [全部▼] 数据库: [全部▼]           │
│ 操作类型: [全部▼] 关键词: [搜索框]          │
│                              [🔍搜索] [重置]│
│                                             │
│ 统计信息:                                   │
│ 总计: 1,247次 | 成功: 1,156次 | 失败: 91次  │
│ 成功率: 92.7% | 平均耗时: 8分32秒           │
│                                             │
│ 记录列表: [每页20条] [导出Excel]             │
│ ┌─────────────────────────────────────────┐ │
│ │ID    │时间     │数据库   │表名 │操作│状态│ │
│ ├─────┼────────┼────────┼────┼───┼───┤ │
│ │#1247│01-15 10:30│report  │orders│碎片│✅│ │
│ │#1246│01-15 09:15│listing │item  │列新增│✅│ │
│ │#1245│01-14 16:22│report  │stats │索引│❌│ │
│ │#1244│01-14 14:30│user    │profile│修改│✅│ │
│ │ ... │   ...     │  ...   │ ... │...│.. │ │
│ └─────┴────────┴────────┴────┴───┴───┘ │
│                                             │
│ 详细操作: [👁查看] [🔄重新执行] [📄报告]    │
└─────────────────────────────────────────────┘

执行详情弹窗：
┌─────────────────────────────────────────────┐
│ 📋 执行详情 #1247                           │
├─────────────────────────────────────────────┤
│ 基本信息:                                   │
│ ├─ 执行时间: 2024-01-15 10:30:00 - 10:38:32│
│ ├─ 总耗时: 8分32秒                          │
│ ├─ 操作人: admin                           │
│ ├─ 目标: report.oms_orders (碎片整理)       │
│ └─ 状态: ✅ 执行成功                        │
│                                             │
│ 执行命令:                                   │
│ ┌─────────────────────────────────────────┐ │
│ │ pt-online-schema-change \               │ │
│ │   --user=fengzhencai \                  │ │
│ │   --host=report.db.yafex.cn \           │ │
│ │   --alter "ENGINE=INNODB" \             │ │
│ │   D=yafex-report-form,t=oms_orders \    │ │
│ │   --print --execute --statistics \      │ │
│ │   --chunk-size=8000                     │ │
│ └─────────────────────────────────────────┘ │
│                                             │
│ 执行统计:                                   │
│ ├─ 处理行数: 4,582,334 rows                │
│ ├─ 平均速度: 8,934 rows/sec                │
│ ├─ 数据大小: 2.3 GB                        │
│ └─ 碎片回收: 456 MB                        │
│                                             │
│ 完整日志: [展开查看] [下载日志]              │
│                                             │
│ [🔄重新执行] [📋复制命令] [关闭]             │
└─────────────────────────────────────────────┘
```

#### 3.4.3 功能要求
- **分页查询**：支持大量历史记录的分页显示
- **多维筛选**：按时间、状态、数据库、操作类型等维度筛选
- **执行统计**：提供成功率、平均耗时等统计信息
- **详细记录**：保存完整的执行命令、日志和结果
- **重复执行**：支持历史命令的一键重复执行
- **报告导出**：支持执行报告的Excel导出

### 3.5 系统管理模块

#### 3.5.1 用户管理
- **用户认证**：用户名密码登录
- **权限控制**：
  - 管理员：所有权限
  - 操作员：执行DDL权限
  - 查看者：仅查看权限
- **操作审计**：记录用户操作日志

#### 3.5.2 系统配置
- **全局参数**：默认chunk-size、负载阈值等
- **Docker配置**：容器资源限制
- **告警配置**：执行失败通知设置
- **备份策略**：执行历史数据保留策略

## 4. 非功能性需求

### 4.1 性能要求
- **响应时间**：界面操作响应时间 < 500ms
- **并发支持**：支持最多10个并发DDL执行
- **大表处理**：支持千万级行数表的DDL操作
- **内存使用**：单个执行任务内存占用 < 512MB

### 4.2 安全要求
- **数据传输**：HTTPS加密传输
- **密码存储**：AES-256加密存储
- **会话管理**：JWT Token认证，24小时过期
- **操作审计**：完整的操作日志记录
- **访问控制**：IP白名单（可选）

### 4.3 可用性要求
- **系统可用性**：99.5%在线时间
- **容错能力**：Docker容器异常自动重启
- **数据备份**：每日自动备份执行历史
- **错误恢复**：执行失败自动回滚机制

### 4.4 兼容性要求
- **浏览器支持**：Chrome 90+, Firefox 88+, Safari 14+
- **移动端**：响应式设计，支持平板查看
- **MySQL版本**：支持MySQL 5.7+ / 8.0+
- **操作系统**：支持Linux、Windows、macOS

## 5. 数据设计

### 5.1 数据库表结构设计

#### 5.1.1 连接信息表 (connections)
```sql
CREATE TABLE connections (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '连接名称',
    environment ENUM('prod','test','dev') DEFAULT 'test' COMMENT '环境类型',
    host VARCHAR(255) NOT NULL COMMENT '主机地址',
    port INT DEFAULT 3306 COMMENT '端口号',
    username VARCHAR(100) NOT NULL COMMENT '用户名',
    password TEXT NOT NULL COMMENT '加密密码',
    database_name VARCHAR(100) NOT NULL COMMENT '数据库名',
    description TEXT COMMENT '描述信息',
    connect_timeout INT DEFAULT 5 COMMENT '连接超时(秒)',
    charset VARCHAR(20) DEFAULT 'utf8mb4' COMMENT '字符集',
    use_ssl BOOLEAN DEFAULT FALSE COMMENT '是否使用SSL',
    created_by VARCHAR(100) COMMENT '创建人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_environment (environment),
    INDEX idx_name (name)
) COMMENT='数据库连接配置表';
```

#### 5.1.2 执行记录表 (execution_records)
```sql
CREATE TABLE execution_records (
    id VARCHAR(36) PRIMARY KEY,
    connection_id VARCHAR(36) NOT NULL COMMENT '连接ID',
    table_name VARCHAR(200) NOT NULL COMMENT '目标表名',
    database_name VARCHAR(100) NOT NULL COMMENT '数据库名',
    ddl_type ENUM('fragment','add_column','modify_column','drop_column','add_index','drop_index','other') COMMENT 'DDL类型',
    original_ddl TEXT COMMENT '原始DDL语句',
    generated_command TEXT NOT NULL COMMENT '生成的pt命令',
    execution_params JSON COMMENT '执行参数配置',
    status ENUM('pending','running','completed','failed','cancelled') DEFAULT 'pending',
    start_time TIMESTAMP NULL COMMENT '开始时间',
    end_time TIMESTAMP NULL COMMENT '结束时间',
    duration_seconds INT COMMENT '执行耗时(秒)',
    processed_rows BIGINT DEFAULT 0 COMMENT '已处理行数',
    total_rows BIGINT DEFAULT 0 COMMENT '总行数',
    avg_speed DECIMAL(10,2) COMMENT '平均处理速度(rows/sec)',
    container_id VARCHAR(64) COMMENT 'Docker容器ID',
    execution_logs LONGTEXT COMMENT '执行日志',
    error_message TEXT COMMENT '错误信息',
    created_by VARCHAR(100) COMMENT '执行人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (connection_id) REFERENCES connections(id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_table (database_name, table_name),
    INDEX idx_created_by (created_by)
) COMMENT='DDL执行记录表';
```

#### 5.1.3 用户表 (users)
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    display_name VARCHAR(100) COMMENT '显示名称',
    email VARCHAR(255) COMMENT '邮箱',
    role ENUM('admin','operator','viewer') DEFAULT 'viewer' COMMENT '角色权限',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_role (role)
) COMMENT='用户表';
```

#### 5.1.4 操作审计表 (audit_logs)
```sql
CREATE TABLE audit_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) COMMENT '用户ID',
    username VARCHAR(50) COMMENT '用户名',
    action VARCHAR(100) NOT NULL COMMENT '操作动作',
    resource_type VARCHAR(50) COMMENT '资源类型',
    resource_id VARCHAR(36) COMMENT '资源ID',
    request_data JSON COMMENT '请求数据',
    response_data JSON COMMENT '响应数据',
    ip_address VARCHAR(45) COMMENT '客户端IP',
    user_agent TEXT COMMENT '用户代理',
    execution_time_ms INT COMMENT '执行时间(毫秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
) COMMENT='操作审计日志表';
```

### 5.2 数据字典

#### 5.2.1 DDL类型枚举
- `fragment`: 碎片整理
- `add_column`: 添加列
- `modify_column`: 修改列
- `drop_column`: 删除列  
- `add_index`: 添加索引
- `drop_index`: 删除索引
- `other`: 其他类型

#### 5.2.2 执行状态枚举
- `pending`: 等待执行
- `running`: 执行中
- `completed`: 执行完成
- `failed`: 执行失败
- `cancelled`: 手动取消

## 6. 接口设计

### 6.1 RESTful API规范

#### 6.1.1 响应格式标准
```json
{
    "code": 200,
    "message": "success",
    "data": {},
    "timestamp": "2024-01-15T10:30:00Z"
}
```

#### 6.1.2 错误码定义
- `200`: 操作成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误
- `1001`: 数据库连接失败
- `1002`: DDL语法错误
- `1003`: 执行超时

### 6.2 主要API接口

#### 6.2.1 认证接口
```
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/profile
```

#### 6.2.2 连接管理接口
```
GET    /api/v1/connections              # 获取连接列表
POST   /api/v1/connections              # 创建连接
GET    /api/v1/connections/:id          # 获取连接详情
PUT    /api/v1/connections/:id          # 更新连接
DELETE /api/v1/connections/:id          # 删除连接
POST   /api/v1/connections/:id/test     # 测试连接
```

#### 6.2.3 DDL执行接口
```
POST   /api/v1/executions               # 提交执行任务
GET    /api/v1/executions               # 获取执行历史
GET    /api/v1/executions/:id           # 获取执行详情
POST   /api/v1/executions/:id/stop      # 停止执行
GET    /api/v1/executions/:id/logs      # 获取执行日志
POST   /api/v1/executions/:id/retry     # 重新执行
```

#### 6.2.4 工具接口
```
POST   /api/v1/tools/ddl/parse          # DDL语句解析
POST   /api/v1/tools/ddl/preview        # 预览生成命令
GET    /api/v1/tools/databases/:conn_id # 获取数据库列表
GET    /api/v1/tools/tables/:conn_id/:db # 获取表列表
```

### 6.3 WebSocket接口
```
/ws/execution/:id                       # 执行进度实时推送
```

WebSocket消息格式：
```json
{
    "type": "progress",
    "data": {
        "execution_id": "uuid",
        "status": "running",
        "progress": 75.5,
        "processed_rows": 3750000,
        "total_rows": 5000000,
        "current_speed": 5420,
        "current_stage": "copying_data",
        "log_line": "[10:35:21] Copied 3750000/5000000 rows (75%)"
    }
}
```

## 7. 部署架构

### 7.1 Docker Compose配置
```yaml
version: '3.8'
services:
  # Web应用服务
  pt-web:
    build: 
      context: .
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=mysql
      - DB_USER=pt_web
      - DB_PASSWORD=${DB_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - DOCKER_HOST=unix:///var/run/docker.sock
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - pt_logs:/app/logs
    depends_on:
      - mysql
      - redis
    restart: unless-stopped
    
  # MySQL数据库
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - MYSQL_DATABASE=pt_web
      - MYSQL_USER=pt_web
      - MYSQL_PASSWORD=${DB_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
      - ./sql/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    restart: unless-stopped
    
  # Redis缓存
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped
    
  # Nginx反向代理 (生产环境)
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - pt-web
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
  pt_logs:
```

### 7.2 容器资源限制
```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 2G
    reservations:
      cpus: '0.5'
      memory: 512M
```

## 8. 安全设计

### 8.1 认证与授权
- **JWT认证**：有状态Token，24小时过期
- **角色权限**：基于RBAC的权限控制
- **密码策略**：最少8位，包含数字和字母
- **会话管理**：支持强制登出和并发限制

### 8.2 数据安全
- **传输加密**：全站HTTPS，TLS 1.2+
- **存储加密**：敏感数据AES-256加密
- **SQL注入防护**：使用ORM和参数化查询
- **XSS防护**：输出转义和CSP策略

### 8.3 操作安全
- **危险操作确认**：DROP操作二次确认
- **操作审计**：完整的操作日志记录
- **IP白名单**：可配置访问IP限制
- **容器隔离**：执行容器网络隔离

## 9. 监控与运维

### 9.1 应用监控
- **健康检查**：HTTP健康检查端点
- **性能指标**：响应时间、并发数、错误率
- **业务指标**：执行成功率、平均耗时
- **资源监控**：CPU、内存、磁盘使用率

### 9.2 日志管理
- **结构化日志**：JSON格式统一输出
- **日志级别**：DEBUG、INFO、WARN、ERROR
- **日志轮转**：按大小和时间自动轮转
- **集中收集**：支持ELK或其他日志平台

### 9.3 告警机制
- **执行失败告警**：DDL执行失败自动通知
- **系统异常告警**：服务不可用、资源不足
- **通知方式**：邮件、企业微信、钉钉

## 10. 测试策略

### 10.1 测试范围
- **单元测试**：覆盖率 > 80%
- **集成测试**：API接口测试
- **端到端测试**：关键业务流程测试
- **性能测试**：并发和大数据量测试
- **安全测试**：漏洞扫描和渗透测试

### 10.2 测试环境
- **开发环境**：本地Docker环境
- **测试环境**：模拟生产环境配置
- **预发环境**：生产数据脱敏测试

## 11. 项目里程碑

### Phase 1: 基础功能 (3周)
- **Week 1**：项目架构搭建、数据库设计
- **Week 2**：连接管理、用户认证功能
- **Week 3**：基础DDL执行功能

### Phase 2: 核心功能 (3周)  
- **Week 4**：实时监控、进度展示
- **Week 5**：历史记录、日志管理
- **Week 6**：前端界面优化

### Phase 3: 完善功能 (2周)
- **Week 7**：安全加固、权限控制
- **Week 8**：测试、部署、文档

### Phase 4: 优化发布 (1周)
- **Week 9**：性能优化、监控告警

## 12. 风险评估

### 12.1 技术风险
- **Docker依赖**：Docker服务不可用影响执行
  - 缓解措施：多Docker Host支持
- **大表DDL**：大表操作可能影响性能
  - 缓解措施：资源限制和负载监控
- **并发执行**：多任务并发可能资源竞争
  - 缓解措施：队列机制和资源调度

### 12.2 业务风险
- **数据安全**：DDL操作可能导致数据丢失
  - 缓解措施：操作确认和备份机制
- **权限滥用**：用户误操作生产数据库
  - 缓解措施：严格权限控制和审计
- **服务依赖**：pt工具版本兼容性问题
  - 缓解措施：版本锁定和兼容性测试

### 12.3 运维风险
- **容器管理**：容器资源泄露
  - 缓解措施：自动清理机制
- **日志存储**：日志文件过大
  - 缓解措施：日志轮转和清理策略
- **备份恢复**：数据备份失效
  - 缓解措施：定期备份验证

---

**文档版本**: v1.0  
**最后更新**: 2024-01-15  
**负责人**: 开发团队  
**审核人**: 技术负责人