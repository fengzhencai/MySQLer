# MySQLer UI视觉规范与组件设计

## 1. 设计系统概述

### 1.1 设计理念
- **专业可靠**: 体现数据库管理工具的专业性和稳定性
- **简洁高效**: 减少认知负荷，提高操作效率
- **安全感知**: 通过视觉元素强化安全操作意识
- **响应式**: 适配不同屏幕尺寸和设备

### 1.2 目标用户
- **主要用户**: 数据库运维工程师（25-40岁，技术背景）
- **次要用户**: 后端开发工程师、DBA
- **使用场景**: 办公室环境、多显示器、长时间使用

## 2. 色彩系统

### 2.1 主色调
```scss
// 主品牌色 - 科技蓝
$primary: #1890ff;
$primary-light: #40a9ff;
$primary-dark: #096dd9;
$primary-hover: #40a9ff;
$primary-active: #096dd9;

// 渐变色
$primary-gradient: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
```

### 2.2 功能色彩
```scss
// 成功色 - 安全绿
$success: #52c41a;
$success-light: #73d13d;
$success-dark: #389e0d;

// 警告色 - 注意橙
$warning: #faad14;
$warning-light: #ffc53d;
$warning-dark: #d48806;

// 错误色 - 危险红
$error: #ff4d4f;
$error-light: #ff7875;
$error-dark: #cf1322;

// 信息色 - 中性蓝
$info: #1890ff;
$info-light: #40a9ff;
$info-dark: #096dd9;
```

### 2.3 中性色彩
```scss
// 文本色
$text-primary: #262626;     // 主要文本
$text-secondary: #595959;   // 次要文本
$text-disabled: #bfbfbf;    // 禁用文本
$text-inverse: #ffffff;     // 反色文本

// 背景色
$bg-primary: #ffffff;       // 主要背景
$bg-secondary: #fafafa;     // 次要背景
$bg-tertiary: #f5f5f5;      // 三级背景
$bg-dark: #001529;          // 深色背景

// 边框色
$border-light: #f0f0f0;     // 浅色边框
$border-base: #d9d9d9;      // 基础边框
$border-dark: #434343;      // 深色边框
```

### 2.4 状态色彩映射
| 状态 | 色彩 | 使用场景 |
|------|------|----------|
| pending | #faad14 | 等待执行 |
| running | #1890ff | 执行中 |
| completed | #52c41a | 执行成功 |
| failed | #ff4d4f | 执行失败 |
| cancelled | #8c8c8c | 已取消 |

## 3. 字体系统

### 3.1 字体选择
```scss
// 主要字体栈
$font-family-base: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 
                   'Helvetica Neue', Arial, 'Noto Sans', sans-serif, 
                   'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 
                   'Noto Color Emoji';

// 代码字体栈
$font-family-code: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, 
                   Courier, monospace;

// 数字字体栈
$font-family-number: 'Helvetica Neue', Helvetica, Arial, sans-serif;
```

### 3.2 字体大小规范
```scss
// 基础字号
$font-size-base: 14px;

// 字号系列
$font-size-xs: 12px;     // 辅助信息
$font-size-sm: 13px;     // 小号文本
$font-size-base: 14px;   // 正文
$font-size-lg: 16px;     // 大号正文
$font-size-xl: 18px;     // 小标题
$font-size-xxl: 20px;    // 标题
$font-size-xxxl: 24px;   // 大标题

// 标题系列
$h1-size: 30px;
$h2-size: 24px;
$h3-size: 20px;
$h4-size: 18px;
$h5-size: 16px;
$h6-size: 14px;
```

### 3.3 行高规范
```scss
$line-height-base: 1.5715;
$line-height-sm: 1.5;
$line-height-lg: 1.8;
```

## 4. 间距系统

### 4.1 基础间距单位
```scss
// 基础间距单位 (8px)
$spacing-unit: 8px;

// 间距系列
$spacing-xs: 4px;    // 0.5x
$spacing-sm: 8px;    // 1x
$spacing-md: 16px;   // 2x
$spacing-lg: 24px;   // 3x
$spacing-xl: 32px;   // 4x
$spacing-xxl: 48px;  // 6x
$spacing-xxxl: 64px; // 8x
```

### 4.2 组件间距规范
| 场景 | 间距值 | 说明 |
|------|--------|------|
| 表单项间距 | 16px | 表单元素垂直间距 |
| 按钮间距 | 8px | 按钮之间的水平间距 |
| 卡片内边距 | 24px | 卡片内容区域内边距 |
| 页面边距 | 24px | 页面内容区域边距 |
| 模块间距 | 32px | 不同模块之间的间距 |

## 5. 布局系统

