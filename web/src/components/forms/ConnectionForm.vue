<template>
  <el-form
    ref="formRef"
    :model="formData"
    :rules="formRules"
    label-width="120px"
    class="connection-form"
  >
    <el-row :gutter="20">
      <el-col :span="12">
        <el-form-item label="连接名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入连接名称"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
      </el-col>
      
      <el-col :span="12">
        <el-form-item label="环境类型" prop="environment">
          <el-select v-model="formData.environment" placeholder="请选择环境">
            <el-option label="生产环境" value="prod">
              <span style="float: left">生产环境</span>
              <span style="float: right; color: #ff4d4f; font-size: 13px">PROD</span>
            </el-option>
            <el-option label="测试环境" value="test">
              <span style="float: left">测试环境</span>
              <span style="float: right; color: #faad14; font-size: 13px">TEST</span>
            </el-option>
            <el-option label="开发环境" value="dev">
              <span style="float: left">开发环境</span>
              <span style="float: right; color: #52c41a; font-size: 13px">DEV</span>
            </el-option>
          </el-select>
        </el-form-item>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="16">
        <el-form-item label="主机地址" prop="host">
          <el-input
            v-model="formData.host"
            placeholder="请输入主机地址或IP"
          />
        </el-form-item>
      </el-col>
      
      <el-col :span="8">
        <el-form-item label="端口" prop="port">
          <el-input-number
            v-model="formData.port"
            :min="1"
            :max="65535"
            placeholder="3306"
            style="width: 100%"
          />
        </el-form-item>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="formData.username"
            placeholder="请输入数据库用户名"
          />
        </el-form-item>
      </el-col>
      
      <el-col :span="12">
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="formData.password"
            type="password"
            placeholder="请输入数据库密码"
            show-password
          />
        </el-form-item>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="24">
        <el-form-item label="数据库名" prop="database_name">
          <el-input
            v-model="formData.database_name"
            placeholder="请输入数据库名称"
          />
        </el-form-item>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="24">
        <el-form-item label="描述信息">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入连接描述（可选）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-col>
    </el-row>

    <!-- 高级配置 -->
    <el-divider content-position="left">
      <span style="color: #666; font-size: 14px">高级配置</span>
    </el-divider>

    <el-row :gutter="20">
      <el-col :span="8">
        <el-form-item label="连接超时" prop="connect_timeout">
          <el-input-number
            v-model="formData.connect_timeout"
            :min="1"
            :max="60"
            placeholder="5"
            style="width: 100%"
          />
          <div class="form-tip">单位：秒</div>
        </el-form-item>
      </el-col>
      
      <el-col :span="8">
        <el-form-item label="字符集" prop="charset">
          <el-select v-model="formData.charset" placeholder="请选择字符集">
            <el-option label="utf8mb4" value="utf8mb4" />
            <el-option label="utf8" value="utf8" />
            <el-option label="latin1" value="latin1" />
            <el-option label="gbk" value="gbk" />
          </el-select>
        </el-form-item>
      </el-col>
      
      <el-col :span="8">
        <el-form-item label="SSL连接">
          <el-switch
            v-model="formData.use_ssl"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-col>
    </el-row>

    <!-- 操作按钮 -->
    <el-form-item>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ mode === 'create' ? '创建连接' : '更新连接' }}
      </el-button>
      <el-button @click="handleTest" :loading="testing">
        测试连接
      </el-button>
      <el-button @click="handleReset">
        重置
      </el-button>
      <el-button @click="handleCancel">
        取消
      </el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import type { Connection, CreateConnectionRequest, Environment } from '@/types/connection'

// Props
interface Props {
  modelValue?: Connection | null
  mode?: 'create' | 'edit'
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: null,
  mode: 'create'
})

// Emits
interface Emits {
  (e: 'update:modelValue', value: Connection | null): void
  (e: 'submit', data: CreateConnectionRequest): void
  (e: 'test', data: CreateConnectionRequest): void
  (e: 'cancel'): void
}

const emit = defineEmits<Emits>()

// 响应式数据
const formRef = ref<FormInstance>()
const submitting = ref(false)
const testing = ref(false)

// 表单数据
const formData = reactive<CreateConnectionRequest>({
  name: '',
  environment: 'test',
  host: '',
  port: 3306,
  username: '',
  password: '',
  database_name: '',
  description: '',
  connect_timeout: 5,
  charset: 'utf8mb4',
  use_ssl: false
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入连接名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  environment: [
    { required: true, message: '请选择环境类型', trigger: 'change' }
  ],
  host: [
    { required: true, message: '请输入主机地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'change' },
    { type: 'number', min: 1, max: 65535, message: '端口号范围 1-65535', trigger: 'change' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  database_name: [
    { required: true, message: '请输入数据库名称', trigger: 'blur' }
  ],
  connect_timeout: [
    { type: 'number', min: 1, max: 60, message: '连接超时范围 1-60 秒', trigger: 'change' }
  ]
}

// 监听props变化
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    // 编辑模式，填充表单数据
    Object.assign(formData, {
      name: newValue.name,
      environment: newValue.environment,
      host: newValue.host,
      port: newValue.port,
      username: newValue.username,
      password: '', // 密码不回显
      database_name: newValue.database_name,
      description: newValue.description || '',
      connect_timeout: newValue.connect_timeout,
      charset: newValue.charset,
      use_ssl: newValue.use_ssl
    })
  }
}, { immediate: true })

// 处理提交
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    submitting.value = true
    emit('submit', { ...formData })
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    submitting.value = false
  }
}

// 处理测试连接
const handleTest = async () => {
  if (!formRef.value) return
  
  try {
    // 验证必要字段
    await formRef.value.validateField(['host', 'port', 'username', 'password', 'database_name'])
    testing.value = true
    emit('test', { ...formData })
  } catch (error) {
    ElMessage.warning('请先填写完整的连接信息')
  } finally {
    testing.value = false
  }
}

// 重置表单
const handleReset = () => {
  formRef.value?.resetFields()
  if (props.mode === 'create') {
    Object.assign(formData, {
      name: '',
      environment: 'test',
      host: '',
      port: 3306,
      username: '',
      password: '',
      database_name: '',
      description: '',
      connect_timeout: 5,
      charset: 'utf8mb4',
      use_ssl: false
    })
  }
}

// 取消操作
const handleCancel = () => {
  emit('cancel')
}

// 暴露方法给父组件
defineExpose({
  validate: () => formRef.value?.validate(),
  resetFields: () => formRef.value?.resetFields()
})
</script>

<style scoped>
.connection-form {
  max-width: 800px;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

:deep(.el-divider__text) {
  background-color: #fafafa;
}
</style>