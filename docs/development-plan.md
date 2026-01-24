# Topology 数据管理器开发计划

## 项目概述

Topology 是一款基于 Wails 框架开发的跨平台数据库管理工具，提供类似 Navicat 的功能体验。项目采用 Vue 3 + TypeScript 前端和 Go 后端架构，实现前后端分离，后端提供空接口返回 mock 数据。

## 技术栈

### 前端
- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **UI 库**: 
  - Naive UI (基础组件)
  - vxe-table (高性能数据网格)
  - Monaco Editor (SQL 编辑器)
  - Lucide Vue Next (图标库)
- **样式**: Tailwind CSS
- **构建工具**: Vite

### 后端
- **语言**: Go 1.23.1
- **框架**: Wails v2
- **架构**: 前后端分离，后端提供空接口返回 mock 数据

## 项目结构

```
topology/
├── app.go                 # 后端接口实现（mock 数据）
├── main.go                # Wails 应用入口
├── frontend/
│   ├── src/
│   │   ├── components/     # Vue 组件
│   │   │   ├── TitleBar.vue
│   │   │   ├── Sidebar.vue
│   │   │   ├── ConnectionTree.vue
│   │   │   ├── TabBar.vue
│   │   │   ├── StatusBar.vue
│   │   │   └── DataGrid.vue
│   │   ├── views/          # 页面视图
│   │   │   ├── MainLayout.vue
│   │   │   ├── ConnectionManager.vue
│   │   │   └── QueryConsole.vue
│   │   ├── services/       # 服务层
│   │   │   ├── connectionService.ts
│   │   │   ├── queryService.ts
│   │   │   └── dataService.ts
│   │   ├── composables/   # 组合式函数
│   │   │   └── useMonaco.ts
│   │   ├── types/          # TypeScript 类型定义
│   │   │   └── index.ts
│   │   ├── App.vue         # 根组件
│   │   ├── main.ts         # 入口文件
│   │   └── style.css       # 全局样式
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   └── package.json
└── docs/                   # 文档目录
```

## 开发任务清单

### ✅ 阶段一：项目配置与基础设置

- [x] 安装项目依赖（naive-ui, monaco-editor, vxe-table, lucide-vue-next等）
- [x] 配置 Tailwind CSS 和全局样式（深色主题、字体、颜色变量）
- [x] 定义 TypeScript 类型（Connection, Table, QueryResult等）
- [x] 在 app.go 中实现所有后端空接口（返回 mock 数据）

### ✅ 阶段二：核心组件实现

- [x] 实现自定义标题栏组件（支持窗口拖拽）
- [x] 实现主布局组件（左中底三栏布局）
- [x] 实现侧边栏组件（可拖拽调整宽度、搜索框、连接树）
- [x] 实现连接树形组件（显示连接->数据库->表层级）
- [x] 实现标签页组件（支持拖拽排序、Query Tab 和 Table Tab）
- [x] 实现状态栏组件（连接信息、查询耗时、行号等）

### ✅ 阶段三：功能视图实现

- [x] 实现连接管理器视图（数据库类型选择、表单、测试连接）
- [x] 实现 SQL 控制台视图（Monaco Editor 集成、执行工具条）
- [x] 实现数据网格组件（vxe-table、虚拟滚动、单元格编辑、筛选）

### ✅ 阶段四：服务层与组合式函数

- [x] 实现服务层（调用 Wails 后端接口）
- [x] 实现组合式函数（useMonaco, useConnection, useQuery, useDataGrid）

### ✅ 阶段五：交互流程完善

- [x] 完善交互流程（创建连接、查看表数据、执行 SQL）
- [x] 修复类型错误和编译问题
- [x] 验证构建流程

## 功能特性

### 1. 连接管理
- ✅ 创建新连接（支持 MySQL、PostgreSQL、SQLite）
- ✅ 测试连接
- ✅ 删除连接
- ✅ 连接状态显示（已连接/未连接）
- ✅ 连接树形展示

### 2. SQL 查询编辑器
- ✅ Monaco Editor 集成
- ✅ SQL 语法高亮
- ✅ 代码自动补全
- ✅ Ctrl+Enter 执行查询
- ✅ SQL 格式化
- ✅ 查询结果展示

