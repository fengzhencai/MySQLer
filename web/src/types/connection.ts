// 环境类型
export type Environment = 'prod' | 'test' | 'dev'

// 连接信息类型
export interface Connection {
  id: string
  name: string
  environment: Environment
  host: string
  port: number
  username: string
  database_name: string
  description?: string
  connect_timeout: number
  charset: string
  use_ssl: boolean
  created_by: string
  created_at: string
  updated_at: string
}

// 创建/更新连接请求类型
export interface CreateConnectionRequest {
  name: string
  environment: Environment
  host: string
  port?: number
  username: string
  password: string
  database_name: string
  description?: string
  connect_timeout?: number
  charset?: string
  use_ssl?: boolean
}

// 连接测试结果类型
export interface ConnectionTestResult {
  success: boolean
  message: string
  latency?: number
  server_version?: string
}