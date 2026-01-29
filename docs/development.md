# Topology 开发者文档

## 一、架构概览

Topology 为基于 **Wails v2** 的桌面应用，前端 Vue 3 + TypeScript，后端 Go。

- **前端**：Vue 3（Composition API）、Naive UI、vxe-table、Monaco Editor、Tailwind CSS。通过 `wailsjs` 生成的绑定调用后端接口。
- **后端**：`app.go` 实现业务逻辑，经 Wails 暴露给前端；`internal/db` 负责数据库连接与查询，`internal/backup` 负责备份/恢复，`internal/logger` 负责日志，`internal/sshtunnel` 负责 SSH 隧道。
- **数据流**：连接配置等持久化在用户配置目录（如 `~/.config/topology/`）；前端调用 `GetConnections`、`ExecuteQuery`、`GetTableData` 等接口，后端通过 GORM 访问 MySQL / PostgreSQL / SQLite。

## 二、本地构建与运行

### 环境要求

- Go 1.23+
- Node.js 18+、npm
- Wails CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- 平台相关：Linux 需 `libwebkit2gtk`、`libgtk-3` 等（见项目 README）

### 安装依赖

```bash
cd frontend && npm install && cd ..
```

### 开发模式

```bash
wails dev
# Ubuntu 22.04: make dev_ubuntu2204  # -tags webkit2_40
# Ubuntu 24.04: make dev_ubuntu2404  # -tags webkit2_41
```

热重载支持；可选在浏览器访问开发服务器地址进行调试。

### 构建

```bash
wails build -clean
# 或 make build_ubuntu2204 / build_ubuntu2404 / build-windows
```

产物在 `build/bin/`。

### 测试

```bash
go test ./...
```

前端构建：`cd frontend && npm run build`。

## 三、目录说明

| 路径 | 说明 |
|------|------|
| `app.go` | 后端主逻辑：连接、查询、表数据、导入导出、执行计划、备份恢复、调度等 |
| `main.go` | Wails 入口，绑定 `App`、启动前端 |
| `internal/db` | 数据库连接、元数据查询、表结构、表数据、EXPLAIN 等 |
| `internal/backup` | mysqldump / pg_dump / sqlite3 备份与恢复 |
| `internal/logger` | 分级日志、文件输出 |
| `internal/sshtunnel` | SSH 隧道 |
| `frontend/src/components` | 通用组件（Sidebar、DataGrid、ConnectionTree、备份/恢复弹窗等） |
| `frontend/src/views` | 主布局、连接管理、查询控制台、表数据查看等 |
| `frontend/src/services` | 对后端接口的封装（connection、query、data、backup 等） |
| `frontend/wailsjs/go/main` | Wails 生成的后端绑定，勿手改 |
| `docs/` | 文档（用户指南、迭代计划、开发计划等） |
| `testdb/` | 测试用数据库配置与脚本 |

## 四、贡献指南

1. **获取代码**：Fork 本仓库，克隆到本地。
2. **开发**：在功能分支上修改，遵循现有代码风格；Go 可用 `gofumpt` 格式化。
3. **测试**：修改后端请跑 `go test ./...`；修改前端请确保 `npm run build` 通过。
4. **提交**： commits 信息清晰；可选在 PR 中引用相关 issue 或文档。
5. **PR**：提交至主仓库对应分支，描述变更内容与测试情况，等待维护者 review。

如需扩展数据库支持、备份逻辑或前端交互，可参考 `internal/db`、`internal/backup` 及 `app.go` 中对应接口实现。

---

更多功能说明见 [用户指南](user-guide.md)，迭代与计划见 [iteration-plan.md](iteration-plan.md)。
