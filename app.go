package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"topology/internal/db"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Connection types
type Connection struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Host      string     `json:"host"`
	Port      int        `json:"port"`
	Username  string     `json:"username"`
	Password  string     `json:"password,omitempty"`
	Database  string     `json:"database,omitempty"`
	UseSSL    bool       `json:"useSSL,omitempty"`
	SSHTunnel *SSHTunnel `json:"sshTunnel,omitempty"`
	Status    string     `json:"status"`
	CreatedAt string     `json:"createdAt,omitempty"`
}

type SSHTunnel struct {
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}

type Table struct {
	Name     string `json:"name"`
	Schema   string `json:"schema,omitempty"`
	Type     string `json:"type"`
	RowCount int    `json:"rowCount,omitempty"`
}

type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"defaultValue,omitempty"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
	IsUnique     bool   `json:"isUnique"`
}

type Index struct {
	Name     string   `json:"name"`
	Columns  []string `json:"columns"`
	IsUnique bool     `json:"isUnique"`
	Type     string   `json:"type"`
}

type ForeignKey struct {
	Name              string   `json:"name"`
	Columns           []string `json:"columns"`
	ReferencedTable   string   `json:"referencedTable"`
	ReferencedColumns []string `json:"referencedColumns"`
	OnDelete          string   `json:"onDelete,omitempty"`
	OnUpdate          string   `json:"onUpdate,omitempty"`
}

type TableSchema struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	Indexes     []Index      `json:"indexes"`
	ForeignKeys []ForeignKey `json:"foreignKeys"`
}