### 3. 数据查看与编辑
- ✅ 表数据浏览
- ✅ 虚拟滚动（支持大数据集）
- ✅ 单元格双击编辑
- ✅ 修改跟踪
- ✅ 数据保存

### 4. 标签页管理
- ✅ 多标签页支持
- ✅ Query Tab（SQL 查询）
- ✅ Table Tab（表数据查看）
- ✅ 标签页关闭
- ✅ 标签页切换

### 5. 状态栏
- ✅ 当前连接信息显示
- ✅ 查询执行时间
- ✅ 受影响行数
- ✅ 编辑器光标位置

## UI 设计规范

### 颜色方案
- **主色调**: `#1677ff` (科技蓝)
- **背景色**: `#1e1e1e` (深灰)
- **侧边栏**: `#252526`
- **边框**: `#333`
- **文字**: `#d4d4d4` / `#cccccc`

### 字体
- **代码字体**: JetBrains Mono, Fira Code, monospace
- **UI 字体**: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto"

### 组件样式
- **圆角**: 统一使用 `rounded` / `rounded-md`
- **间距**: Tailwind 标准间距系统
- **滚动条**: 自定义样式（6px 宽度，深色主题）

## 后端接口清单

### 连接管理
- `GetConnections()` - 获取所有连接
- `CreateConnection(connJSON string)` - 创建连接
- `TestConnection(connJSON string)` - 测试连接
- `DeleteConnection(id string)` - 删除连接

### 查询执行
- `ExecuteQuery(connectionId, sql string)` - 执行 SQL 查询
- `FormatSQL(sql string)` - 格式化 SQL

### 表操作
- `GetTables(connectionId string)` - 获取表列表
- `GetTableData(connectionId, tableName string, limit, offset int)` - 获取表数据
- `UpdateTableData(connectionId, tableName, updatesJSON string)` - 更新表数据
- `GetTableSchema(connectionId, tableName string)` - 获取表结构
- `ExportData(connectionId, tableName, format string)` - 导出数据

## 开发进度

### 已完成 ✅
- [x] 项目初始化和依赖安装
- [x] 类型定义和接口设计
- [x] 所有核心组件实现
- [x] 所有页面视图实现
- [x] 服务层和组合式函数
- [x] 前后端集成
- [x] 构建验证

### 待完善 🔄
- [ ] 表数据视图完整实现（当前为占位符）
- [ ] 数据导入导出功能完善
- [ ] 表结构设计器
- [ ] AI SQL 优化功能
- [ ] 连接持久化存储
- [ ] 查询历史记录
- [ ] 快捷键系统完善
- [ ] 错误处理和用户提示优化

## 构建与运行

### 开发模式
```bash
# 启动开发服务器
wails dev
```

### 构建生产版本
```bash
# 构建应用
wails build

# 清理构建
wails build -clean
```

### 前端单独构建
```bash
cd frontend
npm install
npm run build
```

## 技术要点

### 1. Wails 集成
- 使用 `wails generate module` 生成前端绑定
- 通过 `wailsjs/go/main/App` 调用后端接口
- 使用 `wailsjs/runtime/runtime` 进行窗口操作

### 2. Monaco Editor 集成
- 使用 `@monaco-editor/loader` 动态加载
- 配置 SQL 语言支持
- 实现代码补全和语法高亮

### 3. vxe-table 配置
- 配置虚拟滚动
- 单元格编辑功能
- 数据变更跟踪

### 4. 响应式设计
- 使用 Vue 3 Composition API
- 响应式状态管理
- 组件间通信

## 注意事项

1. **类型安全**: 所有接口调用都有完整的 TypeScript 类型定义
2. **错误处理**: 服务层包含错误处理和日志记录
3. **性能优化**: 使用虚拟滚动处理大数据集
4. **用户体验**: 提供加载状态、错误提示等反馈

## 后续计划

1. **功能完善**
   - 实现完整的表数据视图
   - 添加数据导入导出功能
   - 实现表结构设计器

2. **性能优化**
   - 代码分割和懒加载
   - 查询结果缓存
   - 连接池管理

3. **用户体验**
   - 快捷键系统
   - 主题切换
   - 配置管理

4. **测试**
   - 单元测试
   - 集成测试
   - E2E 测试

---

**最后更新**: 2026-01-24
**状态**: 核心功能已完成，可正常运行
