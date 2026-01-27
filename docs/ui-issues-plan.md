# Topology UI 问题修复计划

本文档基于对当前前端实现的排查，将 UI 问题整理为可执行计划，按优先级与模块划分，便于逐项修复与验收。

---

## 一、概述

| 项目 | 说明 |
|------|------|
| 范围 | 前端 Vue 组件、样式、主题、交互、与设计规范一致性 |
| 参考 | `.cursorrules` 界面设计、`docs/development-plan.md`、现有实现 |
| 状态 | 待修复 |

---

## 二、主题与样式

### 2.1 浅色主题不可用【高】

- **现象**：大量组件硬编码深色色值（`#1e1e1e`、`#252526`、`#333`、`#3c3c3c`、`#37373d` 等），未使用 `theme-bg-*` / `theme-text-*` 或 CSS 变量。
- **影响**：切换浅色主题后，弹窗、表格、侧边栏等仍为深色，亮色主题无法使用。
- **涉及文件**：
  - `ConnectionManager.vue`
  - `DataViewer.vue`、`DataGrid.vue`
  - `Snippets.vue`、`QueryHistory.vue`
  - `TableDesigner.vue`、`DataImporter.vue`
  - `SQLAnalyzer.vue`、`ExecutionPlanViewer.vue`
  - `MainLayout.vue`（删除确认弹窗）
  - `Sidebar.vue`、`QueryConsole.vue` 等
- **计划**：
  - [x] 全局搜索 `#1e1e1e`、`#252526`、`#333`、`#3c3c3c`、`#37373d` 等硬编码色值。
  - [x] 用 `style.css` 中的 `--bg-*`、`--border`、`--text*` 及 `theme-*` 类替换。
  - [x] 为亮色主题补充/校验 `theme-light` 变量，确保对比度与可读性。

### 2.2 Monaco 主题不随应用主题【高】

- **现象**：`QueryConsole.vue` 中 Monaco 写死 `theme: 'vs-dark'`。
- **影响**：浅色主题下 SQL 编辑器仍为深色，与整体不一致。
- **计划**：
  - [x] 在 QueryConsole 中注入 `useTheme`，根据 `theme` 切换 Monaco `theme`（`vs` / `vs-dark`）。
  - [x] 主题切换时调用 `monaco.editor.setTheme(...)` 或重建实例时传入对应 theme。

### 2.3 等宽字体未加载【低】

- **现象**：Tailwind / Monaco 配置了 `JetBrains Mono`、`Fira Code`，但 `index.html` 无字体引入，`assets/fonts` 仅有 Nunito。
- **影响**：SQL / 数据区域实际为系统默认等宽字体，与设计说明不符。
- **计划**：
  - [ ] 在 `index.html` 或全局样式中通过 CDN / `@font-face` 引入 JetBrains Mono 或 Fira Code。
  - [ ] 确认 `font-mono`、Monaco `fontFamily` 使用上述字体。

### 2.4 删除确认弹窗硬编码深色【中】

- **现象**：`MainLayout.vue` 删除连接确认框中使用 `hover:bg-[#37373d]` 等。
- **计划**：
  - [x] 改用 `theme-bg-hover`、`theme-border` 等主题类，确保随主题切换。

---

## 三、布局与设计规范

### 3.1 连接管理器缺「左侧连接列表」【中】

- **规范**：界面 B 要求「左侧列表：展示已保存的连接快照」+ 右侧配置表单。
- **现状**：连接管理器仅弹窗 + 表单，无已保存连接列表。
- **计划**：
  - [ ] 将 ConnectionManager 改为左右分栏布局。
  - [ ] 左侧：已保存连接列表（来自 `connectionService.getConnections`），支持选中、删除等。
  - [ ] 右侧：当前选中连接的配置表单；未选中时可为「新建」表单或空状态。
  - [ ] 底部保留「测试连接」「取消」「确定并连接」等按钮，与现有逻辑衔接。

### 3.2 SQL / 结果区分比与规范不符【中】

- **规范**：上 60% SQL，下 40% 结果。
- **现状**：`QueryConsole.vue` 中 `DEFAULT_SQL_PERCENT = 100/3`，约 33% SQL、67% 结果。
- **计划**：
  - [x] 将默认比例改为 60 / 40（或通过配置项），并更新 `SPLIT_STORAGE_KEY` 的默认值。
  - [x] 保持可拖拽调节，比例限制仍为 15%–85%。

