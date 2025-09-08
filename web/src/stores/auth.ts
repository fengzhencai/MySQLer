import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, LoginRequest, LoginResponse } from '@/types/auth'
import { authApi } from '@/services/auth'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string>('')
  const loading = ref(false)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const canExecuteDDL = computed(() => 
    user.value?.role === 'admin' || user.value?.role === 'operator'
  )

  // 初始化 - 从localStorage恢复状态
  function init() {
    const savedToken = localStorage.getItem('auth_token')
    const savedUser = localStorage.getItem('auth_user')
    
    if (savedToken && savedUser) {
      token.value = savedToken
      try {
        user.value = JSON.parse(savedUser)
      } catch (error) {
        console.error('Failed to parse saved user data:', error)
        logout()
      }
    }
  }

  // 登录
  async function login(loginData: LoginRequest): Promise<void> {
    loading.value = true
    try {
      const response = await authApi.login(loginData)
      
      token.value = response.token
      user.value = response.user
      
      // 保存到localStorage
      localStorage.setItem('auth_token', response.token)
      localStorage.setItem('auth_user', JSON.stringify(response.user))
      
    } catch (error) {
      throw error
    } finally {
      loading.value = false
    }
  }

  // 登出
  async function logout(): Promise<void> {
    try {
      // 调用后端登出接口
      if (token.value) {
        await authApi.logout()
      }
    } catch (error) {
      console.error('Logout API call failed:', error)
    } finally {
      // 清除本地状态
      user.value = null
      token.value = ''
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_user')
    }
  }

  // 获取用户信息
  async function fetchProfile(): Promise<void> {
    try {
      const profile = await authApi.getProfile()
      user.value = profile
      localStorage.setItem('auth_user', JSON.stringify(profile))
    } catch (error) {
      console.error('Failed to fetch user profile:', error)
      logout()
    }
  }

  // 检查权限
  function hasPermission(permission: string): boolean {
    if (!user.value) return false
    
    switch (user.value.role) {
      case 'admin':
        return true
      case 'operator':
        return permission !== 'manage_users'
      case 'viewer':
        return permission === 'view'
      default:
        return false
    }
  }

  return {
    // 状态
    user,
    token,
    loading,
    
    // 计算属性
    isAuthenticated,
    isAdmin,
    canExecuteDDL,
    
    // 方法
    init,
    login,
    logout,
    fetchProfile,
    hasPermission
  }
})