# MySQLer Cursor Rules 规则文件说明

## 📁 规则文件总览

基于项目需求文档、架构设计、API设计、UI规范和开发计划，我已经为 MySQLer 项目生成了 8 个 Cursor Rules 文件，涵盖项目开发的各个方面：

### 📋 生成的规则文件列表

| 文件名 | 大小 | 用途 | 适用范围 |
|--------|------|------|----------|
| `project-overview.mdc` | 1.3KB | 项目总览和核心信息 | **所有文件** (alwaysApply) |
| `backend-development.mdc` | 2.0KB | Go后端开发规范 | `*.go`, `go.mod`, `go.sum` |
| `frontend-development.mdc` | 3.2KB | Vue3前端开发规范 | `*.vue`, `*.ts`, `*.js`, `*.scss`, `*.css` |
| `database-design.mdc` | 4.0KB | 数据库设计规范 | `*.sql`, `*migration*`, `*seed*` |
| `docker-deployment.mdc` | 4.3KB | Docker容器化部署 | `Dockerfile`, `docker-compose*.yml` |
| `security-standards.mdc` | 5.4KB | 安全开发规范 | 手动触发 (description) |
| `development-workflow.mdc` | 5.5KB | 开发流程和代码规范 | 手动触发 (description) |
| `api-testing.mdc` | 13KB | API开发与测试规范 | `*_test.go`, `*.test.ts`, `*.spec.ts` |

## 🎯 规则文件功能说明

### 1. project-overview.mdc (项目总览)
- **触发方式**: 自动应用到所有文件
- **核心内容**:
  - 项目架构概述 (Go + Vue3 + Docker)
  - 核心模块说明 (连接管理、DDL执行、实时监控等)
  - 关键文档引用
  - 基础开发规范

### 2. backend-development.mdc (后端开发)
- **触发方式**: Go文件自动应用
- **核心内容**:
  - Go项目结构规范
  - GORM数据模型设计
  - 统一API响应格式
  - JWT认证和权限控制
  - Docker集成和WebSocket实现

### 3. frontend-development.mdc (前端开发)
- **触发方式**: 前端文件自动应用
- **核心内容**:
  - Vue3 + TypeScript + Vite 技术栈
  - Composition API使用规范
  - Element Plus UI组件规范
  - Pinia状态管理
  - 响应式设计和WebSocket集成

### 4. database-design.mdc (数据库设计)
- **触发方式**: SQL文件和迁移文件自动应用
- **核心内容**:
  - 完整的表结构设计
  - 索引设计优化
  - 数据字典和枚举定义
  - 迁移脚本规范
  - 安全性和性能考虑

### 5. docker-deployment.mdc (Docker部署)
- **触发方式**: Docker相关文件自动应用
- **核心内容**:
  - 多阶段Docker构建
  - docker-compose配置
  - PT工具执行容器管理
  - 环境变量和数据持久化
  - 监控和故障排除

### 6. security-standards.mdc (安全规范)
- **触发方式**: 手动调用 (描述：安全开发规范和最佳实践)
- **核心内容**:
  - JWT认证和权限控制
  - 数据加密和传输安全
  - 输入验证和DDL操作安全
  - 操作审计和API安全
  - 容器安全和错误处理

### 7. development-workflow.mdc (开发流程)
- **触发方式**: 手动调用 (描述：开发流程和代码规范)
- **核心内容**:
  - 9周开发阶段规划
  - Git工作流和提交规范
  - 代码审查和测试策略
  - 开发环境配置
  - 部署流程和监控调试

### 8. api-testing.mdc (API测试)
- **触发方式**: 测试文件自动应用
- **核心内容**:
  - RESTful API设计规范
  - 统一响应格式和错误处理
  - WebSocket接口规范
  - 完整的测试策略 (单元/集成/E2E)
  - 性能测试和并发测试

## 🚀 使用方法

### 自动应用的规则
以下规则会根据文件类型自动应用：
- 编辑任何文件时：`project-overview.mdc`
- 编辑Go文件时：`backend-development.mdc`
- 编辑前端文件时：`frontend-development.mdc`
- 编辑SQL文件时：`database-design.mdc`
- 编辑Docker文件时：`docker-deployment.mdc`
- 编辑测试文件时：`api-testing.mdc`

### 手动调用的规则
- 安全相关问题：可以在对话中提到"安全开发规范"来触发安全规则
- 开发流程问题：可以在对话中提到"开发流程"来触发工作流程规则

## 📚 规则内容特点

### 1. 基于项目文档
- 所有规则都基于项目的官方文档生成
- 包含对源文档的引用 (`mdc:docs/xxx.md`)
- 确保规则与项目需求一致

### 2. 实际可操作
- 提供具体的代码示例和配置
- 包含完整的命令和脚本
- 涵盖常见的开发场景

### 3. 层次化设计
- 从总览到具体实现的递进结构
- 不同技术栈分别定义规则
- 开发阶段和测试分离

### 4. 安全优先
- 强调安全开发规范
- 包含完整的认证授权机制
- 敏感数据处理规范

## 🔧 维护和更新

### 规则文件更新
- 当项目需求变更时，更新对应的规则文件
- 新增功能模块时，扩展相关规则
- 定期检查规则文件的有效性

### 规则优化
- 根据开发过程中的反馈优化规则
- 添加新的最佳实践和经验
- 保持规则的简洁和实用性

---

**生成时间**: 2024-01-15  
**项目阶段**: Phase 1 基础架构搭建  
**规则覆盖**: 前端、后端、数据库、部署、安全、测试、流程  
**总规则数**: 8个文件，约40KB内容