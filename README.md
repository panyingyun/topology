# Topology

一款基于 Wails 框架开发的跨平台数据库管理工具，提供类似 Navicat 的功能体验。

![Status](https://img.shields.io/badge/status-in%20development-yellow)
![Go Version](https://img.shields.io/badge/go-1.23.1-blue)
![Vue Version](https://img.shields.io/badge/vue-3.2.37-green)
![License](https://img.shields.io/badge/license-MIT-blue)

## ✨ 功能特性

### 已实现 ✅

- **连接管理**
  - 支持 MySQL、PostgreSQL、SQLite 数据库连接
  - 创建、测试、删除数据库连接
  - 连接状态实时显示
  - 连接树形结构展示（连接 -> 数据库 -> 表）

- **SQL 查询编辑器**
  - Monaco Editor 集成，提供专业的代码编辑体验
  - SQL 语法高亮和自动补全
  - Ctrl+Enter 快速执行查询
  - SQL 格式化功能
  - 查询结果实时展示

- **数据查看与编辑**
  - 高性能数据网格（基于 vxe-table）
  - 虚拟滚动，支持大数据集
  - 单元格双击编辑
  - 修改跟踪和批量保存
  - 数据导出功能

- **多标签页管理**
  - 支持多个查询标签页
  - 支持表数据查看标签页
  - 标签页切换和关闭

- **用户界面**
  - 现代化深色主题
  - 可拖拽调整的侧边栏
  - 自定义标题栏（支持窗口拖拽）
  - 状态栏显示连接信息和查询统计

### 计划中 🚧

- 表结构设计器
- 数据导入功能
- AI SQL 优化
- 查询历史记录
- 连接持久化存储
- 快捷键系统完善

## 🛠️ 技术栈

### 前端
- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **UI 组件库**:
  - [Naive UI](https://www.naiveui.com/) - 基础组件
  - [vxe-table](https://vxetable.cn/) - 高性能数据表格
  - [Monaco Editor](https://microsoft.github.io/monaco-editor/) - 代码编辑器
  - [Lucide Vue Next](https://lucide.dev/) - 图标库
- **样式**: [Tailwind CSS](https://tailwindcss.com/)
- **构建工具**: [Vite](https://vitejs.dev/)

### 后端
- **语言**: Go 1.23.1
- **框架**: [Wails v2](https://wails.io/)
- **架构**: 前后端分离，后端提供接口返回 mock 数据

## 📁 项目结构

```
topology/
├── app.go                    # 后端接口实现（mock 数据）
├── main.go                   # Wails 应用入口
├── frontend/
│   ├── src/
│   │   ├── components/       # Vue 组件
│   │   │   ├── TitleBar.vue      # 自定义标题栏
│   │   │   ├── Sidebar.vue       # 侧边栏
│   │   │   ├── ConnectionTree.vue # 连接树
│   │   │   ├── TabBar.vue        # 标签页
│   │   │   ├── StatusBar.vue     # 状态栏
│   │   │   └── DataGrid.vue      # 数据网格
│   │   ├── views/            # 页面视图
│   │   │   ├── MainLayout.vue    # 主布局
│   │   │   ├── ConnectionManager.vue # 连接管理器
│   │   │   └── QueryConsole.vue  # SQL 控制台
│   │   ├── services/         # 服务层
│   │   │   ├── connectionService.ts
│   │   │   ├── queryService.ts
│   │   │   └── dataService.ts
│   │   ├── composables/      # 组合式函数
│   │   │   └── useMonaco.ts
│   │   ├── types/            # TypeScript 类型定义
│   │   │   └── index.ts
│   │   ├── App.vue           # 根组件
│   │   ├── main.ts           # 入口文件
│   │   └── style.css         # 全局样式
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   └── package.json
├── docs/                     # 文档目录
│   └── development-plan.md  # 开发计划
└── README.md
```

## 🚀 快速开始

### 环境要求

- Go 1.23.1 或更高版本
- Node.js 18+ 和 npm
- Wails CLI (安装方法见下方)

### 安装 Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 安装依赖

```bash
# 安装前端依赖
cd frontend
npm install

# 返回项目根目录
cd ..
```

### 开发模式

启动开发服务器（支持热重载）：

```bash
wails dev
```

开发服务器会在以下地址运行：
- 桌面应用窗口（主要开发方式）
- 浏览器访问：http://localhost:34115（用于调试）

### 构建应用

构建生产版本：

```bash
# 标准构建
wails build

# 清理构建（推荐）
wails build -clean
```

构建产物位于 `build/bin/` 目录。

### 前端单独构建

如果需要单独构建前端：

```bash
cd frontend
npm run build
```

## 📖 使用说明

### 创建数据库连接

1. 点击侧边栏的 "NEW CONNECTION" 按钮
2. 选择数据库类型（MySQL、PostgreSQL 或 SQLite）
3. 填写连接信息（主机、端口、用户名、密码等）
4. 点击 "Test Connection" 测试连接
5. 点击 "Connect" 创建连接

### 执行 SQL 查询

1. 在连接树中选择一个连接
2. 系统会自动创建一个查询标签页
3. 在 SQL 编辑器中编写 SQL 语句
4. 按 `Ctrl+Enter` 或点击 "EXECUTE" 按钮执行查询
5. 查询结果会在下方数据网格中显示

### 查看表数据

1. 在连接树中展开连接
2. 点击表名
3. 系统会创建一个表数据标签页（待实现完整功能）

### 编辑数据

1. 在数据网格中双击单元格
2. 修改数据后，系统会跟踪变更
3. 点击 "Save Changes" 保存修改

## 🎨 UI 设计

- **主题**: 深色主题
- **主色调**: `#1677ff` (科技蓝)
- **字体**: 
  - 代码: JetBrains Mono, Fira Code
  - UI: 系统默认字体

## 📝 开发状态

### 已完成 ✅

- [x] 项目初始化和依赖安装
- [x] 类型定义和接口设计
- [x] 所有核心组件实现
- [x] 所有页面视图实现
- [x] 服务层和组合式函数
- [x] 前后端集成
- [x] 构建验证

### 进行中 🚧

- [ ] 表数据视图完整实现
- [ ] 数据导入导出功能完善
- [ ] 表结构设计器
- [ ] AI SQL 优化功能

详细开发计划请查看 [docs/development-plan.md](docs/development-plan.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🔗 相关链接

- [Wails 文档](https://wails.io/docs)
- [Vue 3 文档](https://vuejs.org/)
- [Tailwind CSS 文档](https://tailwindcss.com/docs)

---

**最后更新**: 2026-01-24
