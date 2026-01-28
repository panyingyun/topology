# Topology

<img src="build/appicon.png" width="20%" height="20%">

A cross-platform database management tool built with Wails framework, providing a Navicat-like experience.

![Status](https://img.shields.io/badge/status-in%20development-yellow)
![Go Version](https://img.shields.io/badge/go-1.23.1-blue)
![Vue Version](https://img.shields.io/badge/vue-3.2.37-green)
![License](https://img.shields.io/badge/license-GPLv3-blue)

## ✨ Features

### Implemented ✅

- **Connection Management**
  - Support for MySQL and SQLite (PostgreSQL UI ready, backend in progress)
  - Create, test, and delete connections (with confirmation before delete)
  - **Import Navicat connections**: sidebar "Import Navicat" to select .ncx file; parses and creates MySQL/SQLite connections (passwords empty, edit later)
  - Real-time connection status; connection tree (Connection → Database → Table)
  - Connection context menu: Edit, Refresh, Open Monitor, Delete

- **SQL Query Editor**
  - Monaco Editor, SQL highlighting and auto-completion
  - Ctrl+Enter execute; SQL formatting
  - Real-time results; query history (search, quick select)
  - **Execution plan** (MySQL): toolbar "Execution Plan" for EXPLAIN visualization (full scan/index usage, suggestions)
  - 2‑minute timeout and error display

- **Data Viewing & Editing**
  - High-performance data grid (vxe-table), virtual scrolling
  - Cell edit, change tracking, batch save
  - Header filters; export (CSV, JSON, SQL Insert); import (CSV/JSON with preview and column mapping)
  - Table data viewer with paged load

- **Live Monitor** (MySQL)
  - Connection right-click "Open Monitor" for real-time dashboard
  - Refreshes every 5s: threads connected, process list (SHOW FULL PROCESSLIST)
  - Slow query highlighting (≥ 5s)

- **Multi-Tab Management**
  - Multiple query and table-data tabs; tab reorder; switch/close

- **User Interface**
  - Light/dark theme (persisted)
  - Resizable sidebar; custom title bar; window controls
  - Status bar: connection, query stats, editor line/column
  - i18n: Chinese and English

### Planned 🚧

- PostgreSQL backend support
- More keyboard shortcuts
- Table structure visual editor (edit existing tables)
- Data backup/restore

## 🛠️ Tech Stack

### Frontend
- **Framework**: Vue 3 (Composition API)
- **Language**: TypeScript
- **UI Component Libraries**:
  - [Naive UI](https://www.naiveui.com/) - Base components
  - [vxe-table](https://vxetable.cn/) - High-performance data table
  - [Monaco Editor](https://microsoft.github.io/monaco-editor/) - Code editor
  - [Lucide Vue Next](https://lucide.dev/) - Icon library
- **Styling**: [Tailwind CSS](https://tailwindcss.com/)
- **Build Tool**: [Vite](https://vitejs.dev/)

### Backend
- **Language**: Go 1.23.1
- **Framework**: [Wails v2](https://wails.io/)
- **Architecture**: Frontend-backend separation; backend provides real APIs (connections, query, table data, import/export, execution plan, live monitor)

## 📁 Project Structure

```
topology/
├── app.go                    # Backend interface implementation (mock data)
├── main.go                   # Wails application entry point
├── frontend/
│   ├── src/
│   │   ├── components/       # Vue components
│   │   │   ├── TitleBar.vue      # Custom title bar
│   │   │   ├── Sidebar.vue       # Sidebar
│   │   │   ├── ConnectionTree.vue # Connection tree
│   │   │   ├── TabBar.vue        # Tab bar
│   │   │   ├── StatusBar.vue     # Status bar
│   │   │   └── DataGrid.vue      # Data grid
│   │   ├── views/            # Page views
│   │   │   ├── MainLayout.vue    # Main layout
│   │   │   ├── ConnectionManager.vue # Connection manager
│   │   │   └── QueryConsole.vue  # SQL console
│   │   ├── services/         # Service layer
│   │   │   ├── connectionService.ts
│   │   │   ├── queryService.ts
│   │   │   └── dataService.ts
│   │   ├── composables/      # Composable functions
│   │   │   └── useMonaco.ts
│   │   ├── types/            # TypeScript type definitions
│   │   │   └── index.ts
│   │   ├── App.vue           # Root component
│   │   ├── main.ts           # Entry file
│   │   └── style.css         # Global styles
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   └── package.json
├── docs/                     # Documentation directory
│   └── development-plan.md  # Development plan
└── README.md
```

## 🚀 Quick Start

### Requirements

- Go 1.23.1 or higher
- Node.js 18+ and npm
- Wails CLI (installation instructions below)

### Install Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Install Dependencies

```bash
# Install frontend dependencies
cd frontend
npm install

# Return to project root
cd ..
```

### Development Mode

Start the development server (with hot reload):

```bash
wails dev
```

The development server will run at:
- Desktop application window (primary development method)
- Browser access: http://localhost:34115 (for debugging)

### Build Application

Build production version:

```bash
# Standard build
wails build

# Clean build (recommended)
wails build -clean
```

Build artifacts are located in the `build/bin/` directory.

### Frontend Only Build

If you need to build the frontend separately:

```bash
cd frontend
npm run build
```

## 📖 Usage Guide

### Create Database Connection

1. Click the "NEW CONNECTION" button in the sidebar
2. Select database type (MySQL, PostgreSQL, or SQLite)
3. Fill in connection information (host, port, username, password, etc.)
4. Click "Test Connection" to test the connection
5. Click "Connect" to create the connection

### Execute SQL Query

1. Select a connection in the connection tree
2. The system will automatically create a query tab
3. Write SQL statements in the SQL editor
4. Press `Ctrl+Enter` or click the "EXECUTE" button to execute the query
5. Query results will be displayed in the data grid below

### View Table Data

1. Expand the connection in the connection tree
2. Click on a table name
3. The system will create a table data tab (full functionality to be implemented)

### Edit Data

1. Double-click a cell in the data grid
2. After modifying data, the system will track changes
3. Click "Save Changes" to save modifications

## 🎨 UI Design

- **Theme**: Light and dark (persisted)
- **Primary Color**: `#1677ff` (Tech Blue)
- **Fonts**: 
  - Code: JetBrains Mono, Fira Code
  - UI: System default fonts

## 📝 Development Status

### Completed ✅

- [x] Connection persistence, query history, data import/export (CSV, JSON, SQL Insert)
- [x] Table structure designer, SQL analysis
- [x] Execution plan visualization (MySQL), live monitor (MySQL)
- [x] Import Navicat connections (.ncx)
- [x] Delete connection confirmation
- [x] i18n (zh-CN, en-US), light/dark theme
- [x] Build and GoReleaser release (macOS, Ubuntu 22.04/24.04, Windows)

For detailed development plan, please see [docs/development-plan.md](docs/development-plan.md)

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

[GNU General Public License v3.0](LICENSE) (GPLv3)

## 🔗 Related Links

- [Wails Documentation](https://wails.io/docs)
- [Vue 3 Documentation](https://vuejs.org/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)

---

**Last Updated**: 2026-01-25
