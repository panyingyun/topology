<script setup lang="ts">
import { onMounted, onUnmounted, computed } from 'vue'
import MainLayout from './views/MainLayout.vue'
import { useTheme } from './composables/useTheme'
import { darkTheme, lightTheme } from 'naive-ui'

const { theme } = useTheme()
const naiveTheme = computed(() => (theme.value === 'dark' ? darkTheme : lightTheme))

// 禁用默认右键菜单（Back/Forward/Reload/Inspect 等），点击任意处右键无响应
function preventContextMenu(e: MouseEvent) {
  e.preventDefault()
}

onMounted(() => {
  document.addEventListener('contextmenu', preventContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('contextmenu', preventContextMenu)
})
</script>

<template>
  <n-config-provider :theme="naiveTheme">
    <n-message-provider>
      <MainLayout />
    </n-message-provider>
  </n-config-provider>
</template>
