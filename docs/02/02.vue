<script setup lang="ts">
import { ref, reactive } from 'vue'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits(['close', 'connect'])

// 1. 基础状态
const activeDbType = ref('mysql')
const testStatus = ref<'idle' | 'loading' | 'success' | 'error'>('idle')
const errorMessage = ref('')

// 2. 表单数据模型
const form = reactive({
  name: '',
  host: '127.0.0.1',
  port: 3306,
  user: 'root',
  password: '',
  database: '',
  useSSL: false
})

// 3. 切换数据库类型时的默认端口处理
const handleTypeChange = (type: string) => {
  activeDbType.ref = type
  if (type === 'mysql') form.port = 3306
  if (type === 'postgres') form.port = 5432
}

// 4. 模拟测试连接 (实际调用 Wails 后端)
const testConnection = async () => {
  testStatus.value = 'loading'
  
  // 模拟 Wails 后端延迟
  // const result = await window.go.main.App.TestConnection(JSON.stringify(form))
  setTimeout(() => {
    testStatus.value = 'success' // 假设成功
    setTimeout(() => { testStatus.value = 'idle' }, 2000)
  }, 1500)
}
</script>

<template>
  <Transition name="fade