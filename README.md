# MySQLer

MySQLer 是一个基于Web的 pt-online-schema-change 工具管理平台，提供可视化界面来执行MySQL在线表结构变更。

## 项目特点

- 🚀 **在线DDL执行**: 基于pt-online-schema-change的安全DDL操作
- 🎯 **可视化界面**: 友好的Web界面，降低操作门槛
- 📊 **实时监控**: WebSocket实时推送执行进度和日志
- 🔒 **安全可靠**: 完整的权限控制和操作审计
- 🐳 **容器化部署**: Docker Compose一键部署
- 📱 **响应式设计**: 支持桌面端和移动端访问

## 技术栈

### 后端
- **Go 1.21+** - 高性能后端服务
- **Gin** - Web框架
- **GORM** - ORM框架
- **MySQL 8.0+** - 数据存储
- **Redis** - 缓存和会话存储
- **Docker** - 容器化执行环境
- **WebSocket** - 实时通信

### 前端
- **Vue 3** + **TypeScript** - 现代化前端框架
- **Element Plus** - UI组件库
- **Vite** - 构建工具
- **Pinia** - 状态管理
- **Axios** - HTTP客户端

## 快速开始

### 环境要求
- Docker 20.0+
- Docker Compose 2.0+
- Go 1.21+ (开发环境)
- Node.js 18+ (开发环境)

### 开发环境启动

```bash
# 克隆项目
git clone https://github.com/fengzhencai/MySQLer.git
cd MySQLer

# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 前端开发
cd web
npm install
npm run dev

# 后端开发
cd backend
go mod tidy
go run cmd/server/main.go
```

### 生产环境部署

```bash
# 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置数据库密码等

# 启动服务
docker-compose up -d
```

## 核心功能

### 1. 连接管理
- 多环境数据库连接配置
- 连接测试和状态监控
- 安全的密码加密存储

### 2. DDL执行
- **碎片整理**: 一键表优化
- **自定义DDL**: 支持各种表结构变更
- **命令预览**: 执行前预览生成的pt命令
- **安全确认**: 危险操作二次确认

### 3. 实时监控
- 执行进度实时显示
- 详细执行日志查看
- 系统负载监控
- 操作控制（暂停/停止/重启）

### 4. 历史记录
- 完整的执行历史记录
- 多维度筛选和搜索
- 执行统计和报告
- 历史命令重复执行

### 5. 系统管理
- 用户权限管理
- 操作审计日志
- 系统配置管理

## 项目结构

```
MySQLer/
├── backend/                 # Go后端服务
│   ├── cmd/                # 应用入口
│   ├── internal/           # 内部包
│   │   ├── config/        # 配置管理
│   │   ├── models/        # 数据模型
│   │   ├── services/      # 业务逻辑
│   │   ├── handlers/      # HTTP处理器
│   │   └── middleware/    # 中间件
│   └── pkg/               # 公共包
├── web/                    # Vue前端应用
│   ├── src/               # 源代码
│   │   ├── components/    # 组件
│   │   ├── views/         # 页面
│   │   ├── stores/        # 状态管理
│   │   └── services/      # API服务
│   └── dist/              # 构建输出
├── docker/                 # Docker配置
├── docs/                   # 项目文档
└── sql/                    # 数据库脚本
```

## 开发规范

### Git分支策略
- `main` - 主分支，生产环境代码
- `develop` - 开发分支，功能集成
- `feature/*` - 功能分支
- `hotfix/*` - 热修复分支

### 提交规范
```
<type>(<scope>): <subject>

<body>

<footer>
```

类型说明：
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建/工具变动

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 支持

如果您在使用过程中遇到问题，可以通过以下方式获取帮助：

- 📖 [项目文档](docs/)
- 🐛 [Issues](https://github.com/fengzhencai/MySQLer/issues)
- 💬 [Discussions](https://github.com/fengzhencai/MySQLer/discussions)

---

**注意**: 本项目正在积极开发中，功能和API可能会发生变化。