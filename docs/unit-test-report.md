# Topology 单元测试报告

**生成时间**：见下方执行时间  
**命令**：`go test ./... -v -count=1 -cover`

---

## 一、总览

| 项目 | 结果 |
|------|------|
| 测试结果 | **PASS** |
| 总耗时 | ~0.3s（仅 `internal/db` 有用例） |
| 通过 | 21 |
| 跳过 | 1 |
| 失败 | 0 |
| 覆盖率（db 包） | **74.8%** |

---

## 二、包级结果

| 包 | 状态 | 覆盖率 | 说明 |
|----|------|--------|------|
| `topology` | 无测试 | 0% | 主程序入口，无 `*_test.go` |
| `topology/internal/db` | **PASS** | **74.8%** | 配置、DSN、连接池、查询、Schema、集成测试 |
| `topology/internal/sshtunnel` | 无测试 | 0% | 暂无用例 |

---

## 三、用例明细（`internal/db`）

### 3.1 配置与 DSN

| 用例 | 状态 | 说明 |
|------|------|------|
| `TestLoadMySQLTestConfig` | PASS | 解析 `testdb/mysql.url`，校验 host/port/user/password |
| `TestBuildDSN` | PASS | MySQL / SQLite DSN 拼接 |
| `TestSQLiteTestPath` | PASS | `testdb/realm.db` 路径 |
| `TestPingSQLite` | PASS | SQLite Ping |

### 3.2 集成测试（MySQL）

依赖 `testdb` MySQL 容器（`./scripts/local-up.sh`）及 `testdb/mysql.url`。

| 用例 | 状态 | 说明 |
|------|------|------|
| `TestIntegration_PingMySQL` | PASS | MySQL Ping |
| `TestIntegration_OpenMySQL` | PASS | Open + 连接缓存 Get |
| `TestIntegration_RawSelectMySQL` | PASS | `SELECT 1` 结果解析 |
| `TestIntegration_DatabaseNamesMySQL` | PASS | `SHOW DATABASES`，含 `testdb` |
| `TestIntegration_TableNamesMySQL` | PASS | `SHOW TABLES`，临时表 `_topology_itest` |
| `TestIntegration_TableSchemaMySQL` | PASS | `TableSchema` 列信息 |
| `TestIntegration_TableDataMySQL` | PASS | `TableRowCount`、`TableData` 分页 |
| `TestIntegration_RawExecMySQL` | PASS | INSERT / UPDATE / DELETE |

### 3.3 集成测试（SQLite）

依赖 `testdb/realm.db`。

| 用例 | 状态 | 说明 |
|------|------|------|
| `TestIntegration_PingSQLite` | PASS | SQLite Ping |
| `TestIntegration_OpenSQLite` | PASS | Open |
| `TestIntegration_RawSelectSQLite` | PASS | `SELECT 1` |
| `TestIntegration_TableNamesSQLite` | PASS | `sqlite_master` 表列表 |
| `TestIntegration_TableSchemaSQLite` | PASS | `PRAGMA table_info` |
| `TestIntegration_TableDataSQLite` | PASS | 行数、分页数据 |
| `TestIntegration_RawExecSQLite` | PASS | INSERT、DROP |

### 3.4 其他

| 用例 | 状态 | 说明 |
|------|------|------|
| `TestIntegration_CloseAll` | PASS | 关闭全部缓存连接 |
| `TestIntegration_PostgreSQLPlaceholder` | **SKIP** | db 层未支持 PostgreSQL，占位跳过 |
| `TestIsSelect` | PASS | `IsSelect` 对 SELECT/SHOW/INSERT 等判断 |

---

## 四、覆盖率摘要（`internal/db`）

| 文件 | 函数 | 覆盖率 |
|------|------|--------|
| `query.go` | `RawExec` | 100% |
| `query.go` | `IsSelect` | 94.1% |
| `query.go` | `DatabaseNames` | 78.6% |
| `query.go` | `TableNames` | 85.0% |
| `query.go` | `qualTable` | 100% |
| `query.go` | `TableRowCount` | 100% |
| `query.go` | `TableData` | 85.7% |
| `query.go` | `quoteIdent` | 75.0% |
| `schema.go` | `TableSchema` | 80.0% |
| `schema.go` | `mysqlTableSchema` | 64.3% |
| `schema.go` | `sqliteTableSchema` | 80.0% |

---

## 五、复现方式

```bash
cd /path/to/topology
# MySQL 相关用例需先启动 testdb 容器
cd testdb && ./scripts/local-up.sh && cd ..
go test ./... -v -count=1 -cover
```

---

*报告基于 `go test ./... -v -count=1 -cover` 及 `go tool cover -func=coverage.out` 输出整理。*