type QueryResult struct {
	Columns       []string                 `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int                      `json:"rowCount"`
	ExecutionTime int                      `json:"executionTime,omitempty"`
	AffectedRows  int                      `json:"affectedRows,omitempty"`
	Error         string                   `json:"error,omitempty"`
}

type TableData struct {
	Columns   []string                 `json:"columns"`
	Rows      []map[string]interface{} `json:"rows"`
	TotalRows int                      `json:"totalRows"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"pageSize"`
}

type UpdateRecord struct {
	RowIndex int         `json:"rowIndex"`
	Column   string      `json:"column"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

var (
	connMu          sync.RWMutex
	mockConnections []Connection
	seedOnce        sync.Once
)

func seedTestConnections() {
	seedOnce.Do(func() {
		connMu.Lock()
		defer connMu.Unlock()
		mockConnections = make([]Connection, 0, 2)

		// MySQL from testdb/mysql.url
		if cfg, err := db.LoadMySQLTestConfig("testdb/mysql.url"); err == nil {
			mockConnections = append(mockConnections, Connection{
				ID:        "1",
				Name:      "Test MySQL",
				Type:      "mysql",
				Host:      strings.TrimSpace(cfg.Host),
				Port:      cfg.Port,
				Username:  cfg.Username,
				Password:  cfg.Password,
				Database:  "mysql",
				Status:    "disconnected",
				CreatedAt: time.Now().Format(time.RFC3339),
			})
		} else {
			mockConnections = append(mockConnections, Connection{
				ID:        "1",
				Name:      "Test MySQL (default)",
				Type:      "mysql",
				Host:      "127.0.0.1",
				Port:      3306,
				Username:  "root",
				Password:  "",
				Database:  "mysql",
				Status:    "disconnected",
				CreatedAt: time.Now().Format(time.RFC3339),
			})
		}

		// SQLite from testdb/realm.db
		mockConnections = append(mockConnections, Connection{
			ID:        "2",
			Name:      "Test SQLite",
			Type:      "sqlite",
			Host:      "",
			Port:      0,
			Username:  "",
			Password:  "",
			Database:  db.SQLiteTestPath(),
			Status:    "disconnected",
			CreatedAt: time.Now().Format(time.RFC3339),
		})
	})
}

func getConnByID(id string) *Connection {
	connMu.RLock()
	defer connMu.RUnlock()
	for i := range mockConnections {
		if mockConnections[i].ID == id {
			c := mockConnections[i]
			return &c
		}
	}
	return nil
}

func buildDSN(c *Connection) (string, error) {
	return db.BuildDSN(c.Type, c.Host, c.Port, c.Username, c.Password, c.Database)
}

func getOrOpenDB(connID string) (*gorm.DB, error) {
	conn := getConnByID(connID)
	if conn == nil {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}
	driver := conn.Type
	if driver != "mysql" && driver != "sqlite" {
		return nil, fmt.Errorf("unsupported driver: %s (mysql/sqlite only)", driver)
	}
	dsn, err := buildDSN(conn)
	if err != nil {
		return nil, err
	}
	if g, ok := db.Get(connID); ok {
		return g, nil
	}
	return db.Open(connID, driver, dsn)
}

// GetConnections returns all database connections
func (a *App) GetConnections() string {
	seedTestConnections()
	connMu.RLock()
	list := make([]Connection, len(mockConnections))
	copy(list, mockConnections)
	connMu.RUnlock()
	data, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// CreateConnection creates a new database connection
func (a *App) CreateConnection(connJSON string) error {
	var conn Connection
	if err := json.Unmarshal([]byte(connJSON), &conn); err != nil {
		return err
	}
	conn.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	conn.Status = "disconnected"
	conn.CreatedAt = time.Now().Format(time.RFC3339)
	connMu.Lock()
	mockConnections = append(mockConnections, conn)
	connMu.Unlock()
	return nil
}

// TestConnection tests a database connection
func (a *App) TestConnection(connJSON string) bool {
	var conn Connection
	if err := json.Unmarshal([]byte(connJSON), &conn); err != nil {
		return false
	}
	driver := conn.Type
	if driver != "mysql" && driver != "sqlite" {
		return false
	}
	dsn, err := buildDSN(&conn)
	if err != nil {
		return false
	}
	return db.Ping(driver, dsn) == nil
}

// DeleteConnection deletes a connection by ID
func (a *App) DeleteConnection(id string) error {
	db.Close(id)
	connMu.Lock()
	defer connMu.Unlock()
	for i, c := range mockConnections {
		if c.ID == id {
			mockConnections = append(mockConnections[:i], mockConnections[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("connection not found")
}

// ExecuteQuery executes a SQL query
func (a *App) ExecuteQuery(connectionID, sql string) string {
	conn := getConnByID(connectionID)
	if conn == nil {
		return mustMarshalResult(nil, nil, 0, 0, fmt.Sprintf("connection not found: %s", connectionID))
	}
	g, err := getOrOpenDB(connectionID)
	if err != nil {
		return mustMarshalResult(nil, nil, 0, 0, err.Error())
	}
	start := time.Now()

	if db.IsSelect(sql) {
		cols, rows, err := db.RawSelect(g, sql)
		elapsed := int(time.Since(start).Milliseconds())
		if err != nil {
			return mustMarshalResult(nil, nil, 0, elapsed, err.Error())
		}
		return mustMarshalResult(cols, rows, len(rows), elapsed, "")
	}
	affected, err := db.RawExec(g, sql)
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return mustMarshalResult(nil, nil, 0, elapsed, err.Error())
	}
	return mustMarshalResult(nil, nil, 0, elapsed, "", int(affected))
}

func mustMarshalResult(cols []string, rows []map[string]interface{}, rowCount, execMs int, errMsg string, affected ...int) string {
	r := QueryResult{
		Columns:       cols,
		Rows:          rows,
		RowCount:      rowCount,
		ExecutionTime: execMs,
		Error:         errMsg,
	}
	if len(affected) > 0 {
		r.AffectedRows = affected[0]
	}
	data, _ := json.Marshal(r)
	return string(data)
}

// FormatSQL formats a SQL query (no-op for now)
func (a *App) FormatSQL(sql string) string {
	return sql
}

// GetDatabases returns database names for a connection (MySQL: SHOW DATABASES; SQLite: ["main"]).
func (a *App) GetDatabases(connectionID string) string {
	g, err := getOrOpenDB(connectionID)
	if err != nil {
		return "[]"
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return "[]"
	}
	names, err := db.DatabaseNames(g, conn.Type)
	if err != nil {
		return "[]"
	}
	data, _ := json.Marshal(names)
	return string(data)
}

// GetTables returns all tables for a connection and database. For SQLite, database is ignored.
func (a *App) GetTables(connectionID, database string) string {
	g, err := getOrOpenDB(connectionID)
	if err != nil {
		return "[]"
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return "[]"
	}
	names, err := db.TableNames(g, conn.Type, database)
	if err != nil {
		return "[]"
	}
	tables := make([]Table, 0, len(names))
	for _, n := range names {
		tables = append(tables, Table{Name: n, Type: "table"})
	}
	data, _ := json.Marshal(tables)
	return string(data)
}

// GetTableData returns table data with pagination. database is optional (MySQL: qualify db.table).
func (a *App) GetTableData(connectionID, database, tableName string, limit, offset int) string {
	g, err := getOrOpenDB(connectionID)
	if err != nil {
		return `{"columns":[],"rows":[],"totalRows":0,"page":1,"pageSize":` + fmt.Sprint(limit) + `}`
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return `{"columns":[],"rows":[],"totalRows":0,"page":1,"pageSize":` + fmt.Sprint(limit) + `}`
	}
	cols, rows, total, err := db.TableData(g, conn.Type, database, tableName, limit, offset)
	if err != nil {
		return `{"columns":[],"rows":[],"totalRows":0,"page":1,"pageSize":` + fmt.Sprint(limit) + `}`
	}
	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}
	result := TableData{Columns: cols, Rows: rows, TotalRows: total, Page: page, PageSize: limit}
	data, _ := json.Marshal(result)
	return string(data)
}

// UpdateTableData updates table data. database is optional (MySQL: qualify db.table).
func (a *App) UpdateTableData(connectionID, database, tableName, updatesJSON string) error {
	var updates []UpdateRecord
	if err := json.Unmarshal([]byte(updatesJSON), &updates); err != nil {
		return err
	}
	g, err := getOrOpenDB(connectionID)
	if err != nil {
		return err
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return fmt.Errorf("connection not found")
	}
	tbl := db.QualTable(conn.Type, database, tableName)
	for _, u := range updates {
		col := quoteIdent(conn.Type, u.Column)
		q := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ? LIMIT 1", tbl, col, col)
		if res := g.Exec(q, u.NewValue, u.OldValue); res.Error != nil {
			return res.Error
		}
	}
	return nil
}

func quoteIdent(driver, name string) string {
	if driver == "mysql" {
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// GetTableSchema returns table schema. database is optional (MySQL: scope by TABLE_SCHEMA).
func (a *App) GetTableSchema(connectionID, database, tableName string) string {
	g, err := getOrOpenDB(connectionID)
	if err != nil {
		return `{"name":"","columns":[],"indexes":[],"foreignKeys":[]}`
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return `{"name":"","columns":[],"indexes":[],"foreignKeys":[]}`
	}
	info, err := db.TableSchema(g, conn.Type, database, tableName)
	if err != nil {
		return `{"name":"","columns":[],"indexes":[],"foreignKeys":[]}`
	}
	schema := TableSchema{
		Name:        info.Name,
		Columns:     make([]Column, 0, len(info.Columns)),
		Indexes:     nil,
		ForeignKeys: nil,
	}
	for _, c := range info.Columns {
		schema.Columns = append(schema.Columns, Column{
			Name:         c.Name,
			Type:         c.Type,
			Nullable:     c.Nullable,
			DefaultValue: c.DefaultValue,
			IsPrimaryKey: c.IsPrimaryKey,
			IsUnique:     c.IsUnique,
		})
	}
	data, _ := json.Marshal(schema)
	return string(data)
}

// ExportData exports data from a table. database is optional (MySQL: qualify db.table).
func (a *App) ExportData(connectionID, database, tableName, format string) string {
	g, err := getOrOpenDB(connectionID)
	if err != nil {
		return exportError(err.Error())
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return exportError("connection not found")
	}
	cols, rows, _, err := db.TableData(g, conn.Type, database, tableName, 1<<20, 0)
	if err != nil {
		return exportError(err.Error())
	}
	ext := format
	if ext == "" {
		ext = "json"
	}
	fname := tableName + "_export." + ext
	outDir := filepath.Join("build", "export")
	_ = os.MkdirAll(outDir, 0o755)
	path := filepath.Join(outDir, fname)

	switch strings.ToLower(ext) {
	case "csv":
		f, err := os.Create(path)
		if err != nil {
			return exportError(err.Error())
		}
		defer f.Close()
		w := csv.NewWriter(f)
		_ = w.Write(cols)
		for _, r := range rows {
			rec := make([]string, len(cols))
			for i, c := range cols {
				v := r[c]
				if v != nil {
					rec[i] = fmt.Sprint(v)
				}
			}
			_ = w.Write(rec)
		}
		w.Flush()
		if w.Error() != nil {
			return exportError(w.Error().Error())
		}
	case "json":
		f, err := os.Create(path)
		if err != nil {
			return exportError(err.Error())
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(map[string]interface{}{"columns": cols, "rows": rows}); err != nil {
			return exportError(err.Error())
		}
	default:
		return exportError("unsupported format: " + format)
	}
	data, _ := json.Marshal(map[string]interface{}{
		"success":  true,
		"format":   format,
		"filename": fname,
		"path":     path,
	})
	return string(data)
}

func exportError(msg string) string {
	data, _ := json.Marshal(map[string]interface{}{"success": false, "error": msg})
	return string(data)
}

// Greet returns a greeting for the given name (kept for compatibility)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
