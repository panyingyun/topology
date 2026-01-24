import { ref, onMounted, onUnmounted, shallowRef, type Ref } from 'vue'
import loader from '@monaco-editor/loader'

export function useMonaco(container: Ref<HTMLElement | null>) {
  const editor = shallowRef<any>(null)
  const isReady = ref(false)
  const content = ref('')
  const line = ref(1)
  const column = ref(1)

  onMounted(async () => {
    const monaco = await loader.init()
    
    if (container.value) {
      editor.value = monaco.editor.create(container.value, {
        value: content.value,
        language: 'sql',
        theme: 'vs-dark',
        automaticLayout: true,
        minimap: { enabled: true },
        fontSize: 14,
        fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
        padding: { top: 12 },
        scrollBeyondLastLine: false,
        wordWrap: 'on',
      })

      // Listen to content changes
      editor.value.onDidChangeModelContent(() => {
        content.value = editor.value.getValue()
      })

      // Listen to cursor position changes
      editor.value.onDidChangeCursorPosition((e: any) => {
        line.value = e.position.lineNumber
        column.value = e.position.column
      })

      isReady.value = true
    }
  })

  onUnmounted(() => {
    if (editor.value) {
      editor.value.dispose()
    }
  })

  const setValue = (value: string) => {
    if (editor.value) {
      editor.value.setValue(value)
      content.value = value
    }
  }

  const getValue = (): string => {
    if (editor.value) {
      return editor.value.getValue()
    }
    return content.value
  }

  const getSelectedText = (): string => {
    if (editor.value) {
      const selection = editor.value.getSelection()
      if (selection && !selection.isEmpty()) {
        return editor.value.getModel()?.getValueInRange(selection) || ''
      }
    }
    return ''
  }

  const format = async () => {
    if (editor.value) {
      editor.value.getAction('editor.action.formatDocument')?.run()
    }
  }

  return {
    editor,
    isReady,
    content,
    line,
    column,
    setValue,
    getValue,
    getSelectedText,
    format,
  }
}
