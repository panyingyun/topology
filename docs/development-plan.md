# 数据库管理 GUI 开发计划

## 项目概述

基于 Wails 框架，使用 React + TypeScript + Tailwind CSS 实现类似 Navicat 的数据库管理 GUI 工具。优先实现前端展示页面，后端接口使用 mock 数据。包含 6 个核心功能模块：连接管理、数据查看器、SQL 编辑器、表设计器、AI 优化和导入导出。

## 技术栈

- **前端**: React 18 + TypeScript + Tailwind CSS + React Router
- **后端**: Go (Wails v2) — 先实现 mock 接口
- **UI 设计**: 参考 docs 下 01～05 的设计文件，新拟态风格 + 毛玻璃效果

## 文件结构规划

```
frontend/src/
├── components/           # 可复用组件
│   ├── Sidebar.tsx      # 侧边栏导航
│   ├── Header.tsx       # 顶部导航栏
│   ├── ConnectionCard.tsx
│   ├── TableList.tsx
│   └── DataTable.tsx
├── pages/               # 页面组件
│   ├── MainPage.tsx     # 主页面（docs/01 设计）
│   ├── MainLayout.tsx   # 主布局
│   ├── Connections.tsx  # 连接管理
│   ├── QueryEditor.tsx  # SQL 查询编辑器
│   ├── DataViewer.tsx   # 数据查看器
│   ├── TableDesigner.tsx # 表设计器
│   └── ImportExport.tsx # 导入导出
├── services/            # 服务层（mock）
│   ├── connectionService.ts
│   ├── queryService.ts
│   ├── tableService.ts
│   └── importExportService.ts
├── types/
│   └── index.ts
├── utils/
│   └── sqlHighlighter.ts
└── App.tsx
```

## 实现步骤

### 1. 项目基础配置

**涉及文件**: `package.json`, `tailwind.config.js`, `postcss.config.js`

- 安装 Tailwind CSS、PostCSS、react-router-dom
- 自定义主题：主色 `#6366F1`、状态色 `#10B981`、背景色 `#f6f6f8` / `#101122`
- 新拟态阴影、毛玻璃样式

### 2. 全局样式与基础组件

**涉及文件**: `style.css`, `Sidebar.tsx`, `Header.tsx`

- Tailwind 指令、Material Symbols 字体
- `.glass-sidebar`、`.glass-header`、`.neumorphic-card`、`.neumorphic-inset`、`.custom-scrollbar`
- SQL 高亮类：`.sql-keyword`、`.sql-string`、`.sql-number`、`.sql-comment`

### 3. 类型定义

**文件**: `types/index.ts`

- `Connection`, `Database`, `Table`, `Column`, `Index`, `ForeignKey`
- `TableSchema`, `QueryResult`, `OptimizedSQL`
- `ImportFormat`, `ExportFormat`, `ImportOptions`, `ExportOptions`
- `TableData`, `TableDataUpdate`

### 4. Mock 服务层

**目录**: `services/`

- **connectionService**: `getConnections`, `createConnection`, `testConnection`, `deleteConnection`
- **queryService**: `executeQuery`, `optimizeSQL`
- **tableService**: `getTables`, `getTableData`, `updateTableData`, `getTableSchema`, `updateTableSchema`
- **importExportService**: `importData`, `exportData`

### 5. 连接管理页面

**文件**: `Connections.tsx`  
**参考**: `docs/02/code.html`

- 连接列表（分组：Production / Development）、搜索
- New Connection 按钮、连接状态指示
- 连接对话框：数据库类型、Host/Port/Username/Password/Database、测试连接、保存密码

### 6. SQL 查询编辑器

**文件**: `QueryEditor.tsx`, `utils/sqlHighlighter.ts`  
**参考**: `docs/01/code.html`

- 带行号的 SQL 编辑区、语法高亮
- Run、AI Optimize、Save 按钮
- 查询结果区、可编辑数据表格、状态栏

### 7. 数据查看器与编辑器

**文件**: `DataViewer.tsx`, `DataTable.tsx`

- 表格展示、可编辑单元格（input/select）
- Add Row、Delete Row、Commit Changes
- 支持 `queryText` 等扩展 props

### 8. 表设计器

**文件**: `TableDesigner.tsx`  
**参考**: `docs/05/code.html`

- 列编辑（名称、类型、PK/Not Null/Unique、默认值）
- 添加/删除列、索引管理、外键管理
- SQL 预览、应用更改

### 9. AI 优化 SQL

- 集成在 QueryEditor / MainPage
- AI Optimize 按钮、优化前后对比、建议说明、应用优化结果
- Mock：索引提示、优化 SELECT 字段等

### 10. 导入导出

**文件**: `ImportExport.tsx`  
**参考**: `docs/03`, `docs/04`

- **导入**: 文件上传（拖拽）、格式（CSV/JSON/XLSX/SQL）、编码、分隔符、字段映射
- **导出**: 格式（CSV/Excel/SQL Dump）、数据源、行数限制、日期过滤

### 11. 主布局与路由

**文件**: `MainLayout.tsx`, `App.tsx`

- 主布局：Header + Sidebar + 主内容区
- 路由：
  - `/` → MainPage（docs/01 主页面）
  - `/connections` → 连接管理
  - `/query/:connectionId?` → SQL 编辑器
  - `/table/:connectionId/:tableName` → 数据查看
  - `/designer/:connectionId/:tableName?` → 表设计器
  - `/import-export` → 导入导出

### 12. 后端 Mock 接口

**文件**: `app.go`

- 连接：`GetConnections`, `CreateConnection`, `TestConnection`, `DeleteConnection`
- 查询：`ExecuteQuery`, `OptimizeSQL`
- 表：`GetTables`, `GetTableData`, `UpdateTableData`, `GetTableSchema`, `UpdateTableSchema`
- 导入导出：`ImportData`, `ExportData`  
所有接口返回 mock 数据，不连接真实数据库。

## 设计要点

1. **UI**: 新拟态、毛玻璃、支持深色模式
2. **交互**: 动效、响应式、快捷键（如 Ctrl+Enter 执行）
3. **代码**: 组件化、TypeScript 类型完整、服务层可替换为真实 API

## 开发顺序建议

1. 基础配置与样式
2. 主布局与路由
3. 连接管理页面
4. SQL 编辑器（含主页面 docs/01）
5. 数据查看器
6. 表设计器
7. 导入导出
8. 后端 Mock 接口

## 注意事项

- 后端暂全部 mock，前端按真实 API 契约设计
- 遵循 docs 设计稿的视觉与布局
- 保持模块化，便于后续接入真实数据库与 API
