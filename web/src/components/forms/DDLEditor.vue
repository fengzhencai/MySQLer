<template>
  <div class="ddl-editor">
    <!-- 编辑器头部 -->
    <div class="editor-header">
      <div class="header-left">
        <el-radio-group v-model="ddlType" @change="handleDDLTypeChange">
          <el-radio-button label="fragment">碎片整理</el-radio-button>
          <el-radio-button label="other">自定义</el-radio-button>
        </el-radio-group>
      </div>
      
      <div class="header-right">
        <el-button @click="handleFormat" :disabled="!ddlContent.trim()">
          <el-icon><Edit /></el-icon>
          格式化
        </el-button>
        <el-button @click="handleClear">
          <el-icon><Delete /></el-icon>
          清空
        </el-button>
      </div>
    </div>

    <!-- DDL模板提示 -->
    <div v-if="ddlType !== 'other'" class="template-hint">
      <el-alert
        :title="getTemplateTitle()"
        type="info"
        show-icon
        :closable="false"
      >
        <template #default>
          <div>{{ getTemplateDescription() }}</div>
          <div class="template-example">
            <strong>示例：</strong>
            <code>{{ getTemplateExample() }}</code>
          </div>
        </template>
      </el-alert>
    </div>

    <!-- DDL编辑器 -->
    <div class="editor-container">
      <textarea
        v-model="ddlContent"
        class="ddl-textarea"
        :placeholder="getPlaceholder()"
        :readonly="ddlType === 'fragment'"
        @input="handleInput"
      ></textarea>
    </div>

    <!-- 编辑器底部信息 -->
    <div class="editor-footer">
      <div class="footer-left">
        <span class="word-count">字符数: {{ ddlContent.length }}</span>
        <span v-if="ddlContent.trim()" class="line-count">
          行数: {{ ddlContent.split('\n').length }}
        </span>
      </div>
      
      <div class="footer-right">
        <el-tag v-if="isValidDDL" type="success" size="small">
          <el-icon><CircleCheck /></el-icon>
          DDL语法正确
        </el-tag>
        <el-tag v-else-if="ddlContent.trim()" type="warning" size="small">
          <el-icon><Warning /></el-icon>
          请检查DDL语法
        </el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Edit, Delete, CircleCheck, Warning } from '@element-plus/icons-vue'
import type { DDLType } from '@/types/execution'

// Props
interface Props {
  modelValue?: string
  ddlType?: DDLType
  tableName?: string
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  ddlType: 'other',
  tableName: '',
  disabled: false
})

// Emits
interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'update:ddlType', value: DDLType): void
  (e: 'change', data: { ddl: string, type: DDLType }): void
}

const emit = defineEmits<Emits>()

// 响应式数据
const ddlContent = ref(props.modelValue)
const ddlType = ref<DDLType>(props.ddlType)

// 计算属性
const isValidDDL = computed(() => {
  const content = ddlContent.value.trim().toUpperCase()
  if (!content) return false
  
  // 基本DDL语法检查
  const ddlKeywords = ['ALTER', 'CREATE', 'DROP', 'ADD', 'MODIFY', 'CHANGE']
  return ddlKeywords.some(keyword => content.includes(keyword))
})

// 监听器
watch(() => props.modelValue, (newValue) => {
  ddlContent.value = newValue || ''
})

watch(() => props.ddlType, (newValue) => {
  ddlType.value = newValue
})

watch(ddlContent, (newValue) => {
  emit('update:modelValue', newValue)
  emit('change', { ddl: newValue, type: ddlType.value })
})

watch(ddlType, (newValue) => {
  emit('update:ddlType', newValue)
  handleDDLTypeChange()
})

// 方法
const handleDDLTypeChange = () => {
  if (ddlType.value === 'fragment') {
    ddlContent.value = '-- 碎片整理操作会自动生成ALTER TABLE ENGINE=InnoDB语句'
  } else {
    ddlContent.value = getTemplateContent()
  }
}

const handleInput = () => {
  // 输入处理逻辑
}

const handleFormat = () => {
  try {
    // 简单的SQL格式化
    let formatted = ddlContent.value.trim()
    
    // 大写关键字
    const keywords = ['ALTER', 'TABLE', 'ADD', 'DROP', 'MODIFY', 'CHANGE', 'COLUMN', 'INDEX', 'KEY', 'PRIMARY', 'UNIQUE', 'CONSTRAINT']
    keywords.forEach(keyword => {
      const regex = new RegExp(`\\b${keyword}\\b`, 'gi')
      formatted = formatted.replace(regex, keyword)
    })
    
    // 添加换行和缩进
    formatted = formatted
      .replace(/,\s*/g, ',\n  ')
      .replace(/\(\s*/g, '(\n  ')
      .replace(/\s*\)/g, '\n)')
    
    ddlContent.value = formatted
    ElMessage.success('DDL格式化完成')
  } catch (error) {
    ElMessage.error('格式化失败，请检查DDL语法')
  }
}

