<template>
  <div class="history-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <el-icon class="title-icon"><Document /></el-icon>
          执行历史
        </h1>
        <p class="page-description">DDL执行历史记录、日志查看和结果分析</p>
      </div>
      
      <div class="header-right">
        <el-button @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
      </div>
    </div>

    <!-- 筛选区域 -->
    <div class="filter-section">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="执行状态">
          <el-select v-model="filterForm.status" placeholder="全部状态" clearable>
            <el-option label="全部状态" value="" />
            <el-option label="等待中" value="pending" />
            <el-option label="执行中" value="running" />
            <el-option label="已完成" value="completed" />
            <el-option label="失败" value="failed" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="连接">
          <el-select v-model="filterForm.connection_id" placeholder="全部连接" clearable filterable>
            <el-option label="全部连接" value="" />
            <el-option
              v-for="connection in connections"
              :key="connection.id"
              :label="connection.name"
              :value="connection.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            @change="handleDateRangeChange"
          />
        </el-form-item>
        
        <el-form-item>
          <el-input
            v-model="filterForm.keyword"
            placeholder="搜索表名、数据库名或执行ID"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleResetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 执行记录表格 -->
    <div class="table-section">
      <el-table
        :data="executions"
        v-loading="loading"
        stripe
        @row-click="handleRowClick"
        style="width: 100%"
      >
        <el-table-column prop="id" label="执行ID" width="120">
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)">
              {{ row.id.slice(-8) }}
            </el-link>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">
              <el-icon class="status-icon">
                <Loading v-if="row.status === 'running'" />
                <CircleCheck v-else-if="row.status === 'completed'" />
                <CircleClose v-else-if="row.status === 'failed'" />
                <Clock v-else-if="row.status === 'pending'" />
                <CloseBold v-else-if="row.status === 'cancelled'" />
              </el-icon>
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="连接信息" min-width="200">
          <template #default="{ row }">
            <div>
              <div class="connection-name">{{ row.connection?.name || 'N/A' }}</div>
              <div class="connection-details">
                <el-tag :type="getEnvironmentTagType(row.connection?.environment)" size="small">
                  {{ getEnvironmentLabel(row.connection?.environment) }}
                </el-tag>
                <span class="host-info">{{ row.connection?.host }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="目标对象" min-width="180">
          <template #default="{ row }">
            <div>
              <div class="database-name">{{ row.database_name }}</div>
              <div class="table-name">{{ row.table_name }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="ddl_type" label="操作类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ getDDLTypeLabel(row.ddl_type) }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="进度" width="160">
          <template #default="{ row }">
            <div v-if="row.status === 'running'">
              <el-progress
                :percentage="getProgress(row)"
                :status="row.status === 'failed' ? 'exception' : undefined"
                :stroke-width="6"
              />
              <div class="progress-text">
                {{ formatNumber(row.processed_rows) }} / {{ formatNumber(row.total_rows) }}
              </div>
            </div>
            <div v-else-if="row.status === 'completed'">
              <el-progress :percentage="100" status="success" :stroke-width="6" />
              <div class="progress-text">{{ formatNumber(row.total_rows) }} 行</div>
            </div>
            <div v-else>
              <span class="progress-placeholder">-</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="耗时" width="120">
          <template #default="{ row }">
            {{ formatDuration(row.duration_seconds) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click.stop="handleViewDetail(row)">
              详情
            </el-button>
            <el-button 
              v-if="row.status === 'running'" 
              size="small" 
              type="warning"
              @click.stop="handleStop(row)"
            >
              停止
            </el-button>
            <el-button 
              v-else-if="row.status === 'failed' || row.status === 'cancelled'" 
              size="small" 
              type="info"
              @click.stop="handleReExecute(row)"
            >
              重试
            </el-button>
            <el-button 
              size="small" 
              type="danger"
              @click.stop="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && executions.length === 0" class="empty-state">
      <el-empty description="暂无执行记录" :image-size="120">
        <el-button type="primary" @click="$router.push('/execution')">
          立即创建执行任务
        </el-button>
      </el-empty>
    </div>

    <!-- 执行详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="`执行详情 - ${currentExecution?.id?.slice(-8) || ''}`"
      width="80%"
      class="detail-dialog"
    >
      <div v-if="currentExecution" class="execution-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="执行ID">
            {{ currentExecution.id }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusTagType(currentExecution.status)">
              {{ getStatusLabel(currentExecution.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="连接">
            {{ currentExecution.connection?.name }}
          </el-descriptions-item>
          <el-descriptions-item label="目标表">
            {{ currentExecution.database_name }}.{{ currentExecution.table_name }}
          </el-descriptions-item>
          <el-descriptions-item label="操作类型">
            {{ getDDLTypeLabel(currentExecution.ddl_type) }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDateTime(currentExecution.created_at) }}
          </el-descriptions-item>
        </el-descriptions>
        
        <div class="detail-section">
          <h4>执行日志</h4>
          <div class="log-viewer">
            <pre class="log-content">{{ currentExecution.execution_logs || '暂无日志' }}</pre>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Document,
  Refresh,
  Download,
  Search,
  Loading,
  CircleCheck,
  CircleClose,
  Clock,
  CloseBold
} from '@element-plus/icons-vue'
import type { ExecutionRecord } from '@/types/execution'

// 路由
const route = useRoute()
const router = useRouter()

// 响应式数据
const loading = ref(false)
const executions = ref<ExecutionRecord[]>([])
const connections = ref<any[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const detailDialogVisible = ref(false)
const currentExecution = ref<ExecutionRecord | null>(null)

// 筛选表单
const filterForm = reactive({
  status: '',
  connection_id: '',
  keyword: '',
  dateRange: null as [string, string] | null,
  start_date: '',
  end_date: ''
})

// 方法
const fetchExecutions = async () => {
  try {
    loading.value = true
    
    const params = {
      page: currentPage.value,
      size: pageSize.value,
      status: filterForm.status || undefined,
      connection_id: filterForm.connection_id || undefined,
      start_date: filterForm.start_date || undefined,
      end_date: filterForm.end_date || undefined
    }
    
    // TODO: 调用API获取执行记录
    console.log('获取执行记录:', params)
    
    // 模拟数据
    executions.value = []
    total.value = 0
    
    ElMessage.info('执行历史功能开发中...')
  } catch (error) {
    console.error('获取执行历史失败:', error)
    ElMessage.error('获取执行历史失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchExecutions()
}

const handleResetFilter = () => {
  Object.assign(filterForm, {
    status: '',
    connection_id: '',
    keyword: '',
    dateRange: null,
    start_date: '',
    end_date: ''
  })
  currentPage.value = 1
  fetchExecutions()
}

const handleDateRangeChange = (dateRange: [string, string] | null) => {
  if (dateRange) {
    filterForm.start_date = dateRange[0]
    filterForm.end_date = dateRange[1]
  } else {
    filterForm.start_date = ''
    filterForm.end_date = ''
  }
}

const handleRefresh = () => {
  fetchExecutions()
}

const handleExport = () => {
  ElMessage.info('导出功能开发中...')
}

const handleRowClick = (row: ExecutionRecord) => {
  handleViewDetail(row)
}

const handleViewDetail = (execution: ExecutionRecord) => {
  currentExecution.value = execution
  detailDialogVisible.value = true
}

const handleStop = async (execution: ExecutionRecord) => {
  try {
    await ElMessageBox.confirm(
      `确定要停止执行任务 ${execution.id.slice(-8)} 吗？`,
      '确认停止',
      {
        type: 'warning',
        confirmButtonText: '确定停止',
        cancelButtonText: '取消'
      }
    )
    
    ElMessage.info('停止执行功能开发中...')
  } catch (error) {
    // 用户取消操作
  }
}

const handleReExecute = async (execution: ExecutionRecord) => {
  try {
    await ElMessageBox.confirm(
      `确定要重新执行任务 ${execution.id.slice(-8)} 吗？`,
      '确认重新执行',
      {
        type: 'info',
        confirmButtonText: '确定重新执行',
        cancelButtonText: '取消'
      }
    )
    
    ElMessage.info('重新执行功能开发中...')
  } catch (error) {
    // 用户取消操作
  }
}

const handleDelete = async (execution: ExecutionRecord) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除执行记录 ${execution.id.slice(-8)} 吗？删除后不可恢复。`,
      '确认删除',
      {
        type: 'warning',
        confirmButtonText: '确定删除',
        cancelButtonText: '取消'
      }
    )
    
    ElMessage.info('删除功能开发中...')
  } catch (error) {
    // 用户取消操作
  }
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  fetchExecutions()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  fetchExecutions()
}

// 工具方法
const getStatusTagType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'info',
    running: 'warning',
    completed: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    pending: '等待中',
    running: '执行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return labels[status] || status
}

const getEnvironmentTagType = (environment: string) => {
  const types: Record<string, string> = {
    prod: 'danger',
    test: 'warning',
    dev: 'success'
  }
  return types[environment] || 'info'
}

const getEnvironmentLabel = (environment: string) => {
  const labels: Record<string, string> = {
    prod: '生产',
    test: '测试',
    dev: '开发'
  }
  return labels[environment] || environment
}

const getDDLTypeLabel = (ddlType: string) => {
  const labels: Record<string, string> = {
    fragment: '碎片整理',
    add_column: '添加列',
    modify_column: '修改列',
    drop_column: '删除列',
    add_index: '添加索引',
    drop_index: '删除索引',
    other: '自定义'
  }
  return labels[ddlType] || ddlType
}

const getProgress = (execution: ExecutionRecord) => {
  if (execution.total_rows === 0) return 0
  return Math.round((execution.processed_rows / execution.total_rows) * 100)
}

const formatNumber = (num: number) => {
  return num?.toLocaleString() || '0'
}

const formatDuration = (seconds: number | undefined) => {
  if (!seconds) return '-'
  if (seconds < 60) return `${seconds}秒`
  if (seconds < 3600) return `${Math.round(seconds / 60)}分钟`
  return `${Math.round(seconds / 3600)}小时`
}

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 生命周期
onMounted(async () => {
  // 检查URL参数
  const executionId = route.query.execution_id
  if (executionId) {
    console.log('查看特定执行记录:', executionId)
  }
  
  await fetchExecutions()
})
</script>

<style scoped>
.history-page {
  padding: 24px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-left {
  flex: 1;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.title-icon {
  color: #1890ff;
  font-size: 28px;
}

.page-description {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

.header-right {
  display: flex;
  gap: 12px;
}

.filter-section {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.filter-form {
  margin: 0;
}

.table-section {
  background: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.status-icon {
  margin-right: 4px;
}

.connection-name {
  font-weight: 500;
  color: #303133;
}

.connection-details {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.host-info {
  font-size: 12px;
  color: #909399;
}

.database-name {
  font-weight: 500;
  color: #303133;
}

.table-name {
  font-size: 12px;
  color: #606266;
  margin-top: 2px;
}

.progress-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  text-align: center;
}

.progress-placeholder {
  color: #c0c4cc;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.empty-state {
  background: white;
  border-radius: 8px;
  padding: 60px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.detail-dialog :deep(.el-dialog__body) {
  padding: 20px;
}

.execution-detail {
  margin-bottom: 20px;
}

.detail-section {
  margin-top: 24px;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.log-viewer {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.log-content {
  margin: 0;
  padding: 16px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  background-color: #fafbfc;
  color: #606266;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .history-page {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  
  .filter-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .filter-form :deep(.el-form-item) {
    margin-right: 0;
    margin-bottom: 0;
  }
  
  .table-section {
    padding: 16px;
    overflow-x: auto;
  }
}
</style>