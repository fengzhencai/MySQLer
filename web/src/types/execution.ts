// DDL操作类型
export type DDLType = 'fragment' | 'add_column' | 'modify_column' | 'drop_column' | 'add_index' | 'drop_index' | 'other'

// 执行状态类型
export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

// 执行参数类型
export interface ExecutionParams {
  chunk_size: number
  max_load: string
  critical_load: string
  charset: string
  lock_wait_timeout: number
  other_params?: string
}

// 执行记录类型
export interface ExecutionRecord {
  id: string
  connection_id: string
  table_name: string
  database_name: string
  ddl_type?: DDLType
  original_ddl?: string
  generated_command: string
  execution_params?: ExecutionParams
  status: ExecutionStatus
  start_time?: string
  end_time?: string
  duration_seconds?: number
  processed_rows: number
  total_rows: number
  avg_speed?: number
  container_id?: string
  execution_logs?: string
  error_message?: string
  created_by: string
  created_at: string
  updated_at: string
  connection?: {
    id: string
    name: string
    environment: string
    host: string
    database_name: string
  }
}

// 创建执行请求类型
export interface CreateExecutionRequest {
  connection_id: string
  table_name: string
  database_name: string
  ddl_type?: DDLType
  original_ddl?: string
  execution_params?: ExecutionParams
}

// WebSocket消息类型
export interface WebSocketMessage {
  type: 'progress' | 'log' | 'status' | 'error'
  data: {
    execution_id: string
    status?: ExecutionStatus
    progress?: number
    processed_rows?: number
    total_rows?: number
    current_speed?: number
    current_stage?: string
    log_line?: string
    error_message?: string
  }
}