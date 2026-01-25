import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
// @ts-ignore
import monacoEditorPlugin from 'vite-plugin-monaco-editor'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    (monacoEditorPlugin.default || monacoEditorPlugin)({
      languageWorkers: ['editorWorkerService', 'typescript', 'json', 'html', 'css'],
      publicPath: 'monacoeditorwork',
    }),
  ],
})