### 5.1 栅格系统
```scss
// 容器最大宽度
$container-max-width: 1200px;

// 栅格列数
$grid-columns: 24;

// 栅格间距
$grid-gutter: 16px;

// 断点定义
$breakpoints: (
  xs: 480px,
  sm: 576px,
  md: 768px,
  lg: 992px,
  xl: 1200px,
  xxl: 1600px
);
```

### 5.2 页面布局结构
```
┌─────────────────────────────────────────────────────────────┐
│                         Header (64px)                       │
├─────────────────────────────────────────────────────────────┤
│ Side │                                                      │
│ Menu │                 Main Content                         │
│(200px)│                 Area                                │
│      │                                                      │
│      │                                                      │
└─────────────────────────────────────────────────────────────┘
```

### 5.3 响应式断点
| 断点 | 屏幕宽度 | 侧边栏 | 内容区 |
|------|----------|--------|--------|
| xs | <480px | 隐藏 | 全宽 |
| sm | 480-576px | 隐藏 | 全宽 |
| md | 576-768px | 折叠 | 适配 |
| lg | 768-992px | 展开 | 适配 |
| xl | 992-1200px | 展开 | 适配 |
| xxl | >1200px | 展开 | 居中 |

## 6. 组件设计规范

### 6.1 按钮组件
```scss
// 主要按钮
.btn-primary {
  background: $primary;
  border: 1px solid $primary;
  color: white;
  height: 32px;
  padding: 4px 15px;
  border-radius: 6px;
  font-size: 14px;
  
  &:hover {
    background: $primary-hover;
    border-color: $primary-hover;
  }
  
  &:active {
    background: $primary-active;
    border-color: $primary-active;
  }
}

// 危险按钮 (用于停止执行、删除等)
.btn-danger {
  background: $error;
  border: 1px solid $error;
  color: white;
  
  &:hover {
    background: $error-light;
  }
}

// 次要按钮
.btn-secondary {
  background: white;
  border: 1px solid $border-base;
  color: $text-primary;
  
  &:hover {
    color: $primary;
    border-color: $primary;
  }
}
```

### 6.2 表单组件
```scss
// 输入框
.input {
  height: 32px;
  padding: 4px 11px;
  border: 1px solid $border-base;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.3s;
  
  &:focus {
    border-color: $primary;
    box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
    outline: none;
  }
  
  &:disabled {
    background: $bg-tertiary;
    color: $text-disabled;
    cursor: not-allowed;
  }
}

// 文本域
.textarea {
  min-height: 80px;
  padding: 6px 11px;
  resize: vertical;
}

// 选择器
.select {
  min-width: 120px;
  height: 32px;
}
```

### 6.3 卡片组件
```scss
.card {
  background: white;
  border: 1px solid $border-light;
  border-radius: 8px;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.03);
  transition: box-shadow 0.3s;
  
  &:hover {
    box-shadow: 0 4px 8px 0 rgba(0, 0, 0, 0.12);
  }
  
  .card-header {
    padding: 16px 24px;
    border-bottom: 1px solid $border-light;
    font-weight: 500;
  }
  
  .card-body {
    padding: 24px;
  }
}
```

### 6.4 状态指示器
```scss
// 执行状态标签
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  
  &.status-pending {
    background: rgba(250, 173, 20, 0.1);
    color: $warning;
  }
  
  &.status-running {
    background: rgba(24, 144, 255, 0.1);
    color: $primary;
    
    &::before {
      content: '';
      width: 6px;
      height: 6px;
      background: $primary;
      border-radius: 50%;
      margin-right: 4px;
      animation: pulse 1.5s infinite;
    }
  }
  
  &.status-completed {
    background: rgba(82, 196, 26, 0.1);
    color: $success;
  }
  
  &.status-failed {
    background: rgba(255, 77, 79, 0.1);
    color: $error;
  }
}

@keyframes pulse {
  0%, 70%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  35% {
    transform: scale(1.3);
    opacity: 0.7;
  }
}
```

### 6.5 进度条组件
```scss
.progress {
  width: 100%;
  height: 8px;
  background: $bg-tertiary;
  border-radius: 4px;
  overflow: hidden;
  
  .progress-bar {
    height: 100%;
    background: linear-gradient(90deg, $primary, $primary-light);
    border-radius: 4px;
    transition: width 0.3s ease;
    position: relative;
    
    &::after {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: linear-gradient(
        90deg,
        transparent,
        rgba(255, 255, 255, 0.3),
        transparent
      );
      animation: shimmer 1.5s infinite;
    }
  }
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}
```

## 7. 图标系统

### 7.1 图标库选择
- **主要图标库**: Ant Design Icons
- **补充图标**: Feather Icons (简洁线性图标)
- **数据库图标**: 自定义SVG图标

