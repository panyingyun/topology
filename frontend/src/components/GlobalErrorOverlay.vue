<script setup lang="ts">
import { watch, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { globalError, clearGlobalError } from '../state/globalError'

const { t } = useI18n()
const copied = ref(false)

watch(
  () => globalError.value,
  (v) => {
    if (v) copied.value = false
  }
)

function copyError() {
  const e = globalError.value
  if (!e) return
  const text = e.stack ? `${e.message}\n\n${e.stack}` : e.message
  navigator.clipboard?.writeText(text).then(
    () => { copied.value = true },
    () => {}
  )
}

function close() {
  clearGlobalError()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="globalError"
        class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/50 backdrop-blur-sm"
        role="alertdialog"
        aria-labelledby="global-error-title"
        aria-describedby="global-error-desc"
      >
        <div
          class="theme-bg-panel rounded-lg border theme-border w-full max-w-lg shadow-xl overflow-hidden"
          @click.stop
        >
          <div class="px-4 py-3 border-b theme-border flex items-center justify-between">
            <h2 id="global-error-title" class="text-sm font-semibold theme-text">
              {{ t('errorOverlay.title') }}
            </h2>
            <button
              type="button"
              class="theme-text-muted hover:theme-text p-1 rounded transition-colors"
              :aria-label="t('common.close')"
              @click="close"
            >
              ✕
            </button>
          </div>
          <div id="global-error-desc" class="p-4 space-y-3">
            <p class="text-xs theme-text break-words max-h-32 overflow-y-auto">
              {{ globalError?.message }}
            </p>
            <div class="flex gap-2">
              <button
                type="button"
                class="px-3 py-1.5 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
                @click="copyError"
              >
                {{ copied ? t('errorOverlay.copied') : t('errorOverlay.copy') }}
              </button>
              <button
                type="button"
                class="px-3 py-1.5 rounded text-xs bg-[#1677ff] hover:bg-[#4096ff] text-white"
                @click="close"
              >
                {{ t('common.close') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
