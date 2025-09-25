<template>
  <div class="connections-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <el-icon class="title-icon"><Collection /></el-icon>
          连接管理
        </h1>
        <p class="page-description">管理数据库连接配置，支持多环境切换</p>
      </div>
      
      <div class="header-right">
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          新建连接
        </el-button>
      </div>
    </div>

    <!-- 统计信息 -->
    <div class="stats-section">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">{{ connectionStore.connectionCount }}</div>
              <div class="stats-label">总连接数</div>
            </div>
            <el-icon class="stats-icon"><Connection /></el-icon>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stats-card prod">
            <div class="stats-content">
              <div class="stats-number">{{ connectionStore.environmentStats.prod }}</div>
              <div class="stats-label">生产环境</div>
            </div>
            <el-icon class="stats-icon"><Warning /></el-icon>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stats-card test">
            <div class="stats-content">
              <div class="stats-number">{{ connectionStore.environmentStats.test }}</div>
              <div class="stats-label">测试环境</div>
            </div>
            <el-icon class="stats-icon"><Monitor /></el-icon>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stats-card dev">
            <div class="stats-content">
              <div class="stats-number">{{ connectionStore.environmentStats.dev }}</div>
              <div class="stats-label">开发环境</div>
            </div>
            <el-icon class="stats-icon"><Tools /></el-icon>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 搜索和筛选 -->
    <div class="filter-section">
      <el-row :gutter="16" align="middle">
        <el-col :span="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索连接名称、主机地址或数据库名"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        
        <el-col :span="4">
          <el-select
            v-model="selectedEnvironment"
            placeholder="筛选环境"
            clearable
            @change="handleFilter"
          >
            <el-option label="全部环境" value="all" />
            <el-option label="生产环境" value="prod" />
            <el-option label="测试环境" value="test" />
            <el-option label="开发环境" value="dev" />
          </el-select>
        </el-col>
        
        <el-col :span="4">
          <el-select v-model="viewMode" @change="handleViewModeChange">
            <el-option label="卡片视图" value="card">
              <el-icon><Grid /></el-icon>
              卡片视图
            </el-option>
            <el-option label="列表视图" value="table">
              <el-icon><List /></el-icon>
              列表视图
            </el-option>
          </el-select>
        </el-col>
        
        <el-col :span="8" class="text-right">
          <el-button @click="handleRefresh" :loading="connectionStore.loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 连接列表 -->
    <div class="connections-content" v-loading="connectionStore.loading">
      <!-- 卡片视图 -->
      <div v-if="viewMode === 'card'" class="card-view">
        <el-empty v-if="filteredConnections.length === 0" description="暂无连接数据">
          <el-button type="primary" @click="showCreateDialog">
            立即创建
          </el-button>
        </el-empty>
        
        <el-row v-else :gutter="16">
          <el-col
            v-for="connection in filteredConnections"
            :key="connection.id"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
            :xl="6"
          >
            <ConnectionCard
              :connection="connection"
              @edit="handleEdit"
              @delete="handleDelete"
              @duplicate="handleDuplicate"
              @explore="handleExplore"
            />
          </el-col>
        </el-row>
      </div>

      <!-- 表格视图 -->
      <div v-else class="table-view">
        <el-table
          :data="filteredConnections"
          stripe
          style="width: 100%"
          @row-click="handleRowClick"
        >
          <el-table-column prop="name" label="连接名称" min-width="150">
            <template #default="{ row }">
              <div class="connection-name-cell">
                <el-icon class="name-icon"><Collection /></el-icon>
                {{ row.name }}
              </div>
            </template>
          </el-table-column>
          
          <el-table-column prop="environment" label="环境" width="80">
            <template #default="{ row }">
              <el-tag :type="getEnvironmentTagType(row.environment)" size="small">
                {{ getEnvironmentLabel(row.environment) }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column label="主机信息" min-width="200">
            <template #default="{ row }">
              <div>{{ row.host }}:{{ row.port }}</div>
              <el-text type="info" size="small">{{ row.username }}</el-text>
            </template>
          </el-table-column>
          
          <el-table-column prop="database_name" label="数据库" min-width="120" />
          
          <el-table-column label="配置" width="100">
            <template #default="{ row }">
              <div class="config-info">
                <el-tag size="small">{{ row.charset }}</el-tag>
                <el-tag v-if="row.use_ssl" type="success" size="small">SSL</el-tag>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </el-table-column>
          
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click.stop="handleTest(row)" type="primary">
                测试
              </el-button>
              <el-button size="small" @click.stop="handleEdit(row)">
                编辑
              </el-button>
              <el-button size="small" @click.stop="handleDelete(row)" type="danger">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 创建/编辑连接对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '创建连接' : '编辑连接'"
      width="900px"
      @close="handleDialogClose"
    >
      <ConnectionForm
        v-model="currentConnection"
        :mode="dialogMode"
        @submit="handleFormSubmit"
        @test="handleFormTest"
        @cancel="handleDialogClose"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Collection,
  Plus,
  Connection,
  Warning,
  Monitor,
  Tools,
  Search,
  Grid,
  List,
  Refresh
} from '@element-plus/icons-vue'
import { useConnectionStore } from '@/stores/connection'
import ConnectionCard from '@/components/common/ConnectionCard.vue'
import ConnectionForm from '@/components/forms/ConnectionForm.vue'
import type { Connection as ConnectionType, CreateConnectionRequest } from '@/types/connection'

