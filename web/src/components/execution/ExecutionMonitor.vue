<template>
  <div class="execution-monitor">
    <!-- 监控头部 -->
    <div class="monitor-header">
      <div class="header-left">
        <h3>实时监控</h3>
        <div class="connection-status">
          <el-badge :is-dot="true" :type="connectionStatusType">
            <span>{{ connectionStatusText }}</span>
          </el-badge>
        </div>
      </div>
      
      <div class="header-right">
        <el-button @click="handleReconnect" :loading="reconnecting" size="small">
          <el-icon><Refresh /></el-icon>
          重连
        </el-button>
        <el-button @click="handleClearLogs" size="small">
          <el-icon><Delete /></el-icon>
          清空日志
        </el-button>
      </div>
    </div>

    <!-- 任务进度卡片 -->
    <div v-if="currentTask" class="task-progress-card">
      <div class="task-info">
        <div class="task-header">
          <span class="task-id">{{ currentTask.execution_id.slice(-8) }}</span>
          <el-tag :type="getStatusTagType(currentTask.status)" size="small">
            {{ getStatusLabel(currentTask.status) }}
          </el-tag>
        </div>
        
        <div class="task-meta">
          <span>{{ currentTask.database_name }}.{{ currentTask.table_name }}</span>
          <span class="task-stage">{{ currentTask.current_stage }}</span>
        </div>
      </div>
      
      <div class="progress-section">
        <el-progress
          :percentage="Math.round(currentTask.progress)"
          :status="currentTask.status === 'failed' ? 'exception' : undefined"
          :stroke-width="12"
          :show-text="false"
        />
        
        <div class="progress-stats">
          <div class="stat-item">
            <span class="stat-label">进度:</span>
            <span class="stat-value">{{ Math.round(currentTask.progress) }}%</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">已处理:</span>
            <span class="stat-value">{{ formatNumber(currentTask.processed_rows) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">总计:</span>
            <span class="stat-value">{{ formatNumber(currentTask.total_rows) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">速度:</span>
            <span class="stat-value">{{ currentTask.current_speed.toFixed(1) }} rows/s</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">预计剩余:</span>
            <span class="stat-value">{{ estimatedRemaining }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-monitor">
      <el-empty description="暂无正在执行的任务" :image-size="100">
        <el-button type="primary" @click="$router.push('/execution')">
          创建新任务
        </el-button>
      </el-empty>
    </div>

    <!-- 日志区域 -->
    <div class="logs-section">
      <div class="logs-header">
        <h4>执行日志</h4>
        <div class="logs-controls">
          <el-switch
            v-model="autoScroll"
            active-text="自动滚动"
            inactive-text="停止滚动"
            size="small"
          />
          <el-switch
            v-model="showTimestamp"
            active-text="显示时间"
            inactive-text="隐藏时间"
            size="small"
          />
        </div>
      </div>
      
      <div ref="logsContainer" class="logs-container">
        <div v-if="logs.length === 0" class="no-logs">
          暂无日志数据
        </div>
        <div
          v-for="(log, index) in logs"
          :key="index"
          class="log-line"
          :class="getLogLevel(log.content)"
        >
          <span v-if="showTimestamp" class="log-timestamp">
            {{ formatLogTime(log.timestamp) }}
          </span>
          <span class="log-content">{{ log.content }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Delete } from '@element-plus/icons-vue'

// Props
interface Props {
  executionId?: string
  autoConnect?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  executionId: '',
  autoConnect: true
})

// Emits
interface Emits {
  (e: 'statusChange', status: string): void
  (e: 'progressUpdate', progress: any): void
  (e: 'taskComplete', taskId: string): void
}

const emit = defineEmits<Emits>()

// 响应式数据
const ws = ref<WebSocket | null>(null)
const connected = ref(false)
const reconnecting = ref(false)
const currentTask = ref<any>(null)
const logs = ref<{ timestamp: string; content: string }[]>([])
const autoScroll = ref(true)
const showTimestamp = ref(true)
const logsContainer = ref<HTMLElement>()

// 计算属性
const connectionStatusType = computed(() => {
  return connected.value ? 'success' : 'danger'
})

const connectionStatusText = computed(() => {
  return connected.value ? '已连接' : '未连接'
})

const estimatedRemaining = computed(() => {
  if (!currentTask.value || currentTask.value.current_speed === 0) {
    return '-'
  }
  
  const remaining = currentTask.value.total_rows - currentTask.value.processed_rows
  const seconds = remaining / currentTask.value.current_speed
  
  return formatDuration(seconds)
})

// 方法
const connectWebSocket = () => {
  try {
    // 通过Vite代理访问后端WS，并携带token进行鉴权
    const token = localStorage.getItem('auth_token') || ''
    const wsUrl = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/api/v1/ws/execution?execution_id=${props.executionId}&token=${encodeURIComponent(token)}`
    ws.value = new WebSocket(wsUrl)
    
    ws.value.onopen = () => {
      connected.value = true
      reconnecting.value = false
      console.log('WebSocket连接已建立')
      ElMessage.success('实时监控连接成功')
    }
    
    ws.value.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        handleWebSocketMessage(message)
      } catch (error) {
        console.error('解析WebSocket消息失败:', error)
      }
    }
    
    ws.value.onclose = () => {
      connected.value = false
      console.log('WebSocket连接已关闭')
      
      // 自动重连（如果不是主动断开）
      if (!reconnecting.value) {
        setTimeout(() => {
          if (!connected.value) {
            handleReconnect()
          }
        }, 3000)
      }
    }
    
    ws.value.onerror = (error) => {
      console.error('WebSocket连接错误:', error)
      ElMessage.error('实时监控连接失败')
    }
  } catch (error) {
    console.error('创建WebSocket连接失败:', error)
    ElMessage.error('无法建立实时监控连接')
  }
}

const disconnectWebSocket = () => {
  if (ws.value) {
    reconnecting.value = true
    ws.value.close()
    ws.value = null
  }
  connected.value = false
}

const handleWebSocketMessage = (message: any) => {
  switch (message.type) {
    case 'execution_progress':
      handleProgressUpdate(message.data)
      break
    case 'execution_log':
      handleLogMessage(message.data)
      break
    case 'execution_status':
      handleStatusChange(message.data)
      break
    case 'error':
      handleErrorMessage(message.data)
      break
    default:
      console.log('未知消息类型:', message)
  }
}
// 当 executionId 发生变化时，重连并清空旧状态
watch(
  () => props.executionId,
  (newId, oldId) => {
    if (newId && newId !== oldId) {
      // 清空旧任务与日志
      currentTask.value = null
      logs.value = []

      // 重连到新的 executionId
      handleReconnect()
    }
  }
)

const handleProgressUpdate = (data: any) => {
  if (currentTask.value) {
    Object.assign(currentTask.value, data)
  } else {
    currentTask.value = data
  }
  
  emit('progressUpdate', data)
}

const handleLogMessage = (data: any) => {
  const logEntry = {
    timestamp: data.timestamp || new Date().toISOString(),
    content: data.log_line || data.message || ''
  }
  
  logs.value.push(logEntry)
  
  // 限制日志数量
  if (logs.value.length > 1000) {
    logs.value = logs.value.slice(-500)
  }
  
  // 自动滚动到底部
  if (autoScroll.value) {
    nextTick(() => {
      scrollToBottom()
    })
  }
}

const handleStatusChange = (data: any) => {
  if (currentTask.value) {
    currentTask.value.status = data.status
  }
  
  emit('statusChange', data.status)
  
  // 如果任务完成，触发完成事件
  if (data.status === 'completed' || data.status === 'failed' || data.status === 'cancelled') {
    emit('taskComplete', data.execution_id)
  }
}

const handleErrorMessage = (data: any) => {
  const errorLog = {
    timestamp: new Date().toISOString(),
    content: `ERROR: ${data.error_message || data.message || '未知错误'}`
  }
  
  logs.value.push(errorLog)
  ElMessage.error(errorLog.content)
}

const handleReconnect = async () => {
  reconnecting.value = true
  disconnectWebSocket()
  
  await new Promise(resolve => setTimeout(resolve, 1000))
  
  if (props.autoConnect) {
    connectWebSocket()
  }
}

const handleClearLogs = () => {
  logs.value = []
  ElMessage.info('日志已清空')
}

const scrollToBottom = () => {
  if (logsContainer.value) {
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight
  }
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

const getLogLevel = (content: string) => {
  const upperContent = content.toUpperCase()
  if (upperContent.includes('ERROR') || upperContent.includes('FAILED')) {
    return 'log-error'
  }
  if (upperContent.includes('WARN') || upperContent.includes('WARNING')) {
    return 'log-warning'
  }
  if (upperContent.includes('INFO') || upperContent.includes('SUCCESS')) {
    return 'log-info'
  }
  return 'log-default'
}

const formatNumber = (num: number) => {
  return num?.toLocaleString() || '0'
}

const formatDuration = (seconds: number) => {
  if (seconds < 60) return `${Math.round(seconds)}秒`
  if (seconds < 3600) return `${Math.round(seconds / 60)}分钟`
  return `${Math.round(seconds / 3600)}小时`
}

const formatLogTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString('zh-CN')
}

// 暴露方法给父组件
defineExpose({
  connect: connectWebSocket,
  disconnect: disconnectWebSocket,
  clearLogs: handleClearLogs,
  scrollToBottom
})

// 生命周期
onMounted(() => {
  if (props.autoConnect) {
    connectWebSocket()
  }
})

onUnmounted(() => {
  disconnectWebSocket()
})
</script>

<style scoped>
.execution-monitor {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #ebeef5;
  background-color: #fafafa;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.connection-status {
  font-size: 12px;
}

.header-right {
  display: flex;
  gap: 8px;
}

.task-progress-card {
  padding: 20px;
  border-bottom: 1px solid #ebeef5;
}

.task-info {
  margin-bottom: 16px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.task-id {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: #606266;
}

.task-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: #909399;
}

.task-stage {
  font-style: italic;
}

.progress-section {
  margin-bottom: 8px;
}

.progress-stats {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  flex-wrap: wrap;
  gap: 8px;
}

.stat-item {
  font-size: 12px;
}

.stat-label {
  color: #909399;
  margin-right: 4px;
}

.stat-value {
  color: #303133;
  font-weight: 500;
}

.empty-monitor {
  padding: 40px 20px;
  text-align: center;
}

.logs-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 300px;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #ebeef5;
  background-color: #fafafa;
}

.logs-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.logs-controls {
  display: flex;
  gap: 16px;
}

.logs-container {
  flex: 1;
  padding: 0;
  max-height: 400px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  background-color: #fafbfc;
}

.no-logs {
  padding: 40px 20px;
  text-align: center;
  color: #c0c4cc;
  font-style: italic;
}

.log-line {
  padding: 4px 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.log-line:hover {
  background-color: rgba(0, 0, 0, 0.02);
}

.log-timestamp {
  color: #909399;
  font-size: 11px;
  white-space: nowrap;
  min-width: 80px;
}

.log-content {
  flex: 1;
  word-break: break-all;
}

.log-error {
  background-color: #fef0f0;
  color: #f56c6c;
}

.log-warning {
  background-color: #fdf6ec;
  color: #e6a23c;
}

.log-info {
  background-color: #f0f9ff;
  color: #409eff;
}

.log-default {
  color: #606266;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .monitor-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .header-left {
    justify-content: space-between;
  }
  
  .progress-stats {
    justify-content: center;
    gap: 16px;
  }
  
  .stat-item {
    text-align: center;
  }
  
  .logs-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .logs-controls {
    justify-content: center;
  }
}
</style>