<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>仪表板</h1>
      <p>欢迎使用 MySQLer PT-Online-Schema-Change 管理平台</p>
    </div>
    
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon size="24"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.connections }}</div>
              <div class="stat-label">数据库连接</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon running">
              <el-icon size="24"><Operation /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.runningExecutions }}</div>
              <div class="stat-label">正在执行</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon success">
              <el-icon size="24"><SuccessFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.successExecutions }}</div>
              <div class="stat-label">执行成功</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon failed">
              <el-icon size="24"><CircleCloseFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.failedExecutions }}</div>
              <div class="stat-label">执行失败</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 快速操作 -->
    <el-row :gutter="16" class="quick-actions">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>快速操作</span>
          </template>
          <div class="action-buttons">
            <el-button type="primary" :icon="Plus" @click="$router.push('/connections')">
              新建连接
            </el-button>
            <el-button type="success" :icon="Operation" @click="$router.push('/execution')">
              执行DDL
            </el-button>
            <el-button :icon="Document" @click="$router.push('/history')">
              查看历史
            </el-button>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统状态</span>
          </template>
          <div class="system-status">
            <div class="status-item">
              <span class="status-label">服务状态</span>
              <el-tag type="success">正常</el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">数据库连接</span>
              <el-tag type="success">正常</el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">Docker服务</span>
              <el-tag type="success">正常</el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 最近执行记录 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>最近执行记录</span>
          <el-button type="text" @click="$router.push('/history')">查看全部</el-button>
        </div>
      </template>
      
      <el-table :data="recentExecutions" style="width: 100%">
        <el-table-column prop="id" label="执行ID" width="120" />
        <el-table-column prop="database_name" label="数据库" width="120" />
        <el-table-column prop="table_name" label="表名" width="150" />
        <el-table-column prop="ddl_type" label="操作类型" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.ddl_type" size="small">
              {{ getDDLTypeLabel(scope.row.ddl_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">
              {{ getStatusLabel(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_by" label="执行人" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Connection,
  Operation,
  SuccessFilled,
  CircleCloseFilled,
  Plus,
  Document
} from '@element-plus/icons-vue'
import type { ExecutionRecord, DDLType, ExecutionStatus } from '@/types/execution'

// 统计数据
const stats = ref({
  connections: 0,
  runningExecutions: 0,
  successExecutions: 0,
  failedExecutions: 0
})

// 最近执行记录
const recentExecutions = ref<ExecutionRecord[]>([])

// 获取DDL类型标签
const getDDLTypeLabel = (type: DDLType) => {
  const labels = {
    fragment: '碎片整理',
    add_column: '添加列',
    modify_column: '修改列',
    drop_column: '删除列',
    add_index: '添加索引',
    drop_index: '删除索引',
    other: '其他'
  }
  return labels[type] || type
}

// 获取状态类型
const getStatusType = (status: ExecutionStatus) => {
  const types = {
    pending: '',
    running: 'warning',
    completed: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return types[status] || ''
}

// 获取状态标签
const getStatusLabel = (status: ExecutionStatus) => {
  const labels = {
    pending: '等待中',
    running: '执行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return labels[status] || status
}

// 格式化时间
const formatTime = (timeStr: string) => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

// 加载数据
const loadData = async () => {
  try {
    // TODO: 调用API获取实际数据
    // 这里使用模拟数据
    stats.value = {
      connections: 5,
      runningExecutions: 2,
      successExecutions: 156,
      failedExecutions: 3
    }
    
    recentExecutions.value = [
      {
        id: 'exec-001',
        connection_id: 'conn-001',
        table_name: 'users',
        database_name: 'test_db',
        ddl_type: 'add_column',
        status: 'completed',
        created_by: 'admin',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        processed_rows: 0,
        total_rows: 0,
        generated_command: ''
      }
    ]
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
}

.dashboard-header {
  margin-bottom: 24px;
}

.dashboard-header h1 {
  font-size: 24px;
  color: #303133;
  margin-bottom: 8px;
}

.dashboard-header p {
  color: #909399;
  font-size: 14px;
}

.stats-row {
  margin-bottom: 24px;
}

.stat-card {
  margin-bottom: 0;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #409eff;
  color: white;
}

.stat-icon.running {
  background-color: #e6a23c;
}

.stat-icon.success {
  background-color: #67c23a;
}

.stat-icon.failed {
  background-color: #f56c6c;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.quick-actions {
  margin-bottom: 24px;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.system-status {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-label {
  color: #606266;
  font-size: 14px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>