const handleClear = () => {
  ddlContent.value = ''
  ElMessage.info('已清空DDL内容')
}

const getTemplateTitle = () => {
  const titles: Record<DDLType, string> = {
    fragment: '碎片整理操作',
    add_column: '自定义DDL',
    modify_column: '自定义DDL',
    drop_column: '自定义DDL',
    add_index: '自定义DDL',
    drop_index: '自定义DDL',
    other: '自定义DDL'
  }
  return titles[ddlType.value] || '自定义DDL'
}

const getTemplateDescription = () => {
  const descriptions: Record<DDLType, string> = {
    fragment: '对表进行碎片整理，优化存储空间和查询性能。无需手动编写DDL，系统会自动生成。',
    add_column: '自定义DDL操作，支持任何有效的ALTER TABLE语句。',
    modify_column: '自定义DDL操作，支持任何有效的ALTER TABLE语句。',
    drop_column: '自定义DDL操作，支持任何有效的ALTER TABLE语句。',
    add_index: '自定义DDL操作，支持任何有效的ALTER TABLE语句。',
    drop_index: '自定义DDL操作，支持任何有效的ALTER TABLE语句。',
    other: '自定义DDL操作，支持任何有效的ALTER TABLE语句。'
  }
  return descriptions[ddlType.value] || ''
}

const getTemplateExample = () => {
  const examples: Record<DDLType, string> = {
    fragment: '自动生成：ALTER TABLE table_name ENGINE=InnoDB',
    add_column: 'ADD COLUMN new_column VARCHAR(255) NOT NULL DEFAULT \'\'',
    modify_column: 'MODIFY COLUMN existing_column VARCHAR(512) NOT NULL',
    drop_column: 'DROP COLUMN old_column',
    add_index: 'ADD INDEX idx_column_name (column_name)',
    drop_index: 'DROP INDEX idx_name',
    other: 'ADD COLUMN new_col INT, ADD INDEX idx_new_col (new_col)'
  }
  return examples[ddlType.value] || ''
}

const getTemplateContent = () => {
  if (ddlType.value === 'fragment' || !props.tableName) return ''
  
  const templates: Record<DDLType, string> = {
    fragment: '',
    add_column: '',
    modify_column: '',
    drop_column: '',
    add_index: '',
    drop_index: '',
    other: ''
  }
  
  return templates[ddlType.value] || ''
}

const getPlaceholder = () => {
  if (ddlType.value === 'fragment') {
    return '碎片整理操作会自动生成DDL语句，无需手动输入'
  }
  
  const placeholders: Record<DDLType, string> = {
    fragment: '',
    add_column: '请输入自定义的ALTER TABLE DDL语句',
    modify_column: '请输入自定义的ALTER TABLE DDL语句',
    drop_column: '请输入自定义的ALTER TABLE DDL语句',
    add_index: '请输入自定义的ALTER TABLE DDL语句',
    drop_index: '请输入自定义的ALTER TABLE DDL语句',
    other: '请输入自定义的ALTER TABLE DDL语句'
  }
  
  return placeholders[ddlType.value] || '请输入DDL语句'
}
</script>

<style scoped>
.ddl-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
}

.header-left {
  flex: 1;
}

.header-right {
  display: flex;
  gap: 8px;
}

.template-hint {
  margin: 16px;
}

.template-example {
  margin-top: 8px;
  font-size: 12px;
}

.template-example code {
  background-color: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}

.editor-container {
  position: relative;
}

.ddl-textarea {
  width: 100%;
  min-height: 200px;
  padding: 16px;
  border: none;
  outline: none;
  resize: vertical;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  background-color: #fafbfc;
}

.ddl-textarea:focus {
  background-color: #fff;
}

.ddl-textarea::placeholder {
  color: #c0c4cc;
}

.editor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background-color: #f5f7fa;
  border-top: 1px solid #ebeef5;
  font-size: 12px;
}

.footer-left {
  display: flex;
  gap: 16px;
  color: #909399;
}

.word-count,
.line-count {
  display: flex;
  align-items: center;
}

.footer-right {
  display: flex;
  align-items: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .editor-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .header-left :deep(.el-radio-group) {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  
  .header-left :deep(.el-radio-button) {
    flex: 1;
    min-width: 80px;
  }
  
  .template-hint {
    margin: 12px;
  }
  
  .ddl-textarea {
    min-height: 150px;
    padding: 12px;
  }
  
  .editor-footer {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .footer-left {
    justify-content: center;
  }
  
  .footer-right {
    justify-content: center;
  }
}
</style>