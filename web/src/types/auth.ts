// 用户角色类型
export type UserRole = 'admin' | 'operator' | 'viewer'

// 用户信息类型
export interface User {
  id: string
  username: string
  display_name?: string
  email?: string
  role: UserRole
  is_active: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

// 登录请求类型
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应类型
export interface LoginResponse {
  token: string
  expires_in: number
  user: User
}

// API响应基础类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
  timestamp?: string
}