### 3.3 侧边栏宽度范围【低】

- **规范**：宽度 240–300px，可拖拽。
- **现状**：默认 260px，拖拽范围 240–400px。
- **计划**：
  - [ ] 将上限改为 300px，或明确文档与实现一致（例如允许 240–400 并更新规范）。

---

## 四、状态与数据展示

### 4.1 状态栏连接与当前 Tab 不一致【高】

- **现象**：`currentConnection` 仅在新建/连接或加载连接列表时更新；切换 Tab 不随 **当前 Tab 的 connection** 变化。
- **影响**：多连接、多 Tab 时，状态栏显示的 Host/Port/User 可能不是当前 Tab 对应连接。
- **计划**：
  - [x] 在 MainLayout 中根据 `activeTab?.connectionId` 解析出对应 `Connection`。
  - [x] 状态栏展示「当前 Tab 的连接」；无 Tab 时再回退到 `currentConnection`。

### 4.2 状态栏的查询结果信息易误导【高】

- **现象**：`queryResult` 为全局 ref，切换 Tab 后仍显示最近一次执行的查询的耗时/行数。
- **影响**：例如在 Tab A 执行查询后切到 Tab B（表视图），状态栏仍显示 Tab A 的统计。
- **计划**：
  - [x] 状态栏的查询统计仅展示 **当前激活的 Query Tab** 的 `queryResult`。
  - [x] 当前为 Table Tab 或非 Query Tab 时，隐藏或清零查询耗时/行数等。

### 4.3 QueryConsole 传入的 connection 可能错误【高】

- **现象**：QueryConsole 使用 `connection="currentConnection"`，当激活 Tab 属于其他连接时，`currentConnection` 未必是该 Tab 的连接。
- **影响**：SQL 分析、补全等若依赖 `connection`，可能用错连接。
- **计划**：
  - [x] 按 `activeTab.connectionId` 解析 `Connection`，传入 QueryConsole 的 `connection` prop，与状态栏逻辑统一。

---

## 五、功能缺失与半成品

### 5.1 查询结果不支持导出【高】

- **现象**：QueryConsole 的 DataGrid 仅 `@export="... => console.log(...)"`，无实际导出；`dataService.exportData` 仅支持按表导出。
- **影响**：任意 SQL 的查询结果无法导出为 CSV/JSON/SQL。
- **计划**：
  - [x] 方案 A：前端将当前 `queryResult` 转为 CSV/JSON，触发下载（可先不支持 SQL Insert）。
  - [ ] 方案 B：新增后端「导出查询结果」接口，前端传 `connectionId`、`sql`、`format` 等，再接好 DataGrid `@export`。
  - [x] 在 QueryConsole 中实现 `handleExport`，替换 `console.log`，并统一导出成功/失败提示。

### 5.2 查询结果单元格修改未持久化【中】

- **现象**：DataGrid `@update` 仅 `console.log`，未调用更新 API。
- **影响**：在查询结果中编辑单元格不会落库；若设计上查询结果只读，则与可编辑 UI 矛盾。
- **计划**：
  - [ ] 明确产品策略：查询结果是否可编辑。
  - [ ] 若可编辑：需后端支持「按结果集更新」的接口，前端在 DataGrid `@update` 中调用并刷新。
  - [ ] 若只读：在 QueryConsole 的 DataGrid 上禁用编辑（如 `editConfig` 关闭），避免误导。

### 5.3 保存脚本未实现【中】

- **现象**：`saveScript` 仅 `console.log('Save script:', ...)`，无持久化。
- **计划**：
  - [ ] 实现保存到本地文件（如通过 Wails 对话选择路径）或会话/本地存储。
  - [ ] 提供保存成功/失败的用户反馈（如 Naive UI Message）。

### 5.4 密码明文切换未实现【中】

- **规范**：连接配置中密码支持明文/密文切换。
- **现状**：仅 `type="password"`，无切换。
- **计划**：
  - [x] 在 ConnectionManager 密码输入框旁增加显隐切换（如眼睛图标），切换 `type="password"` / `type="text"`。
  - [x] 注意无障碍与安全提示（如默认密文）。

---

## 六、交互与反馈

### 6.1 表导出成功无统一提示【中】

