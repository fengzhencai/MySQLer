import api from './api'
import type { ExecutionRecord, CreateExecutionRequest, ExecutionParams } from '@/types/execution'
import type { ApiResponse } from '@/types/auth'

// 预览命令请求类型
export interface PreviewCommandRequest {
  connection_id: string
  table_name: string
  database_name: string
  ddl_type?: string
  original_ddl?: string
  execution_params?: ExecutionParams
}

// 预览命令响应类型
export interface PreviewCommandResponse {
  command: string
  risk_analysis?: any
  table_info?: any
  estimated_time?: string
  recommended_chunk_size?: number
  no_check_alter?: boolean
}

// 执行状态响应类型
export interface ExecutionStatusResponse {
  execution_id: string
  status: string
  progress: number
  processed_rows: number
  total_rows: number
  current_speed: number
  current_stage: string
  error_message?: string
  start_time?: string
  estimated_remaining?: string
}

// 执行管理API服务
export class ExecutionService {
  // 获取执行历史列表
  static async getExecutions(params?: {
    page?: number
    size?: number
    status?: string
    connection_id?: string
    start_date?: string
    end_date?: string
  }): Promise<{ records: ExecutionRecord[], total: number }> {
    const response = await api.get<ApiResponse<{ records: ExecutionRecord[], total: number }>>('/executions', { params })
    return response.data.data || { records: [], total: 0 }
  }

  // 根据ID获取执行详情
  static async getExecution(id: string): Promise<ExecutionRecord> {
    const response = await api.get<ApiResponse<ExecutionRecord>>(`/executions/${id}`)
    return response.data.data
  }

  // 预览PT命令
  static async previewCommand(data: PreviewCommandRequest): Promise<PreviewCommandResponse> {
    const response = await api.post<ApiResponse<PreviewCommandResponse>>('/executions/preview', data)
    return response.data.data
  }

  // 创建并启动执行
  static async createExecution(data: CreateExecutionRequest): Promise<ExecutionRecord> {
    const response = await api.post<ApiResponse<ExecutionRecord>>('/executions', data)
    return response.data.data
  }

  // 启动执行
  static async startExecution(executionId: string): Promise<void> {
    await api.post(`/executions/${executionId}/start`)
  }

  // 停止执行
  static async stopExecution(executionId: string): Promise<void> {
    await api.post(`/executions/${executionId}/stop`)
  }

  // 取消执行
  static async cancelExecution(executionId: string): Promise<void> {
    await api.post(`/executions/${executionId}/cancel`)
  }

  // 获取执行状态
  static async getExecutionStatus(executionId: string): Promise<ExecutionStatusResponse> {
    const response = await api.get<ApiResponse<ExecutionStatusResponse>>(`/executions/${executionId}/status`)
    return response.data.data
  }

  // 获取正在运行的任务
  static async getRunningTasks(): Promise<ExecutionStatusResponse[]> {
    const response = await api.get<ApiResponse<ExecutionStatusResponse[]>>('/executions/running')
    return response.data.data || []
  }

  // 获取执行日志
  static async getExecutionLogs(executionId: string): Promise<string[]> {
    const response = await api.get<ApiResponse<string[]>>(`/executions/${executionId}/logs`)
    return response.data.data || []
  }

  // 删除执行记录
  static async deleteExecution(executionId: string): Promise<void> {
    await api.delete(`/executions/${executionId}`)
  }

  // 重新执行
  static async reExecute(executionId: string): Promise<ExecutionRecord> {
    const response = await api.post<ApiResponse<ExecutionRecord>>(`/executions/${executionId}/retry`)
    return response.data.data
  }

  // 获取执行统计
  static async getExecutionStats(params?: {
    start_date?: string
    end_date?: string
    connection_id?: string
  }): Promise<{
    total_executions: number
    success_rate: number
    avg_duration: number
    status_distribution: Record<string, number>
  }> {
    const response = await api.get('/executions/stats', { params })
    return response.data.data || {
      total_executions: 0,
      success_rate: 0,
      avg_duration: 0,
      status_distribution: {}
    }
  }
}

export default ExecutionService