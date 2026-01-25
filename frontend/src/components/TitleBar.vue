<script setup lang="ts">
import { Database } from 'lucide-vue-next'
import { WindowMinimise, Quit } from '../../wailsjs/runtime/runtime'
import { useI18n } from 'vue-i18n'
import LanguageSwitcher from './LanguageSwitcher.vue'

const { t } = useI18n()

const handleMinimize = () => {
  WindowMinimise()
}

const handleClose = () => {
  Quit()
}
</script>

<template>
  <header
    class="flex h-10 items-center justify-between bg-[#252526] px-3 border-b border-[#333] z-50 select-none"
    style="--wails-draggable: drag"
  >
    <div class="flex items-center gap-2 text-sm font-semibold">
      <div class="w-5 h-5 bg-[#1677ff] rounded flex items-center justify-center">
        <Database :size="12" class="text-white" />
      </div>
      <span class="tracking-tight text-gray-300">{{ t('common.appName') }}</span>
    </div>

    <div class="flex items-center gap-2 no-drag" style="--wails-draggable: no-drag">
      <LanguageSwitcher />
      <button
        @click="handleMinimize"
        class="hover:bg-[#37373d] w-10 h-10 transition-colors text-xs flex items-center justify-center"
      >
        −
      </button>
      <button
        @click="handleClose"
        class="hover:bg-red-600 w-10 h-10 transition-colors text-xs flex items-center justify-center"
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