### 7.2 图标使用规范
```scss
// 图标尺寸
$icon-size-xs: 12px;
$icon-size-sm: 14px;
$icon-size-base: 16px;
$icon-size-lg: 18px;
$icon-size-xl: 20px;
$icon-size-xxl: 24px;

// 图标颜色
.icon {
  color: $text-secondary;
  
  &.icon-primary { color: $primary; }
  &.icon-success { color: $success; }
  &.icon-warning { color: $warning; }
  &.icon-error { color: $error; }
}
```

### 7.3 常用图标映射
| 功能 | 图标 | 含义 |
|------|------|------|
| 数据库连接 | database | 数据库 |
| DDL执行 | play-circle | 执行 |
| 停止执行 | stop | 停止 |
| 执行历史 | history | 历史 |
| 设置 | setting | 配置 |
| 用户 | user | 用户 |
| 刷新 | reload | 刷新 |
| 下载 | download | 下载 |
| 删除 | delete | 删除 |
| 编辑 | edit | 编辑 |

## 8. 页面布局设计

### 8.1 顶部导航栏
```html
<header class="app-header">
  <div class="header-left">
    <img src="/logo.svg" alt="MySQLer" class="logo">
    <h1 class="app-title">MySQLer</h1>
  </div>
  <div class="header-right">
    <div class="user-info">
      <span class="username">admin</span>
      <div class="user-dropdown">
        <a href="#" class="dropdown-item">个人设置</a>
        <a href="#" class="dropdown-item">退出登录</a>
      </div>
    </div>
  </div>
</header>
```

### 8.2 侧边导航菜单
```html
<aside class="sidebar">
  <nav class="nav-menu">
    <div class="menu-group">
      <div class="group-title">执行管理</div>
      <a href="/execution" class="menu-item active">
        <i class="icon icon-play-circle"></i>
        <span>DDL执行</span>
      </a>
      <a href="/history" class="menu-item">
        <i class="icon icon-history"></i>
        <span>执行历史</span>
      </a>
    </div>
    <div class="menu-group">
      <div class="group-title">系统管理</div>
      <a href="/connections" class="menu-item">
        <i class="icon icon-database"></i>
        <span>连接管理</span>
      </a>
      <a href="/users" class="menu-item">
        <i class="icon icon-user"></i>
        <span>用户管理</span>
      </a>
    </div>
  </nav>
</aside>
```

### 8.3 主内容区域
```html
<main class="main-content">
  <div class="page-header">
    <h2 class="page-title">DDL执行器</h2>
    <div class="page-actions">
      <button class="btn btn-primary">新建连接</button>
    </div>
  </div>
  <div class="page-content">
    <!-- 页面具体内容 -->
  </div>
</main>
```

## 9. 核心页面设计

### 9.1 DDL执行页面
```html
<div class="execution-page">
  <!-- 执行配置区 -->
  <div class="execution-config">
    <div class="card">
      <div class="card-header">
        <h3>执行配置</h3>
      </div>
      <div class="card-body">
        <form class="execution-form">
          <div class="form-group">
            <label>数据库连接</label>
            <select class="select">
              <option>生产环境-报表库</option>
            </select>
          </div>
          <div class="form-group">
            <label>目标表</label>
            <input type="text" class="input" placeholder="请输入表名">
          </div>
          <div class="form-group">
            <label>操作类型</label>
            <div class="radio-group">
              <label class="radio-item">
                <input type="radio" name="type" value="fragment">
                <span>碎片整理</span>
              </label>
              <label class="radio-item">
                <input type="radio" name="type" value="custom">
                <span>自定义DDL</span>
              </label>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
  
  <!-- 执行监控区 -->
  <div class="execution-monitor">
    <div class="card">
      <div class="card-header">
        <h3>执行监控</h3>
        <div class="status-badge status-running">执行中</div>
      </div>
      <div class="card-body">
        <div class="progress-section">
          <div class="progress-info">
            <span>进度: 75%</span>
            <span>剩余时间: 约3分钟</span>
          </div>
          <div class="progress">
            <div class="progress-bar" style="width: 75%"></div>
          </div>
        </div>
        <div class="log-section">
          <div class="log-viewer">
            <!-- 实时日志内容 -->
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
```

