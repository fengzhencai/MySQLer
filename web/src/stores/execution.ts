import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ExecutionService, { type PreviewCommandRequest, type PreviewCommandResponse, type ExecutionStatusResponse } from '@/services/execution'
import type { ExecutionRecord, CreateExecutionRequest, ExecutionStatus } from '@/types/execution'

export const useExecutionStore = defineStore('execution', () => {
  // 状态
  const executions = ref<ExecutionRecord[]>([])
  const currentExecution = ref<ExecutionRecord | null>(null)
  const runningTasks = ref<ExecutionStatusResponse[]>([])
  const loading = ref(false)
  const previewLoading = ref(false)
  const executionStats = ref({
    total_executions: 0,
    success_rate: 0,
    avg_duration: 0,
    status_distribution: {} as Record<string, number>
  })

  // 分页相关
  const currentPage = ref(1)
  const pageSize = ref(20)
  const total = ref(0)

  // 计算属性
  const executionsByStatus = computed(() => {
    const groups: Record<ExecutionStatus, ExecutionRecord[]> = {
      pending: [],
      running: [],
      completed: [],
      failed: [],
      cancelled: []
    }
    
    executions.value.forEach(execution => {
      groups[execution.status].push(execution)
    })
    
    return groups
  })

  const runningCount = computed(() => runningTasks.value.length)
  
  const successRate = computed(() => {
    const total = executions.value.length
    if (total === 0) return 0
    const success = executions.value.filter(e => e.status === 'completed').length
    return Math.round((success / total) * 100)
  })

  // 方法
  const fetchExecutions = async (params?: {
    page?: number
    size?: number
    status?: string
    connection_id?: string
    start_date?: string
    end_date?: string
  }) => {
    try {
      loading.value = true
      const result = await ExecutionService.getExecutions({
        page: currentPage.value,
        size: pageSize.value,
        ...params
      })
      executions.value = result.records
      total.value = result.total
    } catch (error) {
      console.error('获取执行列表失败:', error)
      ElMessage.error('获取执行列表失败')
    } finally {
      loading.value = false
    }
  }

  const getExecution = async (id: string) => {
    try {
      const execution = await ExecutionService.getExecution(id)
      currentExecution.value = execution
      return execution
    } catch (error) {
      console.error('获取执行详情失败:', error)
      ElMessage.error('获取执行详情失败')
      throw error
    }
  }

  const previewCommand = async (data: PreviewCommandRequest): Promise<PreviewCommandResponse | null> => {
    try {
      previewLoading.value = true
      const result = await ExecutionService.previewCommand(data)
      return result
    } catch (error) {
      console.error('预览命令失败:', error)
      ElMessage.error('预览命令失败')
      return null
    } finally {
      previewLoading.value = false
    }
  }

  const createExecution = async (data: CreateExecutionRequest) => {
    try {
      loading.value = true
      const newExecution = await ExecutionService.createExecution(data)
      
      // 添加到列表开头
      executions.value.unshift(newExecution)
      total.value++
      
      ElMessage.success('执行任务创建成功')
      return newExecution
    } catch (error) {
      console.error('创建执行失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const startExecution = async (executionId: string) => {
    try {
      await ExecutionService.startExecution(executionId)
      
      // 更新本地状态
      const execution = executions.value.find(e => e.id === executionId)
      if (execution) {
        execution.status = 'running'
        execution.start_time = new Date().toISOString()
      }
      
      ElMessage.success('执行已启动')
      await fetchRunningTasks()
    } catch (error) {
      console.error('启动执行失败:', error)
      ElMessage.error('启动执行失败')
    }
  }

  const stopExecution = async (executionId: string) => {
    try {
      await ElMessageBox.confirm(
        '确定要停止这个执行任务吗？',
        '确认停止',
        {
          type: 'warning',
          confirmButtonText: '确定停止',
          cancelButtonText: '取消'
        }
      )

      await ExecutionService.stopExecution(executionId)
      
      // 更新本地状态
      const execution = executions.value.find(e => e.id === executionId)
      if (execution) {
        execution.status = 'cancelled'
        execution.end_time = new Date().toISOString()
      }
      
      ElMessage.success('执行已停止')
      await fetchRunningTasks()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('停止执行失败:', error)
        ElMessage.error('停止执行失败')
      }
    }
  }

  const cancelExecution = async (executionId: string) => {
    try {
      await ElMessageBox.confirm(
        '确定要取消这个执行任务吗？',
        '确认取消',
        {
          type: 'warning',
          confirmButtonText: '确定取消',
          cancelButtonText: '取消'
        }
      )

      await ExecutionService.cancelExecution(executionId)
      
      // 更新本地状态
      const execution = executions.value.find(e => e.id === executionId)
      if (execution) {
        execution.status = 'cancelled'
        execution.end_time = new Date().toISOString()
      }
      
      ElMessage.success('执行已取消')
      await fetchRunningTasks()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('取消执行失败:', error)
        ElMessage.error('取消执行失败')
      }
    }
  }

  const getExecutionStatus = async (executionId: string) => {
    try {
      const status = await ExecutionService.getExecutionStatus(executionId)
      
      // 更新本地状态
      const execution = executions.value.find(e => e.id === executionId)
      if (execution) {
        execution.status = status.status as ExecutionStatus
        execution.processed_rows = status.processed_rows
        execution.total_rows = status.total_rows
      }
      
      return status
    } catch (error) {
      console.error('获取执行状态失败:', error)
      return null
    }
  }

  const fetchRunningTasks = async () => {
    try {
      runningTasks.value = await ExecutionService.getRunningTasks()
    } catch (error) {
      console.error('获取运行中任务失败:', error)
    }
  }

  const getExecutionLogs = async (executionId: string) => {
    try {
      return await ExecutionService.getExecutionLogs(executionId)
    } catch (error) {
      console.error('获取执行日志失败:', error)
      ElMessage.error('获取执行日志失败')
      return []
    }
  }

  const deleteExecution = async (executionId: string) => {
    try {
      await ElMessageBox.confirm(
        '确定要删除这个执行记录吗？删除后不可恢复。',
        '确认删除',
        {
          type: 'warning',
          confirmButtonText: '确定删除',
          cancelButtonText: '取消'
        }
      )

      await ExecutionService.deleteExecution(executionId)
      
      // 从本地状态中移除
      executions.value = executions.value.filter(e => e.id !== executionId)
      total.value--
      
      if (currentExecution.value?.id === executionId) {
        currentExecution.value = null
      }
      
      ElMessage.success('执行记录删除成功')
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除执行记录失败:', error)
        ElMessage.error('删除执行记录失败')
      }
    }
  }

  const reExecute = async (executionId: string) => {
    try {
      await ElMessageBox.confirm(
        '确定要重新执行这个任务吗？',
        '确认重新执行',
        {
          type: 'info',
          confirmButtonText: '确定重新执行',
          cancelButtonText: '取消'
        }
      )

      loading.value = true
      const newExecution = await ExecutionService.reExecute(executionId)
      
      // 添加到列表开头
      executions.value.unshift(newExecution)
      total.value++
      
      ElMessage.success('重新执行任务已创建')
      return newExecution
    } catch (error) {
      if (error !== 'cancel') {
        console.error('重新执行失败:', error)
        ElMessage.error('重新执行失败')
      }
    } finally {
      loading.value = false
    }
  }

  const fetchExecutionStats = async (params?: {
    start_date?: string
    end_date?: string
    connection_id?: string
  }) => {
    try {
      executionStats.value = await ExecutionService.getExecutionStats(params)
    } catch (error) {
      console.error('获取执行统计失败:', error)
    }
  }

  // 搜索执行记录
  const searchExecutions = (keyword: string) => {
    if (!keyword.trim()) {
      return executions.value
    }
    
    const lowerKeyword = keyword.toLowerCase()
    return executions.value.filter(execution => 
      execution.table_name.toLowerCase().includes(lowerKeyword) ||
      execution.database_name.toLowerCase().includes(lowerKeyword) ||
      execution.connection?.name.toLowerCase().includes(lowerKeyword) ||
      execution.id.toLowerCase().includes(lowerKeyword)
    )
  }

  // 按状态过滤
  const filterByStatus = (status: ExecutionStatus | 'all') => {
    if (status === 'all') {
      return executions.value
    }
    return executions.value.filter(execution => execution.status === status)
  }

  // 分页方法
  const setPage = (page: number) => {
    currentPage.value = page
  }

  const setPageSize = (size: number) => {
    pageSize.value = size
    currentPage.value = 1
  }

  // 重置状态
  const resetState = () => {
    executions.value = []
    currentExecution.value = null
    runningTasks.value = []
    loading.value = false
    previewLoading.value = false
    currentPage.value = 1
    total.value = 0
  }

  return {
    // 状态
    executions,
    currentExecution,
    runningTasks,
    loading,
    previewLoading,
    executionStats,
    currentPage,
    pageSize,
    total,
    
    // 计算属性
    executionsByStatus,
    runningCount,
    successRate,
    
    // 方法
    fetchExecutions,
    getExecution,
    previewCommand,
    createExecution,
    startExecution,
    stopExecution,
    cancelExecution,
    getExecutionStatus,
    fetchRunningTasks,
    getExecutionLogs,
    deleteExecution,
    reExecute,
    fetchExecutionStats,
    searchExecutions,
    filterByStatus,
    setPage,
    setPageSize,
    resetState
  }
})