// 响应式数据
const connectionStore = useConnectionStore()
const searchKeyword = ref('')
const selectedEnvironment = ref<string>('all')
const viewMode = ref<'card' | 'table'>('card')
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const currentConnection = ref<ConnectionType | null>(null)

// 计算属性
const filteredConnections = computed(() => {
  let connections = connectionStore.connections

  // 环境筛选
  if (selectedEnvironment.value && selectedEnvironment.value !== 'all') {
    connections = connectionStore.filterByEnvironment(selectedEnvironment.value as any)
  }

  // 关键词搜索
  if (searchKeyword.value.trim()) {
    connections = connectionStore.searchConnections(searchKeyword.value)
  }

  return connections
})

// 方法
const handleSearch = () => {
  // 搜索逻辑已在computed中处理
}

const handleFilter = () => {
  // 筛选逻辑已在computed中处理
}

const handleViewModeChange = () => {
  // 视图模式切换
}

const handleRefresh = async () => {
  await connectionStore.fetchConnections()
}

const showCreateDialog = () => {
  dialogMode.value = 'create'
  currentConnection.value = null
  dialogVisible.value = true
}

const handleEdit = (connection: ConnectionType) => {
  dialogMode.value = 'edit'
  currentConnection.value = connection
  dialogVisible.value = true
}

const handleDelete = async (connection: ConnectionType) => {
  await connectionStore.deleteConnection(connection.id)
}

const handleDuplicate = (connection: ConnectionType) => {
  dialogMode.value = 'create'
  currentConnection.value = {
    ...connection,
    id: '',
    name: `${connection.name} - 副本`,
    created_at: '',
    updated_at: ''
  }
  dialogVisible.value = true
}

const handleExplore = (connection: ConnectionType) => {
  // TODO: 跳转到数据库浏览器
  ElMessage.info('数据库浏览功能开发中...')
}

const handleTest = async (connection: ConnectionType) => {
  await connectionStore.testConnection(connection.id)
}

const handleRowClick = (row: ConnectionType) => {
  handleEdit(row)
}

const handleFormSubmit = async (data: CreateConnectionRequest) => {
  try {
    if (dialogMode.value === 'create') {
      await connectionStore.createConnection(data)
    } else if (currentConnection.value) {
      await connectionStore.updateConnection(currentConnection.value.id, data)
    }
    dialogVisible.value = false
  } catch (error) {
    console.error('操作失败:', error)
  }
}

const handleFormTest = async (data: CreateConnectionRequest) => {
  try {
    await connectionStore.testConnectionByParams(data)
  } catch (e) {
    // 错误提示已在store内处理
  }
}

const handleDialogClose = () => {
  dialogVisible.value = false
  currentConnection.value = null
}

const getEnvironmentTagType = (environment: string) => {
  switch (environment) {
    case 'prod': return 'danger'
    case 'test': return 'warning'
    case 'dev': return 'success'
    default: return 'info'
  }
}

const getEnvironmentLabel = (environment: string) => {
  switch (environment) {
    case 'prod': return '生产'
    case 'test': return '测试'
    case 'dev': return '开发'
    default: return '未知'
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 生命周期
onMounted(async () => {
  await connectionStore.fetchConnections()
})
</script>

<style scoped>
.connections-page {
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

.stats-section {
  margin-bottom: 24px;
}

.stats-card {
  position: relative;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
}

.stats-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stats-card.prod {
  border-left: 4px solid #ff4d4f;
}

.stats-card.test {
  border-left: 4px solid #faad14;
}

.stats-card.dev {
  border-left: 4px solid #52c41a;
}

.stats-content {
  position: relative;
  z-index: 2;
}

.stats-number {
  font-size: 32px;
  font-weight: 700;
  color: #303133;
  line-height: 1;
}

.stats-label {
  font-size: 14px;
  color: #606266;
  margin-top: 8px;
}

.stats-icon {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 40px;
  color: #e4e7ed;
  z-index: 1;
}

.filter-section {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.connections-content {
  background: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  min-height: 400px;
}

.card-view {
  min-height: 300px;
}

.table-view {
  min-height: 300px;
}

.connection-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name-icon {
  color: #1890ff;
}

.config-info {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.text-right {
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .connections-page {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  
  .stats-section :deep(.el-col) {
    margin-bottom: 16px;
  }
  
  .filter-section :deep(.el-row) {
    flex-direction: column;
    gap: 12px;
  }
  
  .filter-section :deep(.el-col) {
    width: 100%;
  }
  
  .text-right {
    text-align: left;
  }
}
</style>