### 9.2 连接管理页面
```html
<div class="connections-page">
  <div class="page-header">
    <h2>连接管理</h2>
    <div class="actions">
      <button class="btn btn-primary">
        <i class="icon icon-plus"></i>
        新建连接
      </button>
    </div>
  </div>
  
  <div class="filters">
    <div class="filter-group">
      <label>环境</label>
      <select class="select">
        <option value="">全部</option>
        <option value="prod">生产</option>
        <option value="test">测试</option>
      </select>
    </div>
    <div class="filter-group">
      <label>关键词</label>
      <input type="text" class="input" placeholder="搜索连接名称">
    </div>
  </div>
  
  <div class="connections-grid">
    <div class="connection-card">
      <div class="card-header">
        <div class="connection-info">
          <h4>生产环境-报表库</h4>
          <span class="env-tag env-prod">生产</span>
        </div>
        <div class="connection-actions">
          <button class="btn-icon" title="测试连接">
            <i class="icon icon-wifi"></i>
          </button>
          <button class="btn-icon" title="编辑">
            <i class="icon icon-edit"></i>
          </button>
        </div>
      </div>
      <div class="card-body">
        <div class="connection-details">
          <div class="detail-item">
            <span class="label">主机:</span>
            <span class="value">prod-db.example.com</span>
          </div>
          <div class="detail-item">
            <span class="label">数据库:</span>
            <span class="value">report_db</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
```

## 10. 动效设计

### 10.1 过渡动画
```scss
// 基础过渡时间
$transition-fast: 0.15s;
$transition-base: 0.3s;
$transition-slow: 0.5s;

// 缓动函数
$ease-out: cubic-bezier(0.215, 0.61, 0.355, 1);
$ease-in-out: cubic-bezier(0.645, 0.045, 0.355, 1);

// 通用过渡
.transition {
  transition: all $transition-base $ease-out;
}

// 页面切换动画
.page-enter-active, .page-leave-active {
  transition: opacity $transition-base $ease-out;
}

.page-enter-from, .page-leave-to {
  opacity: 0;
}
```

### 10.2 加载动画
```scss
// 旋转加载器
.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid $border-light;
  border-top: 2px solid $primary;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

// 脉冲加载器
.pulse {
  animation: pulse-animation 1.5s infinite;
}

@keyframes pulse-animation {
  0% { opacity: 1; }
  50% { opacity: 0.5; }
  100% { opacity: 1; }
}
```

### 10.3 微交互
```scss
// 按钮点击反馈
.btn {
  transform: translateY(0);
  transition: all $transition-fast $ease-out;
  
  &:active {
    transform: translateY(1px);
  }
}

// 卡片悬停效果
.card {
  transition: transform $transition-base $ease-out,
              box-shadow $transition-base $ease-out;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  }
}
```

## 11. 暗色主题

### 11.1 暗色主题色彩
```scss
// 暗色主题变量
$dark-bg-primary: #141414;
$dark-bg-secondary: #1f1f1f;
$dark-bg-tertiary: #262626;

$dark-text-primary: #ffffff;
$dark-text-secondary: #a6a6a6;
$dark-text-disabled: #595959;

$dark-border-light: #303030;
$dark-border-base: #434343;

// 主题切换
@media (prefers-color-scheme: dark) {
  :root {
    --bg-primary: #{$dark-bg-primary};
    --bg-secondary: #{$dark-bg-secondary};
    --text-primary: #{$dark-text-primary};
    --text-secondary: #{$dark-text-secondary};
  }
}
```

## 12. 无障碍设计

### 12.1 键盘导航
```scss
// 焦点指示器
.focus-visible {
  outline: 2px solid $primary;
  outline-offset: 2px;
}

// 跳过链接
.skip-link {
  position: absolute;
  top: -40px;
  left: 6px;
  background: $primary;
  color: white;
  padding: 8px;
  text-decoration: none;
  z-index: 1000;
  
  &:focus {
    top: 6px;
  }
}
```

### 12.2 ARIA标签
```html
<!-- 进度条无障碍 -->
<div class="progress" 
     role="progressbar" 
     aria-valuenow="75" 
     aria-valuemin="0" 
     aria-valuemax="100"
     aria-label="DDL执行进度">
  <div class="progress-bar" style="width: 75%"></div>
</div>

<!-- 状态指示器无障碍 -->
<span class="status-badge status-running" 
      role="status" 
      aria-label="当前状态：执行中">
  执行中
</span>
```

## 13. 组件库配置

### 13.1 Element Plus主题定制
```scss
// 覆盖Element Plus默认样式
@import 'element-plus/theme-chalk/src/common/var.scss';

// 自定义主题变量
:root {
  --el-color-primary: #{$primary};
  --el-color-primary-light-3: #{$primary-light};
  --el-color-primary-dark-2: #{$primary-dark};
  --el-border-radius-base: 6px;
  --el-font-size-base: 14px;
}
```

### 13.2 自定义组件前缀
```scss
// 统一组件前缀
.mysql-* {
  // 自定义组件样式
}
```

---

**文档版本**: v1.0  
**创建时间**: 2024-01-15  
**维护人员**: 前端设计团队