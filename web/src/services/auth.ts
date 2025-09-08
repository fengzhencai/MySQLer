import api from './api'
import type { LoginRequest, LoginResponse, User, ApiResponse } from '@/types/auth'

export const authApi = {
  // 用户登录
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await api.post<ApiResponse<LoginResponse>>('/auth/login', data)
    return response.data.data!
  },

  // 用户登出
  async logout(): Promise<void> {
    await api.post<ApiResponse>('/auth/logout')
  },

  // 获取用户信息
  async getProfile(): Promise<User> {
    const response = await api.get<ApiResponse<User>>('/auth/profile')
    return response.data.data!
  },
}