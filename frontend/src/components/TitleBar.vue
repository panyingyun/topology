<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Database, Square, Minimize2, Sun, Moon } from 'lucide-vue-next'
import { WindowMinimise, WindowToggleMaximise, WindowIsMaximised, Quit } from '../../wailsjs/runtime/runtime'
import { useI18n } from 'vue-i18n'
import LanguageSwitcher from './LanguageSwitcher.vue'
import { useTheme } from '../composables/useTheme'

const { t } = useI18n()
const { theme, toggleTheme } = useTheme()
const isMaximised = ref(false)

const updateMaximised = async () => {
  isMaximised.value = await WindowIsMaximised()
}

const handleMinimize = () => {
  WindowMinimise()
}

const handleMaximize = async () => {
  WindowToggleMaximise()
  setTimeout(updateMaximised, 100)
}

const handleClose = () => {
  Quit()
}

onMounted(() => {
  updateMaximised()
})
</script>

<template>
  <header
    class="flex h-10 items-center justify-between theme-bg-panel px-3 border-b theme-border z-50 select-none"
    style="--wails-draggable: drag"
  >
    <div class="flex items-center gap-2 text-sm font-semibold theme-text">
      <div class="w-5 h-5 bg-[#1677ff] rounded flex items-center justify-center">
        <Database :size="12" class="text-white" />
      </div>
      <span class="tracking-tight">{{ t('common.appName') }}</span>
    </div>

    <div class="flex items-center gap-2 no-drag" style="--wails-draggable: no-drag">
      <button
        @click="toggleTheme"
        class="flex items-center justify-center w-10 h-10 rounded theme-bg-hover transition-colors"
        :title="theme === 'dark' ? t('theme.light') : t('theme.dark')"
      >
        <Sun v-if="theme === 'dark'" :size="14" class="theme-text-muted" />
        <Moon v-else :size="14" class="theme-text-muted" />
      </button>
      <LanguageSwitcher />
      <button
        @click="handleMinimize"
        class="theme-bg-hover w-10 h-10 transition-colors text-xs flex items-center justify-center theme-text"
      >
        −
      </button>
      <button
        @click="handleMaximize"
        class="theme-bg-hover w-10 h-10 transition-colors flex items-center justify-center"
        :title="isMaximised ? t('window.restore') : t('window.maximize')"
      >
        <Minimize2 v-if="isMaximised" :size="14" class="theme-text-muted" />
        <Square v-else :size="12" class="theme-text-muted" />
      </button>
      <button
        @click="handleClose"
        class="hover:bg-red-600 w-10 h-10 transition-colors text-xs flex items-center justify-center theme-text"
      >
        ×
      </button>
    </div>
  </header>
</template>

<style scoped>
.no-drag {
  --wails-draggable: no-drag;
}
</style>