- **现象**：DataViewer 的 `handleExport` 仅 `console.log`；MainLayout 表右键导出用 `alert`。
- **计划**：
  - [ ] 统一使用 Naive UI Message/Notification 等组件做成功/失败提示。
  - [ ] DataViewer 导出成功/失败时同样给出提示，并与 MainLayout 表导出文案风格一致。

### 6.2 错误与加载态反馈不统一【中】

- **现象**：部分用 `alert`（如建表、Navicat 导入、表导出），部分仅 `console.error`；Loading 样式不统一。
- **计划**：
  - [ ] 统一采用 Naive UI 的 Message/Notification/Dialog 做错误与成功提示。
  - [ ] 统一 Loading 组件（如 NSpin 或共用的 loading 样式），替换各处的自定义 spin。

---

## 七、数据网格与表格

### 7.1 DataGrid 筛选与文档一致【低】

- **现状**：列配置了 `filters`、`filterRender`，需确认 vxe-table 表头 Filter 图标与「包含/等于/不为空」等行为符合设计。
- **计划**：
  - [ ] 核对 vxe-table 的 filter 配置与文档，确保表头有 Filter 图标且筛选逻辑正确。
  - [ ] 若有差异，调整配置或文档。

### 7.2 DataViewer 使用主题变量【高】

- **现象**：`DataViewer.vue` 使用 `bg-[#1e1e1e]`、`bg-[#252526]`、`border-[#333]` 等硬编码。
- **计划**：
  - [x] 改为 `theme-bg-content`、`theme-bg-panel`、`theme-border` 等，与 2.1 一并处理。

### 7.3 表视图分页与「加载更多」【低】

- **现状**：有「加载更多」与上一页/下一页，`pageSize` 固定 100。
- **计划**：
  - [ ] 视需求增加每页条数选择、总页数展示。
  - [ ] 明确「加载更多」与分页的配合方式（如仅分页、或分页+加载更多并存）。

---

## 八、其他

### 8.2 无障碍与键盘支持【低】

- **计划**：
  - [ ] 为关键操作（执行、保存、关闭弹窗、切换 Tab）提供快捷键与焦点管理。
  - [ ] 为按钮、输入框等补充 `aria-*` 等无障碍属性，便于键盘与读屏使用。

### 8.3 窗口缩放与响应式【低】

- **计划**：
  - [ ] 小窗口下保证布局不严重错位、关键按钮可触达。
  - [ ] 侧边栏可考虑可折叠或最小宽度，避免挤压主工作区。

---

## 九、修复优先级汇总

| 优先级 | 编号 | 标题 |
|--------|------|------|
| 高 | 2.1 | 浅色主题不可用 |
| 高 | 2.2 | Monaco 主题不随应用主题 |
| 高 | 4.1 | 状态栏连接与当前 Tab 不一致 |
| 高 | 4.2 | 状态栏查询结果信息易误导 |
| 高 | 4.3 | QueryConsole 传入的 connection 可能错误 |
| 高 | 5.1 | 查询结果不支持导出 |
| 高 | 7.2 | DataViewer 使用主题变量 |
| 中 | 2.4 | 删除确认弹窗硬编码深色 |
| 中 | 3.1 | 连接管理器缺左侧连接列表 |
| 中 | 3.2 | SQL/结果区分比与规范不符 |
| 中 | 5.2 | 查询结果单元格修改未持久化 |
| 中 | 5.3 | 保存脚本未实现 |
| 中 | 5.4 | 密码明文切换未实现 |
| 中 | 6.1 | 表导出成功无统一提示 |
| 中 | 6.2 | 错误与加载态反馈不统一 |
| 低 | 2.3 | 等宽字体未加载 |
| 低 | 3.3 | 侧边栏宽度范围 |
| 低 | 7.1 | DataGrid 筛选与文档一致 |
| 低 | 7.3 | 表视图分页与加载更多 |
| 低 | 8.2 | 无障碍与键盘支持 |
| 低 | 8.3 | 窗口缩放与响应式 |

---

## 十、验收建议

- 每完成一项，在对应 `计划` 下的 `[ ]` 中勾选为 `[x]`。
- 高优先级项建议先做，再做中、低优先级。
- 主题相关（2.1、2.2、2.4、7.2）可集中处理，统一跑一遍亮/暗色主题回归。
- 状态栏与 Tab（4.1、4.2、4.3）可一并改，需验证多连接、多 Tab 切换场景。

---

*文档生成时间：2025-01-27*
