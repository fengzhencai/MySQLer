<template>
  <div class="execution-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">
          <el-icon class="title-icon"><Operation /></el-icon>
          DDL执行
        </h1>
        <p class="page-description">安全、高效的在线DDL执行和实时监控</p>
      </div>
      
      <div class="header-right">
        <el-button @click="handleRefreshTasks">
          <el-icon><Refresh /></el-icon>
          刷新状态
        </el-button>
      </div>
    </div>

    <!-- 执行统计 -->
    <div class="stats-section">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">0</div>
              <div class="stats-label">运行中任务</div>
            </div>
            <el-icon class="stats-icon running"><Loading /></el-icon>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">0</div>
              <div class="stats-label">总执行次数</div>
            </div>
            <el-icon class="stats-icon"><DataAnalysis /></el-icon>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stats-card success">
            <div class="stats-content">
              <div class="stats-number">0%</div>
              <div class="stats-label">成功率</div>
            </div>
            <el-icon class="stats-icon"><CircleCheck /></el-icon>
          </el-card>
        </el-col>
        
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-number">0分钟</div>
              <div class="stats-label">平均耗时</div>
            </div>
            <el-icon class="stats-icon"><Timer /></el-icon>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 主要内容区域 -->
    <el-row :gutter="24">
      <!-- 左侧：新建执行 -->
      <el-col :span="16">
        <el-card class="execution-form-card">
          <template #header>
            <div class="card-header">
              <h3>新建DDL执行</h3>
              <el-button v-if="previewResult" @click="handleClearPreview" size="small">
                清空预览
              </el-button>
            </div>
          </template>

          <el-form
            ref="formRef"
            :model="formData"
            :rules="formRules"
            label-width="120px"
          >
            <!-- 连接选择 -->
            <el-form-item label="目标连接" prop="connection_id">
              <el-select
                v-model="formData.connection_id"
                placeholder="请选择数据库连接"
                filterable
                @change="handleConnectionChange"
                style="width: 100%"
              >
                <el-option
                  v-for="connection in connections"
                  :key="connection.id"
                  :label="connection.name"
                  :value="connection.id"
                >
                  <div class="connection-option">
                    <span>{{ connection.name }}</span>
                    <el-tag :type="getEnvironmentTagType(connection.environment)" size="small">
                      {{ getEnvironmentLabel(connection.environment) }}
                    </el-tag>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>

            <!-- 数据库和表选择 -->
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="目标数据库" prop="database_name">
                  <el-select
                    v-model="formData.database_name"
                    placeholder="请选择数据库"
                    filterable
                    :loading="databasesLoading"
                    @change="handleDatabaseChange"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="database in databases"
                      :key="database"
                      :label="database"
                      :value="database"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              
              <el-col :span="12">
                <el-form-item label="目标表" prop="table_name">
                  <el-select
                    v-model="formData.table_name"
                    placeholder="请选择表"
                    filterable
                    :loading="tablesLoading"
                    @change="handleTableChange"
                    style="width: 100%"
                  >
                    <el-option
                      v-for="table in tables"
                      :key="table.table_name || table"
                      :label="table.table_name || table"
                      :value="table.table_name || table"
                    >
                      <div class="table-option">
                        <span>{{ table.table_name || table }}</span>
                        <span class="table-comment">{{ table.table_comment || '无注释' }}</span>
                      </div>
                    </el-option>
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>

            <!-- DDL编辑器 -->
            <el-form-item label="DDL操作" prop="ddl_content">
              <DDLEditor
                v-model="formData.original_ddl"
                v-model:ddl-type="formData.ddl_type"
                :table-name="formData.table_name"
                @change="handleDDLChange"
              />
            </el-form-item>

            <!-- 执行参数 -->
            <el-divider content-position="left"><span>执行参数</span></el-divider>
            <el-form-item label-width="0" class="params-form">
              <el-row :gutter="16">
                <el-col :span="12">
                  <el-form-item label="块大小" prop="chunk_size">
                    <el-input-number
                      v-model="formData.execution_params.chunk_size"
                      :min="100"
                      :max="10000"
                      placeholder="1000"
                      controls-position="right"
                      style="width: 100%"
                    />
                  </el-form-item>
                </el-col>
                
                <el-col :span="12">
                  <el-form-item label="最大负载">
                    <el-input
                      v-model="formData.execution_params.max_load"
                      placeholder="Threads_running=25"
                      style="width: 100%"
                    />
                  </el-form-item>
                </el-col>
                
                <el-col :span="12">
                  <el-form-item label="临界负载">
                    <el-input
                      v-model="formData.execution_params.critical_load"
                      placeholder="Threads_running=50"
                      style="width: 100%"
                    />
                  </el-form-item>
                </el-col>
                
                <el-col :span="12">
                  <el-form-item label="锁等待超时">
                    <el-input-number
                      v-model="formData.execution_params.lock_wait_timeout"
                      :min="1"
                      :max="300"
                      placeholder="60"
                      controls-position="right"
                      style="width: 100%"
                    />
                  </el-form-item>
                </el-col>

                <el-col :span="12">
                  <el-form-item label="跳过ALTER检查">
                    <el-switch
                      v-model="formData.execution_params.no_check_alter"
                      :active-text="'--no-check-alter'"
                      :inactive-text="'默认检查'"
                    />
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form-item>

            <!-- 操作按钮 -->
            <el-form-item>
              <el-button type="primary" @click="handlePreview">
                <el-icon><View /></el-icon>
                预览命令
              </el-button>
              <el-button
                type="success"
                @click="handleExecute"
                :disabled="!previewResult"
              >
                <el-icon><VideoPlay /></el-icon>
                执行DDL
              </el-button>
              <el-button @click="handleReset">
                <el-icon><RefreshLeft /></el-icon>
                重置
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 预览结果 -->
          <div v-if="previewResult" class="preview-section">
            <el-divider content-position="left">
              <span>命令预览</span>
            </el-divider>
            <pre style="background:#f7f7f7;padding:12px;border-radius:6px;white-space:pre-wrap">{{ previewResult.command }}</pre>
            <div style="margin-top:8px;color:#909399;font-size:12px">
              <span v-if="formData.execution_params.no_check_alter">已启用 --no-check-alter</span>
              <span v-else>未启用 --no-check-alter</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：执行日志（仅显示日志，不展示前端“监控任务”概念） -->
      <el-col :span="8">
        <el-card class="running-tasks-card">
          <template #header>
            <div class="card-header">
              <h3>执行日志</h3>
            </div>
          </template>

          <div v-if="!currentExecutionId" class="empty-tasks">
            <el-empty description="暂无日志可展示" :image-size="80" />
          </div>
          <div v-else>
            <ExecutionMonitor :execution-id="currentExecutionId" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  Operation,
  Refresh,
  Loading,
  DataAnalysis,
  CircleCheck,
  Timer,
  View,
  VideoPlay,
  RefreshLeft
} from '@element-plus/icons-vue'
import DDLEditor from '@/components/forms/DDLEditor.vue'
import type { CreateExecutionRequest, DDLType } from '@/types/execution'
import ConnectionService from '@/services/connection'
import ExecutionService from '@/services/execution'
import ExecutionMonitor from '@/components/execution/ExecutionMonitor.vue'

