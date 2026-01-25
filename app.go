package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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

type QueryHistory struct {
	ID          string `json:"id"`
	ConnectionID string `json:"connectionId"`
	SQL         string `json:"sql"`
	ExecutedAt  string `json:"executedAt"`
	Success     bool   `json:"success"`
	Duration    int    `json:"duration,omitempty"` // milliseconds
	RowCount    int    `json:"rowCount,omitempty"`
}

var (
	connMu          sync.RWMutex
	mockConnections []Connection
	seedOnce        sync.Once
	connFileOnce    sync.Once
	connFilePath    string
	historyMu      sync.RWMutex
	queryHistory   []QueryHistory
	historyFileOnce sync.Once
	historyFilePath string
	maxHistorySize  = 100 // Keep last 100 queries
)

const (
	connFileName    = "connections.json"
	historyFileName = "query_history.json"
	encKey          = "topology-connection-key-2026" // In production, use a proper key management system
)

func getConnectionsFilePath() string {
	connFileOnce.Do(func() {
		// Use user config directory
		homeDir, err := os.UserConfigDir()
		if err != nil {
			// Fallback to current directory
			homeDir = "."
		}
		appDir := filepath.Join(homeDir, "topology")
		_ = os.MkdirAll(appDir, 0o755)
		connFilePath = filepath.Join(appDir, connFileName)
	})
	return connFilePath
}

func getHistoryFilePath() string {
	historyFileOnce.Do(func() {
		homeDir, err := os.UserConfigDir()
		if err != nil {
			homeDir = "."
		}
		appDir := filepath.Join(homeDir, "topology")
		_ = os.MkdirAll(appDir, 0o755)
		historyFilePath = filepath.Join(appDir, historyFileName)
	})
	return historyFilePath
}

func loadConnectionsFromFile() []Connection {
	filePath := getConnectionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	var connections []Connection
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil
	}
	// Decrypt passwords
	for i := range connections {
		if connections[i].Password != "" {
			if decrypted, err := decryptPassword(connections[i].Password); err == nil {
				connections[i].Password = decrypted
			}
		}
	}
	return connections
}

func saveConnectionsToFile(connections []Connection) error {
	// Create a copy to encrypt passwords
	saveConnections := make([]Connection, len(connections))
	copy(saveConnections, connections)
	for i := range saveConnections {
		if saveConnections[i].Password != "" {
			if encrypted, err := encryptPassword(saveConnections[i].Password); err == nil {
				saveConnections[i].Password = encrypted
			}
		}
	}
	data, err := json.MarshalIndent(saveConnections, "", "  ")
	if err != nil {
		return err
	}
	filePath := getConnectionsFilePath()
	return os.WriteFile(filePath, data, 0o600)
}

