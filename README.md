# Topology

基于 Wails 的数据库管理 GUI，类似 Navicat。支持连接管理、SQL 编辑与执行、数据查看/编辑、表设计、AI 优化 SQL、导入导出等功能。

## 功能

- **连接管理** — 创建、测试、管理数据库连接（PostgreSQL、MySQL、Redis、MongoDB 等）
- **数据查看与编辑** — 表格展示与内联编辑、增删行、提交变更
- **SQL 查询编辑器** — 带行号、语法高亮，支持执行与保存
- **数据库对象设计** — 表结构编辑、列/索引/外键管理、SQL 预览
- **AI 优化 SQL** — 优化建议、前后对比、一键应用
- **导入导出** — CSV/JSON/Excel/SQL 等格式的导入与导出

当前版本优先完成前端界面，后端接口使用 mock 数据。

## 技术栈

| 层 | 技术 |
|----|------|
| 桌面框架 | [Wails](https://wails.io) v2 |
| 前端 | React 18、TypeScript、Tailwind CSS、React Router |
| 后端 | Go |
| UI 风格 | 新拟态 + 毛玻璃，参考 `docs/` 下设计稿 |

## 项目结构

```
.
├── app.go              # 后端逻辑与 Mock 接口
├── main.go             # Wails 入口
├── frontend/           # React 前端
│   ├── src/
│   │   ├── components/ # 通用组件
│   │   ├── pages/      # 页面（MainPage、Connections、QueryEditor 等）
│   │   ├── services/   # Mock 服务层
│   │   ├── types/      # TS 类型
│   │   └── utils/      # 工具（如 SQL 高亮）
│   └── wailsjs/        # Wails 生成的前端绑定
├── docs/               # 设计稿与开发计划
│   ├── 01/ ~ 05/       # 各功能界面设计
│   └── development-plan.md
├── build/              # 构建输出
└── wails.json          # Wails 配置
```

## 开发

### 环境

- Go 1.23+
- Node.js 18+
- 系统需满足 [Wails 前置要求](https://wails.io/docs/gettingstarted/installation)

### 安装依赖

```bash
# 前端
cd frontend && npm install && cd ..

# 若有 wails CLI，可一键安装
wails doctor
```

### 本地运行

```bash
wails dev
```

启动后会打开桌面窗口，并启用前端热重载。前端 DevServer 默认为 `http://localhost:5173`，Wails 开发服务为 `http://localhost:34115`。

### 构建

```bash
wails build
```

产出的可执行文件在 `build/bin/` 下，具体路径见终端输出。

### 仅构建前端

```bash
cd frontend && npm run build
```

## 配置

- 应用标题、窗口尺寸、资源目录等见 `wails.json`
- 开发计划与实现步骤见 [docs/development-plan.md](docs/development-plan.md)

## 路由

| 路径 | 说明 |
|------|------|
| `/` | 主页面（SQL 编辑器 + 数据视图，参考 docs/01） |
| `/connections` | 连接管理 |
| `/query/:connectionId?` | SQL 查询编辑器 |
| `/table/:connectionId/:tableName` | 表数据查看 |
| `/designer/:connectionId/:tableName?` | 表设计器 |
| `/import-export` | 导入导出 |

## License

见 [LICENSE](LICENSE) 文件。
