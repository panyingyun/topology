# Topology

<img src="build/appicon.png" width="20%" height="20%">

一款基于 Wails 框架开发的跨平台数据库管理工具，提供类似 Navicat 的功能体验。

![Status](https://img.shields.io/badge/status-in%20development-yellow)
![Go Version](https://img.shields.io/badge/go-1.23.1-blue)
![Vue Version](https://img.shields.io/badge/vue-3.2.37-green)
![License](https://img.shields.io/badge/license-GPLv3-blue)

## ✨ 功能特性

### 已实现 ✅

- **连接管理**
  - 支持 MySQL、PostgreSQL、SQLite 数据库连接
  - 创建、测试、删除数据库连接（删除前确认提示）
  - **导入 Navicat 连接**：侧栏「导入 Navicat」可选择 .ncx 文件，自动解析并创建 MySQL/SQLite 连接（密码需后续编辑填写）
  - 连接状态实时显示
  - 连接树形结构展示（连接 -> 数据库 -> 表）
  - 连接右键：编辑、刷新、**立即备份** / **从备份恢复**、打开监控、删除

- **SQL 查询编辑器**
  - Monaco Editor 集成，提供专业的代码编辑体验
  - SQL 语法高亮和自动补全
  - Ctrl+Enter 快速执行查询
  - SQL 格式化功能
  - 查询结果实时展示
  - 查询历史记录（自动保存、搜索、快速选择）
  - 表名右键「查询」打开 SQL 查询窗口并预填该表查询语句
  - 查询窗口展示当前库/表标签（库 main / 表 xxx）
  - SQL 区域与结果区域比例可拖拽调节，默认 SQL 1/3、结果 2/3
  - 切换标签后返回查询窗口时，自动恢复该标签的 SQL 与查询结果
  - 查询执行超时（2 分钟）与错误提示，避免一直「运行中」
  - **执行计划**（MySQL / PostgreSQL）：工具栏「执行计划」按钮，可视化 EXPLAIN 结果（全表扫描/索引使用、优化建议）

- **数据查看与编辑**
  - 高性能数据网格（基于 vxe-table）
  - 虚拟滚动，支持大数据集（10w+ 数据不卡顿）
  - 单元格双击编辑
  - 修改跟踪和批量保存
  - 单元格编辑后黄色背景标记
  - 表头筛选功能（包含、等于、不为空、为空等）
  - 数据导出功能（CSV、JSON、SQL Insert）
  - 数据导入功能（CSV、JSON，支持预览和列映射）
  - 表数据查看器（DataViewer），支持分页加载

- **多标签页管理**
  - 支持多个查询标签页
  - 支持表数据查看标签页
  - 标签页切换和关闭
  - 标签页拖拽排序

- **实时监控**（MySQL）
  - 连接右键「打开监控」打开实时监控弹窗
  - 每 5 秒刷新：活跃连接数、进程列表（SHOW FULL PROCESSLIST）
  - 慢查询高亮（执行时间 ≥ 5 秒）

- **备份与恢复**（MySQL / PostgreSQL / SQLite）
  - **立即备份**：连接右键「立即备份」，选择保存路径，调用 mysqldump / pg_dump / sqlite3 .dump 生成 SQL 备份文件
  - **从备份恢复**：连接右键「从备份恢复」，从最近备份列表选择或「选择文件」，二次确认后执行恢复
  - **定时备份**：侧栏「备份管理」→ 定时备份，可配置每日/每周、执行时间与输出目录，后台按周期自动执行
  - **备份管理**：备份列表查看、验证（文件存在与大小）、删除；最近 50 次备份记录保存于用户配置目录

- **日志**
  - 分级日志（DEBUG / INFO / WARN / ERROR）写入 `用户配置目录/topology/logs/topology.log`
  - 通过环境变量 `TOPOLOGY_LOG_LEVEL` 配置最低级别（默认 INFO）

- **用户界面**
  - 亮色 / 暗色主题切换（标题栏主题按钮，偏好持久化）
  - 可拖拽调整的侧边栏
  - 自定义标题栏（支持窗口拖拽）
  - 窗口控制：最小化、最大化/还原、关闭
  - 状态栏显示连接信息、查询统计和编辑器位置（行号、列号）
  - 多语言支持（中文、英文），支持语言切换

### 计划中 🚧

- 快捷键系统完善（更多快捷键支持）
- PostgreSQL 执行计划、特色类型等增强（基础连接与查询已支持）
- 表结构可视化编辑（编辑现有表结构）

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
- **架构**: 前后端分离，后端提供真实数据接口（连接、查询、表数据、导入导出、执行计划、实时监控等）

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
│   │   │   ├── DataGrid.vue      # 数据网格
│   │   │   ├── ExecutionPlanViewer.vue # 执行计划可视化
│   │   │   └── LiveMonitor.vue   # 实时监控弹窗
│   │   ├── views/            # 页面视图
│   │   │   ├── MainLayout.vue    # 主布局
│   │   │   ├── ConnectionManager.vue # 连接管理器
│   │   │   ├── QueryConsole.vue  # SQL 控制台
│   │   │   └── DataViewer.vue    # 表数据查看器
│   │   ├── services/         # 服务层
│   │   │   ├── connectionService.ts
│   │   │   ├── queryService.ts
│   │   │   ├── dataService.ts
│   │   │   └── monitorService.ts # 实时监控
│   │   ├── composables/      # 组合式函数
│   │   │   ├── useMonaco.ts
│   │   │   ├── useTheme.ts   # 亮/暗主题切换与持久化
│   │   │   └── useSchemaMetadata.ts
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

### 发布构建（GoReleaser）

多平台分别使用对应配置文件发布，校验和文件名含平台信息（如 `topology_1.3.0_Ubuntu22.04_checksums.txt`）：

```bash
# 检查配置
goreleaser check -f .goreleaser.macos.yml
goreleaser check -f .goreleaser.ubuntu22.04.yml
goreleaser check -f .goreleaser.ubuntu24.04.yml
goreleaser check -f .goreleaser.windows.yml

# 发布（需先设置 GITHUB_TOKEN 并打 tag）
goreleaser release -f .goreleaser.macos.yml --clean
goreleaser release -f .goreleaser.ubuntu22.04.yml --clean
goreleaser release -f .goreleaser.ubuntu24.04.yml --clean
goreleaser release -f .goreleaser.windows.yml --clean
```

详见 [release.md](release.md)。

### 前端单独构建

如果需要单独构建前端：

```bash
cd frontend
npm run build
```

## 📖 使用说明

### 创建数据库连接

1. 点击侧边栏的「新建连接」按钮；或点击「导入 Navicat」选择 .ncx 文件批量导入（仅 MySQL/SQLite，密码需后续编辑）
2. 选择数据库类型（MySQL 或 SQLite，PostgreSQL 支持开发中）
3. 填写连接信息（主机、端口、用户名、密码等）
4. 点击「测试连接」测试连接
5. 点击「连接」创建连接

### 执行 SQL 查询

1. 在连接树中选择一个连接，系统会自动创建/聚焦该连接的查询标签页；或对表名**右键 →「查询」**，打开查询窗口并预填该表的 `SELECT * FROM … LIMIT 100`。
2. 在 SQL 编辑器中编写或修改 SQL 语句（当前库/表会显示在「执行」按钮左侧）。
3. 按 `Ctrl+Enter` 或点击「执行」按钮执行查询。
4. 查询结果在下方数据网格中显示；可拖拽中间的横条调整 SQL 区与结果区的比例（默认 1:2）。
5. 切换到其他标签再切回时，该查询标签的 SQL 与查询结果会自动恢复。

### 查看表数据

1. 在连接树中展开连接和数据库
2. 双击表名
3. 系统会创建一个表数据标签页，自动加载表数据
4. 支持分页浏览（Load More、Previous、Next）
5. 可以在数据网格中编辑、筛选和导出数据

### 编辑数据

1. 在数据网格中双击单元格进入编辑模式
2. 修改数据后，单元格会显示黄色背景标记
3. 系统会跟踪所有变更，显示待保存的变更数量
4. 点击 "Save Changes" 批量保存所有修改

### 筛选数据

1. 点击表头的筛选图标
2. 选择筛选条件（包含、等于、不为空、为空等）
3. 输入筛选值
4. 数据网格会自动应用筛选条件

### 导出数据

1. 在数据网格顶部点击 "Export" 按钮
2. 选择导出格式：
   - CSV：逗号分隔值格式
   - JSON：JSON 格式
   - SQL Insert：SQL INSERT 语句
3. 导出文件会保存到 `build/export/` 目录

## 🎨 UI 设计

- **主题**: 支持亮色与暗色主题，标题栏提供主题切换按钮，偏好保存在本地。
- **主色调**: `#1677ff` (科技蓝)
- **字体**: 
  - 代码: JetBrains Mono, Fira Code
  - UI: 系统默认字体

## 📝 开发状态

### 已完成 ✅

- [x] 项目初始化和依赖安装
- [x] 类型定义和接口设计
- [x] 所有核心组件实现
- [x] 所有页面视图实现（包括 DataViewer 表数据查看器）
- [x] 服务层和组合式函数
- [x] 前后端集成
- [x] 构建验证
- [x] 表数据视图完整实现（分页加载、数据编辑、筛选、导出）
- [x] 数据导出功能（CSV、JSON）
- [x] 标签页拖拽排序
- [x] 状态栏编辑器位置显示
- [x] 单元格编辑后视觉标记
- [x] 表头筛选功能
- [x] SQL Insert 导出格式实现
- [x] 连接持久化存储（文件存储 + AES-256 密码加密）
- [x] 查询历史记录功能（自动保存、搜索、快速选择）
- [x] 数据导入功能（CSV/JSON 导入、预览、列映射、批量插入）
- [x] 表结构设计器（可视化设计、生成 CREATE TABLE 语句）
- [x] SQL 分析功能（性能评估、优化建议、警告提示）
- [x] 多语言支持（国际化）- 支持中文和英文，语言切换功能
- [x] 表名右键「查询」打开 SQL 查询窗口并预填该表 SQL
- [x] 查询窗口当前库/表展示、SQL/结果区可拖拽比例（默认 1/3、2/3）
- [x] 切换标签后恢复查询窗口的 SQL 与查询结果
- [x] 查询执行超时与错误提示（避免一直运行中）
- [x] 亮色/暗色主题切换与持久化
- [x] 标题栏最大化/还原按钮
- [x] 执行计划可视化（MySQL / PostgreSQL EXPLAIN，流程图展示、全表扫描/索引标注）
- [x] 实时监控（MySQL：活跃连接数、进程列表、慢查询高亮，Wails 事件推送）
- [x] 导入 Navicat 连接文件（.ncx，自动创建 MySQL/SQLite 连接）
- [x] 删除连接前确认（是/否）

### 已完成功能详情 ✅

#### P0 - 高优先级（核心功能完善）

- [x] **SQL Insert 导出格式实现** ✅
  - 实现 SQL INSERT 语句格式导出
  - 支持批量 INSERT 语句生成
  - 处理特殊字符转义

- [x] **连接持久化存储** ✅
  - 连接信息保存到本地文件（JSON格式，存储在用户配置目录）
  - 应用启动时自动加载已保存的连接
  - 连接密码 AES-256 加密存储

#### P1 - 中高优先级（常用功能）

- [x] **查询历史记录** ✅
  - 保存最近执行的 SQL 查询（自动保存，最多保留 100 条）
  - 历史记录列表展示（侧边栏面板）
  - 快速选择历史查询重新执行（点击历史记录自动填充到编辑器）
  - 历史记录搜索功能（支持按 SQL 内容搜索）
  - 按连接 ID 筛选历史记录
  - 显示查询执行时间、行数、成功/失败状态
  - 清除历史记录功能

#### P2 - 中优先级（增强功能）

- [x] **数据导入功能** ✅
  - 支持 CSV 文件导入
  - 支持 JSON 文件导入
  - 导入数据预览和映射（前 10 行预览）
  - 列映射功能（文件列 -> 表列）
  - 批量插入数据（每批 100 条）
  - 导入结果反馈

#### P3 - 低优先级（高级功能）

- [x] **表结构设计器** ✅
  - 可视化表结构编辑
  - 字段类型选择（支持常见数据类型）
  - 索引和外键管理
  - 生成 CREATE TABLE 语句
  - 支持主键、唯一约束、NULL 约束
  - 外键级联操作配置

- [x] **SQL 分析功能** ✅
  - SQL 查询类型识别（SELECT/INSERT/UPDATE/DELETE）
  - 性能评估（复杂度分析）
  - 优化建议生成（常见问题检测）
  - 警告提示（如缺少 WHERE 条件、SELECT * 等）
  - 索引使用建议

## 📚 文档

- [用户指南](docs/user-guide.md)：快速入门、连接配置、查询与执行计划、备份恢复、常见问题  
- [开发者文档](docs/development.md)：架构概览、本地构建与运行、目录说明、贡献指南  
- 开发与迭代计划：[docs/development-plan.md](docs/development-plan.md)、[docs/iteration-plan.md](docs/iteration-plan.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[GNU General Public License v3.0](LICENSE) (GPLv3)

## 🔗 相关链接

- [Wails 文档](https://wails.io/docs)
- [Vue 3 文档](https://vuejs.org/)
- [Tailwind CSS 文档](https://tailwindcss.com/docs)

## 支持作者

如果你觉得 Topology 软件对你有帮助，欢迎请作者喝一杯咖啡 ☕

<div style="display: flex; gap: 10px;">
  <img src="docs/alipay.jpg" alt="支付宝" width="200"  height="373"/>
  <img src="docs/wcpay.png" alt="微信支付" width="200" height="373"/>
</div>

---

**最后更新**: 2026-01-25

## 📦 静态资源本地化

所有静态资源（字体、JS、CSS等）已完全本地化，不再依赖外部 CDN：

- ✅ **Monaco Editor**: 使用 `vite-plugin-monaco-editor` 插件，所有 worker 文件打包到本地
- ✅ **字体**: 使用系统字体，无需外部 CDN
- ✅ **依赖库**: 所有依赖通过 npm 安装，构建时打包到本地

**优势**:
- 完全离线可用
- 更快的加载速度
- 更高的稳定性
- 更好的安全性

详细说明请参考: [CDN_LOCALIZATION.md](frontend/CDN_LOCALIZATION.md)

## 🌐 多语言支持

Topology 现已支持多语言界面：

- **中文（简体）** - 默认语言
- **English** - 完整英文界面

### 语言切换

点击标题栏右侧的地球图标 🌐 即可切换语言。语言偏好会自动保存，下次启动应用时会记住您的选择。

### 已国际化的功能模块

- ✅ 连接管理界面
- ✅ SQL 查询编辑器
- ✅ 数据网格和表格
- ✅ 查询历史记录
- ✅ 数据导入/导出
- ✅ 表结构设计器
- ✅ SQL 分析功能
- ✅ 执行计划、实时监控、Navicat 导入
- ✅ 状态栏和提示信息

## 📋 最近更新

### 2026-01-26（执行计划、实时监控、Navicat 导入）

- ✅ **执行计划**：MySQL / PostgreSQL 下可查看 SELECT 的 EXPLAIN 结果，流程图展示节点类型、全表扫描/索引使用、优化建议
- ✅ **实时监控**：连接右键「打开监控」，弹窗展示活跃连接数与进程列表，每 5 秒刷新，慢查询高亮
- ✅ **导入 Navicat 连接**：侧栏「导入 Navicat」选择 .ncx 文件，解析并自动创建 MySQL/SQLite 连接（密码为空需后续编辑）
- ✅ **删除连接**：删除前弹出确认「是否确认删除连接「xxx」？」选是/否
- ✅ **发布**：GoReleaser 校验和文件名含平台（如 `topology_1.3.0_Ubuntu22.04_checksums.txt`），详见 release.md

### 2026-01-25（界面与查询体验）

- ✅ **查询窗口增强**
  - 表名右键「查询」改为打开 SQL 查询窗口并预填该表查询（不再打开表数据视图）
  - 查询工具栏展示当前库/表（如「库 main / 表 realm」）
  - SQL 区与结果区比例可拖拽调节，默认 SQL 1/3、结果 2/3，比例持久化到本地
  - 切换标签后返回查询窗口时，自动恢复该标签的 SQL 与查询结果
  - 查询执行增加 2 分钟超时与错误结果展示，避免一直显示「运行中」
- ✅ **界面与窗口**
  - 标题栏新增最大化/还原按钮（根据状态切换图标）
  - 亮色/暗色主题切换（标题栏太阳/月亮按钮，偏好持久化）
  - 移除查询工具栏右侧无效的 Dialect/Schema 固定文案
  - 查询结果区域默认占一半高度（与 SQL 区 50:50，可拖拽调整）
- ✅ 完成多语言文案：窗口（最大化/还原）、主题（亮色/暗色）

### 2026-01-25（功能与存储）

- ✅ 完成 SQL Insert 导出格式实现
- ✅ 完成连接持久化存储功能
  - 连接信息自动保存到用户配置目录（`~/.config/topology/connections.json`）
  - 应用启动时自动加载已保存的连接
  - 密码使用 AES-256 加密存储
- ✅ 完成查询历史记录功能
  - 自动保存最近执行的 SQL 查询（最多 100 条）
  - 历史记录存储在 `~/.config/topology/query_history.json`
  - 侧边栏历史面板，支持搜索和快速选择
  - 显示查询执行时间、行数、成功/失败状态
  - 支持按连接 ID 筛选历史记录
- ✅ 完成数据导入功能
  - 支持 CSV 和 JSON 文件导入
  - 数据预览功能（显示前 10 行）
  - 列映射功能（文件列映射到表列）
  - 批量插入（每批 100 条记录）
  - 导入结果反馈（成功/失败、插入行数）
- ✅ 完成表结构设计器
  - 可视化表结构编辑界面
  - 支持添加/删除列、索引、外键
  - 字段类型选择（常见数据类型）
  - 主键、唯一约束、NULL 约束配置
  - 外键级联操作配置（RESTRICT/CASCADE/SET NULL）
  - 自动生成 CREATE TABLE 和 CREATE INDEX 语句
- ✅ 完成 SQL 分析功能执行计划、实时监控、Navicat
  - SQL 查询类型自动识别
  - 性能复杂度评估（low/medium/high）
  - 常见问题检测和警告（如缺少 WHERE、SELECT * 等）
  - 优化建议生成（索引使用、查询优化等）
- ✅ 完成多语言支持（国际化）
  - 支持中文（zh-CN）和英文（en-US）
  - 语言切换功能（标题栏语言切换器）
  - 自动检测浏览器语言
  - 语言偏好保存到本地存储
  - 所有界面文本已国际化



