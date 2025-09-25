import api from './api'
import type { Connection, CreateConnectionRequest, ConnectionTestResult } from '@/types/connection'
import type { ApiResponse } from '@/types/auth'

// 连接管理API服务
export class ConnectionService {
  // 获取连接列表
  static async getConnections(): Promise<Connection[]> {
    const response = await api.get<ApiResponse<Connection[]>>('/connections')
    return response.data.data || []
  }

  // 根据ID获取连接详情
  static async getConnection(id: string): Promise<Connection> {
    const response = await api.get<ApiResponse<Connection>>(`/connections/${id}`)
    return response.data.data
  }

  // 创建连接
  static async createConnection(data: CreateConnectionRequest): Promise<Connection> {
    const response = await api.post<ApiResponse<Connection>>('/connections', data)
    return response.data.data
  }

  // 更新连接
  static async updateConnection(id: string, data: CreateConnectionRequest): Promise<Connection> {
    const response = await api.put<ApiResponse<Connection>>(`/connections/${id}`, data)
    return response.data.data
  }

  // 删除连接
  static async deleteConnection(id: string): Promise<void> {
    await api.delete(`/connections/${id}`)
  }

  // 测试连接
  static async testConnection(id: string): Promise<ConnectionTestResult> {
    const response = await api.post<ApiResponse<ConnectionTestResult>>(`/connections/${id}/test`)
    return response.data.data
  }

  // 获取数据库列表
  static async getDatabases(id: string): Promise<string[]> {
    const response = await api.get<ApiResponse<string[]>>(`/tools/connections/${id}/databases`)
    return response.data.data || []
  }

  // 获取表列表
  static async getTables(id: string, database: string): Promise<any[]> {
    const response = await api.get<ApiResponse<any[]>>(`/tools/connections/${id}/databases/${database}/tables`)
    return response.data.data || []
  }
}

export default ConnectionService