func seedTestConnections() {
	seedOnce.Do(func() {
		connMu.Lock()
		defer connMu.Unlock()
		
		// Try to load from file first
		saved := loadConnectionsFromFile()
		if len(saved) > 0 {
			mockConnections = saved
			return
		}

		// Otherwise, use default test connections
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
	// Save to file
	return saveConnectionsToFile(mockConnections)
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

// UpdateConnection updates an existing connection by ID. ID must exist.
func (a *App) UpdateConnection(connJSON string) error {
	var conn Connection
	if err := json.Unmarshal([]byte(connJSON), &conn); err != nil {
		return err
	}
	if conn.ID == "" {
		return fmt.Errorf("connection ID required")
	}
	db.Close(conn.ID)
	connMu.Lock()
	defer connMu.Unlock()
	for i, c := range mockConnections {
		if c.ID == conn.ID {
			conn.CreatedAt = c.CreatedAt
			if conn.Status == "" {
				conn.Status = c.Status
			}
			mockConnections[i] = conn
			// Save to file
			return saveConnectionsToFile(mockConnections)
		}
	}
	return fmt.Errorf("connection not found")
}

// ReconnectConnection closes cached DB for the connection so it reconnects on next use.
func (a *App) ReconnectConnection(id string) error {
	db.Close(id)
	return nil
}

// DeleteConnection deletes a connection by ID
func (a *App) DeleteConnection(id string) error {
	db.Close(id)
	connMu.Lock()
	defer connMu.Unlock()
	for i, c := range mockConnections {
		if c.ID == id {
			mockConnections = append(mockConnections[:i], mockConnections[i+1:]...)
			// Save to file
			return saveConnectionsToFile(mockConnections)
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
	var result string
	var success bool
	var rowCount int
	var elapsed int

	if db.IsSelect(sql) {
		cols, rows, err := db.RawSelect(g, sql)
		elapsed = int(time.Since(start).Milliseconds())
		if err != nil {
			result = mustMarshalResult(nil, nil, 0, elapsed, err.Error())
			success = false
		} else {
			rowCount = len(rows)
			result = mustMarshalResult(cols, rows, rowCount, elapsed, "")
			success = true
		}
	} else {
		affected, err := db.RawExec(g, sql)
		elapsed = int(time.Since(start).Milliseconds())
		if err != nil {
			result = mustMarshalResult(nil, nil, 0, elapsed, err.Error())
			success = false
		} else {
			result = mustMarshalResult(nil, nil, 0, elapsed, "", int(affected))
			success = true
		}
	}

	// Save to history
	saveQueryHistory(connectionID, sql, success, elapsed, rowCount)

	return result
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

func escapeSQLValue(value, driver string) string {
	// Escape single quotes and backslashes
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `''`)
	// Wrap in single quotes
	return "'" + value + "'"
}

// Password encryption/decryption using AES-256
func getEncryptionKey() []byte {
	hash := sha256.Sum256([]byte(encKey))
	return hash[:]
}

func encryptPassword(password string) (string, error) {
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptPassword(encrypted string) (string, error) {
	key := getEncryptionKey()
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// Query history functions
func loadQueryHistory() {
	filePath := getHistoryFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		queryHistory = make([]QueryHistory, 0)
		return
	}
	if err := json.Unmarshal(data, &queryHistory); err != nil {
		queryHistory = make([]QueryHistory, 0)
	}
}

func saveQueryHistory(connectionID, sql string, success bool, duration, rowCount int) {
	historyMu.Lock()
	defer historyMu.Unlock()

	// Load history if not loaded
	if queryHistory == nil {
		loadQueryHistory()
	}

	// Add new history entry
	history := QueryHistory{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		ConnectionID: connectionID,
		SQL:         sql,
		ExecutedAt:  time.Now().Format(time.RFC3339),
		Success:     success,
		Duration:    duration,
		RowCount:    rowCount,
	}
	queryHistory = append([]QueryHistory{history}, queryHistory...)

	// Keep only last maxHistorySize entries
	if len(queryHistory) > maxHistorySize {
		queryHistory = queryHistory[:maxHistorySize]
	}

	// Save to file
	saveHistoryToFile()
}

func saveHistoryToFile() {
	data, err := json.MarshalIndent(queryHistory, "", "  ")
	if err != nil {
		return
	}
	filePath := getHistoryFilePath()
	_ = os.WriteFile(filePath, data, 0o600)
}

// GetQueryHistory returns query history, optionally filtered by connectionID and search term
func (a *App) GetQueryHistory(connectionID, searchTerm string, limit int) string {
	historyMu.RLock()
	defer historyMu.RUnlock()

	// Load history if not loaded
	if queryHistory == nil {
		loadQueryHistory()
	}

	var filtered []QueryHistory
	for _, h := range queryHistory {
		// Filter by connection ID if provided
		if connectionID != "" && h.ConnectionID != connectionID {
			continue
		}
		// Filter by search term if provided
		if searchTerm != "" && !strings.Contains(strings.ToLower(h.SQL), strings.ToLower(searchTerm)) {
			continue
		}
		filtered = append(filtered, h)
		if limit > 0 && len(filtered) >= limit {
			break
		}
	}

	data, err := json.Marshal(filtered)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// ClearQueryHistory clears all query history
func (a *App) ClearQueryHistory() error {
	historyMu.Lock()
	defer historyMu.Unlock()
	queryHistory = make([]QueryHistory, 0)
	filePath := getHistoryFilePath()
	return os.Remove(filePath)
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
	case "sql":
		f, err := os.Create(path)
		if err != nil {
			return exportError(err.Error())
		}
		defer f.Close()
		tbl := db.QualTable(conn.Type, database, tableName)
		// Generate INSERT statements
		for _, r := range rows {
			colNames := make([]string, 0, len(cols))
			values := make([]string, 0, len(cols))
			for _, col := range cols {
				colNames = append(colNames, quoteIdent(conn.Type, col))
				val := r[col]
				if val == nil {
					values = append(values, "NULL")
				} else {
					valStr := escapeSQLValue(fmt.Sprint(val), conn.Type)
					values = append(values, valStr)
				}
			}
			insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
				tbl, strings.Join(colNames, ", "), strings.Join(values, ", "))
			_, _ = f.WriteString(insertSQL)
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