// 响应式数据
const formRef = ref<FormInstance>()
const connections = ref<any[]>([])
const databases = ref<string[]>([])
const tables = ref<any[]>([])
const databasesLoading = ref(false)
const tablesLoading = ref(false)
const previewResult = ref<any>(null)
const currentExecutionId = ref<string>('')

// 表单数据
const formData = reactive<CreateExecutionRequest & { original_ddl: string }>({
  connection_id: '',
  database_name: '',
  table_name: '',
  ddl_type: 'other',
  original_ddl: '',
  execution_params: {
    chunk_size: 3000,
    max_load: 'Threads_running=8000',
    critical_load: 'Threads_running=10000',
    charset: 'utf8mb4',
    lock_wait_timeout: 60,
    other_params: '',
    no_check_alter: false
  }
})

// 表单验证规则
const formRules: FormRules = {
  connection_id: [
    { required: true, message: '请选择数据库连接', trigger: 'change' }
  ],
  database_name: [
    { required: true, message: '请选择目标数据库', trigger: 'change' }
  ],
  table_name: [
    { required: true, message: '请选择目标表', trigger: 'change' }
  ]
}

// 方法
const handleConnectionChange = async () => {
  formData.database_name = ''
  formData.table_name = ''
  databases.value = []
  tables.value = []
  
  if (formData.connection_id) {
    try {
      databasesLoading.value = true
      databases.value = await ConnectionService.getDatabases(formData.connection_id)
    } catch (e) {
      ElMessage.error('加载数据库列表失败')
    } finally {
      databasesLoading.value = false
    }
  }
}

