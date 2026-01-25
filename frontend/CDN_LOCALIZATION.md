# CDN 资源本地化说明

本项目已将所有静态资源（字体、JS、CSS等）本地化，不再依赖外部 CDN。

## 已本地化的资源

### 1. Monaco Editor

**之前**: 使用 `@monaco-editor/loader` 从 CDN 动态加载 Monaco Editor 资源

**现在**: 使用 `vite-plugin-monaco-editor` 插件，所有资源打包到本地

**配置位置**: `vite.config.ts`

```typescript
import monacoEditorPlugin from 'vite-plugin-monaco-editor'

export default defineConfig({
  plugins: [
    vue(),
    monacoEditorPlugin({
      languageWorkers: ['editorWorkerService', 'typescript', 'json', 'html', 'css'],
      publicPath: 'monacoeditorwork',
    }),
  ],
})
```

**代码变更**:
- `useMonaco.ts`: 从 `import loader from '@monaco-editor/loader'` 改为 `import * as monaco from 'monaco-editor'`
- `QueryConsole.vue`: 同样改为直接导入 monaco-editor

**Worker 文件位置**: 
- 构建后，worker 文件会自动生成在 `dist/monacoeditorwork/` 目录
- 这些文件会在构建时自动打包，无需手动下载

### 2. 字体文件

**当前状态**: 使用系统字体，无需外部 CDN

**字体配置**: `src/style.css`
```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", sans-serif;
```

**Monaco Editor 字体**: 使用系统等宽字体
```typescript
fontFamily: "'JetBrains Mono', 'Fira Code', monospace"
```

**本地字体文件**: 
- `src/assets/fonts/nunito-v16-latin-regular.woff2` - 已存在但未使用

### 3. 其他依赖

所有其他依赖（Vue、Naive UI、VXE Table、Lucide Icons 等）都通过 npm 安装，已包含在 `node_modules` 中，构建时会打包到 `dist` 目录。

## 构建说明

运行构建命令时，所有资源会自动打包：

```bash
npm run build
```

构建后的文件结构：
```
dist/
├── assets/          # 打包后的 JS、CSS 文件
├── monacoeditorwork/ # Monaco Editor worker 文件
└── index.html
```

## 优势

1. **离线可用**: 应用完全离线可用，不依赖外部网络
2. **性能提升**: 本地资源加载更快，无 CDN 延迟
3. **稳定性**: 不依赖外部 CDN 的可用性
4. **安全性**: 所有资源都经过构建过程，更安全

## 注意事项

- `@monaco-editor/loader` 包仍在 `package.json` 中，但已不再使用
- 如需移除，可以运行: `npm uninstall @monaco-editor/loader`
- Monaco Editor worker 文件会在构建时自动生成，无需手动管理
