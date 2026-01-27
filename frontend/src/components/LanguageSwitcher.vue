<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Globe } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { setLocale, type SupportedLocale } from '../locales'

const { locale } = useI18n()

const showMenu = ref(false)
const menuRef = ref<HTMLElement | null>(null)

const languages: Array<{ code: SupportedLocale; label: string; flag: string }> = [
  { code: 'zh-CN', label: '简体中文', flag: '🇨🇳' },
  { code: 'en-US', label: 'English', flag: '🇺🇸' },
]

const currentLanguage = computed(() => {
  return languages.find((lang) => lang.code === locale.value) || languages[0]
})

const handleLanguageChange = (langCode: SupportedLocale) => {
  setLocale(langCode)
  showMenu.value = false
}

const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (menuRef.value && !menuRef.value.contains(target)) {
    showMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="relative" ref="menuRef">
    <button
      @click.stop="showMenu = !showMenu"
      class="flex items-center gap-1.5 px-2 py-1 rounded text-xs theme-bg-input theme-bg-input-hover theme-text transition-colors"
      :title="currentLanguage.label"
    >
      <Globe :size="12" />
      <span class="text-[10px]">{{ currentLanguage.flag }}</span>
    </button>

    <Transition name="fade">
      <div
        v-if="showMenu"
        class="absolute right-0 top-full mt-1 theme-bg-panel border theme-border rounded shadow-lg py-1 min-w-[140px] z-50"
        @click.stop
      >
        <button
          v-for="lang in languages"
          :key="lang.code"
          @click="handleLanguageChange(lang.code)"
          :class="[
            'w-full px-4 py-2 text-left text-xs transition-colors flex items-center gap-2',
            locale === lang.code
              ? 'bg-[#1677ff]/20 text-[#1677ff]'
              : 'theme-text theme-bg-hover'
          ]"
        >
          <span>{{ lang.flag }}</span>
          <span>{{ lang.label }}</span>
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
