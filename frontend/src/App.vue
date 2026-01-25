<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import MainLayout from './views/MainLayout.vue'

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
  <MainLayout />
</template>
