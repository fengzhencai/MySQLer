import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ConnectionService from '@/services/connection'
import type { Connection, CreateConnectionRequest, Environment } from '@/types/connection'

export const useConnectionStore = defineStore('connection', () => {
  // 状态
  const connections = ref<Connection[]>([])
  const loading = ref(false)
  const currentConnection = ref<Connection | null>(null)

  // 计算属性
  const connectionsByEnvironment = computed(() => {
    const groups: Record<Environment, Connection[]> = {
      prod: [],
      test: [],
      dev: []
    }
    
    connections.value.forEach(conn => {
      groups[conn.environment].push(conn)
    })
    
    return groups
  })

  const connectionCount = computed(() => connections.value.length)

  // 按环境统计
  const environmentStats = computed(() => ({
    prod: connections.value.filter(c => c.environment === 'prod').length,
    test: connections.value.filter(c => c.environment === 'test').length,
    dev: connections.value.filter(c => c.environment === 'dev').length
  }))

  // 方法
  const fetchConnections = async () => {
    try {
      loading.value = true
      connections.value = await ConnectionService.getConnections()
    } catch (error) {
      console.error('获取连接列表失败:', error)
      ElMessage.error('获取连接列表失败')
    } finally {
      loading.value = false
    }
  }

  const getConnection = async (id: string) => {
    try {
      const connection = await ConnectionService.getConnection(id)
      currentConnection.value = connection
      return connection
    } catch (error) {
      console.error('获取连接详情失败:', error)
      ElMessage.error('获取连接详情失败')
      throw error
    }
  }

  const createConnection = async (data: CreateConnectionRequest) => {
    try {
      loading.value = true
      const newConnection = await ConnectionService.createConnection(data)
      connections.value.push(newConnection)
      ElMessage.success('连接创建成功')
      return newConnection
    } catch (error) {
      console.error('创建连接失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const updateConnection = async (id: string, data: CreateConnectionRequest) => {
    try {
      loading.value = true
      const updatedConnection = await ConnectionService.updateConnection(id, data)
      
      // 更新本地状态
      const index = connections.value.findIndex(c => c.id === id)
      if (index !== -1) {
        connections.value[index] = updatedConnection
      }
      
      if (currentConnection.value?.id === id) {
        currentConnection.value = updatedConnection
      }
      
      ElMessage.success('连接更新成功')
      return updatedConnection
    } catch (error) {
      console.error('更新连接失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const deleteConnection = async (id: string) => {
    try {
      await ElMessageBox.confirm(
        '确定要删除这个连接吗？删除后不可恢复。',
        '确认删除',
        {
          type: 'warning',
          confirmButtonText: '确定删除',
          cancelButtonText: '取消'
        }
      )

      await ConnectionService.deleteConnection(id)
      
      // 从本地状态中移除
      connections.value = connections.value.filter(c => c.id !== id)
      
      if (currentConnection.value?.id === id) {
        currentConnection.value = null
      }
      
      ElMessage.success('连接删除成功')
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除连接失败:', error)
        ElMessage.error('删除连接失败')
      }
    }
  }

  const testConnection = async (id: string) => {
    try {
      const result = await ConnectionService.testConnection(id)
      
      if (result.success) {
        ElMessage.success(`连接测试成功`)
      } else {
        ElMessage.error(`连接测试失败: ${result.message}`)
      }
      
      return result
    } catch (error) {
      console.error('连接测试失败:', error)
      ElMessage.error('连接测试失败')
      throw error
    }
  }

  // 表单参数测试连接（不保存）
  const testConnectionByParams = async (data: CreateConnectionRequest) => {
    try {
      const result = await ConnectionService.testConnectionByParams(data)
      if (result.success) {
        ElMessage.success('连接测试成功')
      } else {
        ElMessage.error(`连接测试失败: ${result.error || result.message || ''}`)
      }
      return result
    } catch (error) {
      console.error('连接测试失败:', error)
      ElMessage.error('连接测试失败')
      throw error
    }
  }

  const getDatabases = async (id: string) => {
    try {
      return await ConnectionService.getDatabases(id)
    } catch (error) {
      console.error('获取数据库列表失败:', error)
      ElMessage.error('获取数据库列表失败')
      return []
    }
  }

  const getTables = async (id: string, database: string) => {
    try {
      return await ConnectionService.getTables(id, database)
    } catch (error) {
      console.error('获取表列表失败:', error)
      ElMessage.error('获取表列表失败')
      return []
    }
  }

  // 搜索连接
  const searchConnections = (keyword: string) => {
    if (!keyword.trim()) {
      return connections.value
    }
    
    const lowerKeyword = keyword.toLowerCase()
    return connections.value.filter(conn => 
      conn.name.toLowerCase().includes(lowerKeyword) ||
      conn.host.toLowerCase().includes(lowerKeyword) ||
      conn.database_name.toLowerCase().includes(lowerKeyword) ||
      (conn.description && conn.description.toLowerCase().includes(lowerKeyword))
    )
  }

  // 按环境过滤
  const filterByEnvironment = (env: Environment | 'all') => {
    if (env === 'all') {
      return connections.value
    }
    return connections.value.filter(conn => conn.environment === env)
  }

  // 重置状态
  const resetState = () => {
    connections.value = []
    currentConnection.value = null
    loading.value = false
  }

  return {
    // 状态
    connections,
    loading,
    currentConnection,
    
    // 计算属性
    connectionsByEnvironment,
    connectionCount,
    environmentStats,
    
    // 方法
    fetchConnections,
    getConnection,
    createConnection,
    updateConnection,
    deleteConnection,
    testConnection,
    getDatabases,
    getTables,
    testConnectionByParams,
    searchConnections,
    filterByEnvironment,
    resetState
  }
})