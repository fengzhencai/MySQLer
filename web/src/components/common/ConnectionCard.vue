<template>
  <el-card class="connection-card" shadow="hover">
    <!-- 卡片头部 -->
    <template #header>
      <div class="card-header">
        <div class="connection-info">
          <div class="connection-name">
            <el-icon class="name-icon">
              <Database />
            </el-icon>
            {{ connection.name }}
          </div>
          <el-tag
            :type="environmentTagType"
            size="small"
            class="env-tag"
          >
            {{ environmentLabel }}
          </el-tag>
        </div>
        
        <div class="card-actions">
          <el-button
            type="primary"
            size="small"
            @click="handleTest"
            :loading="testing"
            circle
          >
            <el-icon><Connection /></el-icon>
          </el-button>
          
          <el-dropdown @command="handleCommand">
            <el-button size="small" circle>
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-dropdown-item>
                <el-dropdown-item command="explore">
                  <el-icon><Search /></el-icon>
                  浏览数据库
                </el-dropdown-item>
                <el-dropdown-item command="duplicate">
                  <el-icon><CopyDocument /></el-icon>
                  复制配置
                </el-dropdown-item>
                <el-dropdown-item command="delete" divided>
                  <el-icon><Delete /></el-icon>
                  删除
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </template>

    <!-- 卡片内容 -->
    <div class="card-content">
      <!-- 连接信息 -->
      <div class="connection-details">
        <div class="detail-item">
          <el-icon class="detail-icon"><Monitor /></el-icon>
          <span class="detail-label">主机:</span>
          <span class="detail-value">{{ connection.host }}:{{ connection.port }}</span>
        </div>
        
        <div class="detail-item">
          <el-icon class="detail-icon"><User /></el-icon>
          <span class="detail-label">用户:</span>
          <span class="detail-value">{{ connection.username }}</span>
        </div>
        
        <div class="detail-item">
          <el-icon class="detail-icon"><Folder /></el-icon>
          <span class="detail-label">数据库:</span>
          <span class="detail-value">{{ connection.database_name }}</span>
        </div>
        
        <div class="detail-item">
          <el-icon class="detail-icon"><Setting /></el-icon>
          <span class="detail-label">字符集:</span>
          <span class="detail-value">{{ connection.charset }}</span>
          <el-tag v-if="connection.use_ssl" type="success" size="small" class="ssl-tag">
            SSL
          </el-tag>
        </div>
      </div>

      <!-- 描述信息 -->
      <div v-if="connection.description" class="connection-description">
        <el-text type="info" size="small">
          {{ connection.description }}
        </el-text>
      </div>

      <!-- 创建信息 -->
      <div class="connection-meta">
        <div class="meta-item">
          <el-text type="info" size="small">
            创建者: {{ connection.created_by }}
          </el-text>
        </div>
        <div class="meta-item">
          <el-text type="info" size="small">
            创建时间: {{ formatDate(connection.created_at) }}
          </el-text>
        </div>
      </div>
    </div>

    <!-- 测试状态 -->
    <div v-if="testResult" class="test-result">
      <el-alert
        :type="testResult.success ? 'success' : 'error'"
        :title="testResult.success ? '连接正常' : '连接失败'"
        :description="testResult.message"
        :closable="false"
        show-icon
      />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Database,
  Connection,
  MoreFilled,
  Edit,
  Search,
  CopyDocument,
  Delete,
  Monitor,
  User,
  Folder,
  Setting
} from '@element-plus/icons-vue'
import type { Connection as ConnectionType, ConnectionTestResult } from '@/types/connection'
import { useConnectionStore } from '@/stores/connection'

// Props
interface Props {
  connection: ConnectionType
}

const props = defineProps<Props>()

// Emits
interface Emits {
  (e: 'edit', connection: ConnectionType): void
  (e: 'delete', connection: ConnectionType): void
  (e: 'duplicate', connection: ConnectionType): void
  (e: 'explore', connection: ConnectionType): void
}

const emit = defineEmits<Emits>()

// 响应式数据
const connectionStore = useConnectionStore()
const testing = ref(false)
const testResult = ref<ConnectionTestResult | null>(null)

// 计算属性
const environmentTagType = computed(() => {
  switch (props.connection.environment) {
    case 'prod':
      return 'danger'
    case 'test':
      return 'warning'
    case 'dev':
      return 'success'
    default:
      return 'info'
  }
})

const environmentLabel = computed(() => {
  switch (props.connection.environment) {
    case 'prod':
      return '生产'
    case 'test':
      return '测试'
    case 'dev':
      return '开发'
    default:
      return '未知'
  }
})

// 方法
const handleTest = async () => {
  try {
    testing.value = true
    testResult.value = await connectionStore.testConnection(props.connection.id)
  } catch (error) {
    console.error('连接测试失败:', error)
  } finally {
    testing.value = false
  }
}

const handleCommand = (command: string) => {
  switch (command) {
    case 'edit':
      emit('edit', props.connection)
      break
    case 'delete':
      emit('delete', props.connection)
      break
    case 'duplicate':
      emit('duplicate', props.connection)
      break
    case 'explore':
      emit('explore', props.connection)
      break
    default:
      console.warn('未知操作:', command)
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}
</script>

<style scoped>
.connection-card {
  margin-bottom: 16px;
  transition: all 0.3s ease;
}

.connection-card:hover {
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.connection-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.connection-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 16px;
  color: #303133;
}

.name-icon {
  color: #1890ff;
}

.env-tag {
  font-weight: 500;
}

.card-actions {
  display: flex;
  gap: 8px;
}

.card-content {
  padding: 0;
}

.connection-details {
  margin-bottom: 16px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 14px;
}

.detail-icon {
  color: #909399;
  font-size: 16px;
  width: 16px;
}

.detail-label {
  color: #606266;
  font-weight: 500;
  min-width: 60px;
}

.detail-value {
  color: #303133;
  flex: 1;
}

.ssl-tag {
  margin-left: 8px;
}

.connection-description {
  margin-bottom: 16px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
  border-left: 3px solid #1890ff;
}

.connection-meta {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
}

.meta-item {
  font-size: 12px;
}

.test-result {
  margin-top: 16px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .connection-meta {
    flex-direction: column;
    gap: 4px;
  }
  
  .detail-item {
    flex-wrap: wrap;
  }
  
  .detail-label {
    min-width: auto;
  }
}
</style>