const handleDatabaseChange = async () => {
  formData.table_name = ''
  tables.value = []
  
  if (formData.database_name) {
    try {
      tablesLoading.value = true
      tables.value = await ConnectionService.getTables(formData.connection_id, formData.database_name)
    } catch (e) {
      ElMessage.error('加载表列表失败')
    } finally {
      tablesLoading.value = false
    }
  }
}

const handleTableChange = () => {
  // 表选择变化时的处理
}

const handleDDLChange = (data: { ddl: string, type: DDLType }) => {
  formData.original_ddl = data.ddl
  formData.ddl_type = data.type
  
  // 清空之前的预览结果
  previewResult.value = null
}

const handlePreview = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    // 后端预览接口仅支持 fragment 与 custom，其它类型统一映射为 custom
    const ddlTypeForPreview = formData.ddl_type === 'fragment' ? 'fragment' : 'custom'
    const res = await ExecutionService.previewCommand({
      connection_id: formData.connection_id,
      database_name: formData.database_name,
      table_name: formData.table_name,
      ddl_type: ddlTypeForPreview,
      original_ddl: formData.original_ddl,
      execution_params: formData.execution_params
    })
    previewResult.value = res
  } catch (error) {
    console.error('预览失败:', error)
  }
}

const handleExecute = async () => {
  if (!previewResult.value) {
    ElMessage.warning('请先预览命令')
    return
  }
  
  try {
    const created = await ExecutionService.createExecution({
      connection_id: formData.connection_id,
      database_name: formData.database_name,
      table_name: formData.table_name,
      ddl_type: formData.ddl_type as DDLType,
      original_ddl: formData.original_ddl,
      execution_params: formData.execution_params
    })
    await ExecutionService.startExecution(created.id)
    currentExecutionId.value = created.id
    ElMessage.success('执行已启动')
  } catch (e) {
    ElMessage.error('执行启动失败')
  }
}

const handleReset = () => {
  formRef.value?.resetFields()
  Object.assign(formData, {
    connection_id: '',
    database_name: '',
    table_name: '',
    ddl_type: 'other',
    original_ddl: '',
    execution_params: {
      chunk_size: 3000,
      max_load: 'Threads_running=8000',
      critical_load: 'Threads_running=10000',
      charset: 'utf8mb4',
      lock_wait_timeout: 60,
      other_params: ''
    }
  })
  databases.value = []
  tables.value = []
  previewResult.value = null
}

const handleClearPreview = () => {
  previewResult.value = null
}

const handleRefreshTasks = async () => {
  // TODO: 刷新运行中任务
  ElMessage.info('刷新功能开发中...')
}

// 工具方法
const getEnvironmentLabel = (environment: string) => {
  const labels: Record<string, string> = {
    prod: '生产',
    test: '测试',
    dev: '开发'
  }
  return labels[environment] || environment
}

const getEnvironmentTagType = (environment: string) => {
  const types: Record<string, string> = {
    prod: 'danger',
    test: 'warning',
    dev: 'success'
  }
  return types[environment] || 'info'
}

// 生命周期
onMounted(async () => {
  try {
    connections.value = await ConnectionService.getConnections()
  } catch (e) {
    ElMessage.error('加载连接列表失败')
  }
})
</script>

<style scoped>
.execution-page {
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

.stats-card.success {
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

.stats-icon.running {
  color: #faad14;
}

.execution-form-card,
.running-tasks-card {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.connection-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.table-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.table-comment {
  font-size: 12px;
  color: #909399;
}

.preview-section {
  margin-top: 24px;
}

.empty-tasks {
  padding: 20px;
  text-align: center;
}

/* 参数区交互与可读性优化 */
.params-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.params-form :deep(.el-input),
.params-form :deep(.el-input__inner),
.params-form :deep(.el-input-number),
.params-form :deep(.el-input-number .el-input__inner) {
  height: 36px;
  line-height: 36px;
  font-size: 14px;
}

@media (max-width: 1200px) {
  .params-form :deep(.el-col-12) {
    width: 100% !important;
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  .execution-page {
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
}
</style>