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

	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	ID           string `json:"id"`
	ConnectionID string `json:"connectionId"`
	SQL          string `json:"sql"`
	ExecutedAt   string `json:"executedAt"`
	Success      bool   `json:"success"`
	Duration     int    `json:"duration,omitempty"` // milliseconds
	RowCount     int    `json:"rowCount,omitempty"`
}

// Snippet holds a saved SQL fragment with an alias for quick insert.
type Snippet struct {
	ID        string `json:"id"`
	Alias     string `json:"alias"`
	SQL       string `json:"sql"`
	CreatedAt string `json:"createdAt"`
}

// Schema metadata for SQL completion (tables + columns per connection).
type SchemaTableMeta struct {
	Name    string             `json:"name"`
	Columns []SchemaColumnMeta `json:"columns"`
}
type SchemaColumnMeta struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}
type SchemaDBMeta struct {
	Name   string            `json:"name"`
	Tables []SchemaTableMeta `json:"tables"`
}
type SchemaMetadata struct {
	ConnectionID string         `json:"connectionId"`
	Databases    []SchemaDBMeta `json:"databases"`
}

var (
	connMu           sync.RWMutex
	mockConnections  []Connection
	seedOnce         sync.Once
	schemaMetaMu     sync.RWMutex
	schemaMetaCache  = make(map[string]SchemaMetadata)
	connFileOnce     sync.Once
	connFilePath     string
	historyMu        sync.RWMutex
	queryHistory     []QueryHistory
	historyFileOnce  sync.Once
	historyFilePath  string
	maxHistorySize   = 100 // Keep last 100 queries
	snippetsMu       sync.RWMutex
	snippets         []Snippet
	snippetsFileOnce sync.Once
	snippetsFilePath string
)

const (
	connFileName     = "connections.json"
	historyFileName  = "query_history.json"
	snippetsFileName = "snippets.json"
	encKey           = "topology-connection-key-2026" // In production, use a proper key management system
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

func getSnippetsFilePath() string {
	snippetsFileOnce.Do(func() {
		homeDir, err := os.UserConfigDir()
		if err != nil {
			homeDir = "."
		}
		appDir := filepath.Join(homeDir, "topology")
		_ = os.MkdirAll(appDir, 0o755)
		snippetsFilePath = filepath.Join(appDir, snippetsFileName)
	})
	return snippetsFilePath
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

// getOrOpenDB returns a working DB for the connection (and optional session). Uses cache if ping succeeds, otherwise reconnects.
// Empty sessionID uses shared connection per connID; non-empty isolates per tab/session.
func getOrOpenDB(connID, sessionID string) (*gorm.DB, error) {
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
	if g, ok := db.Get(connID, sessionID); ok {
		sqlDB, err := g.DB()
		if err == nil && sqlDB.Ping() == nil {
			return g, nil
		}
		db.Close(connID, sessionID)
	}
	return db.Open(connID, sessionID, driver, dsn)
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
	db.CloseConnection(conn.ID)
	schemaMetaMu.Lock()
	delete(schemaMetaCache, conn.ID)
	schemaMetaMu.Unlock()
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
	db.CloseConnection(id)
	schemaMetaMu.Lock()
	delete(schemaMetaCache, id)
	schemaMetaMu.Unlock()
	return nil
}

// DeleteConnection deletes a connection by ID
func (a *App) DeleteConnection(id string) error {
	db.CloseConnection(id)
	schemaMetaMu.Lock()
	delete(schemaMetaCache, id)
	schemaMetaMu.Unlock()
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

// ExecuteQuery executes a SQL query. sessionID optionally isolates this tab's DB session (e.g. tab id).
func (a *App) ExecuteQuery(connectionID, sessionID, sql string) string {
	conn := getConnByID(connectionID)
	if conn == nil {
		return mustMarshalResult(nil, nil, 0, 0, fmt.Sprintf("connection not found: %s", connectionID))
	}
	g, err := getOrOpenDB(connectionID, sessionID)
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

// ReleaseSession closes the DB session for the given connection and tab/session. Call when a tab is closed so transactions do not leak.
func (a *App) ReleaseSession(connectionID, sessionID string) {
	if sessionID == "" {
		return
	}
	db.Close(connectionID, sessionID)
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

// LoadSchemaMetadata starts a background goroutine to fetch all databases, tables, and columns for the connection.
// When done, caches the result and emits "schema-metadata-ready" with connectionID for the frontend.
func (a *App) LoadSchemaMetadata(connectionID string) {
	go a.loadSchemaMetadataWorker(connectionID)
}

func (a *App) loadSchemaMetadataWorker(connectionID string) {
	meta := SchemaMetadata{ConnectionID: connectionID}
	g, err := getOrOpenDB(connectionID, "")
	if err != nil {
		schemaMetaMu.Lock()
		schemaMetaCache[connectionID] = meta
		schemaMetaMu.Unlock()
		runtime.EventsEmit(a.ctx, "schema-metadata-ready", connectionID)
		return
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		schemaMetaMu.Lock()
		schemaMetaCache[connectionID] = meta
		schemaMetaMu.Unlock()
		runtime.EventsEmit(a.ctx, "schema-metadata-ready", connectionID)
		return
	}
	dbNames, err := db.DatabaseNames(g, conn.Type)
	if err != nil {
		dbNames = nil
	}
	for _, dbName := range dbNames {
		dbMeta := SchemaDBMeta{Name: dbName}
		tableNames, err := db.TableNames(g, conn.Type, dbName)
		if err != nil {
			meta.Databases = append(meta.Databases, dbMeta)
			continue
		}
		for _, tblName := range tableNames {
			tblMeta := SchemaTableMeta{Name: tblName}
			schemaJSON := a.GetTableSchema(connectionID, dbName, tblName, "")
			var ts TableSchema
			if json.Unmarshal([]byte(schemaJSON), &ts) == nil {
				for _, c := range ts.Columns {
					tblMeta.Columns = append(tblMeta.Columns, SchemaColumnMeta{Name: c.Name, Type: c.Type})
				}
			}
			dbMeta.Tables = append(dbMeta.Tables, tblMeta)
		}
		meta.Databases = append(meta.Databases, dbMeta)
	}
	schemaMetaMu.Lock()
	schemaMetaCache[connectionID] = meta
	schemaMetaMu.Unlock()
	runtime.EventsEmit(a.ctx, "schema-metadata-ready", connectionID)
}

// GetSchemaMetadata returns cached schema metadata (JSON) for the connection. Empty object if not loaded yet.
func (a *App) GetSchemaMetadata(connectionID string) string {
	schemaMetaMu.RLock()
	meta, ok := schemaMetaCache[connectionID]
	schemaMetaMu.RUnlock()
	if !ok {
		return `{"connectionId":"` + connectionID + `","databases":[]}`
	}
	data, _ := json.Marshal(meta)
	return string(data)
}

// GetDatabases returns database names for a connection (MySQL: SHOW DATABASES; SQLite: ["main"]). sessionID optional for tab isolation.
func (a *App) GetDatabases(connectionID, sessionID string) string {
	g, err := getOrOpenDB(connectionID, sessionID)
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

// GetTables returns all tables for a connection and database. For SQLite, database is ignored. sessionID optional for tab isolation.
func (a *App) GetTables(connectionID, database, sessionID string) string {
	g, err := getOrOpenDB(connectionID, sessionID)
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

// GetTableData returns table data with pagination. database is optional (MySQL: qualify db.table). sessionID optional for tab isolation.
func (a *App) GetTableData(connectionID, database, tableName string, limit, offset int, sessionID string) string {
	g, err := getOrOpenDB(connectionID, sessionID)
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

// UpdateTableData updates table data in a single transaction. database is optional (MySQL: qualify db.table).
// On any failure, the whole transaction is rolled back. sessionID optional for tab isolation.
func (a *App) UpdateTableData(connectionID, database, tableName, updatesJSON, sessionID string) error {
	var updates []UpdateRecord
	if err := json.Unmarshal([]byte(updatesJSON), &updates); err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}
	g, err := getOrOpenDB(connectionID, sessionID)
	if err != nil {
		return err
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return fmt.Errorf("connection not found")
	}
	tbl := db.QualTable(conn.Type, database, tableName)
	return g.Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			col := quoteIdent(conn.Type, u.Column)
			q := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ? LIMIT 1", tbl, col, col)
			if res := tx.Exec(q, u.NewValue, u.OldValue); res.Error != nil {
				return res.Error
			}
		}
		return nil
	})
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
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		ConnectionID: connectionID,
		SQL:          sql,
		ExecutedAt:   time.Now().Format(time.RFC3339),
		Success:      success,
		Duration:     duration,
		RowCount:     rowCount,
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

func loadSnippets() {
	filePath := getSnippetsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		snippets = make([]Snippet, 0)
		return
	}
	if err := json.Unmarshal(data, &snippets); err != nil {
		snippets = make([]Snippet, 0)
	}
}

func saveSnippetsToFile() {
	data, err := json.MarshalIndent(snippets, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(getSnippetsFilePath(), data, 0o600)
}

// GetSnippets returns all saved SQL snippets (alias + sql) as JSON array.
func (a *App) GetSnippets() string {
	snippetsMu.RLock()
	defer snippetsMu.RUnlock()
	if snippets == nil {
		loadSnippets()
	}
	data, err := json.Marshal(snippets)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// SaveSnippet adds or updates a snippet by alias. If alias exists, the snippet is updated.
func (a *App) SaveSnippet(alias, sql string) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return fmt.Errorf("alias is required")
	}
	snippetsMu.Lock()
	defer snippetsMu.Unlock()
	if snippets == nil {
		loadSnippets()
	}
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	for i := range snippets {
		if snippets[i].Alias == alias {
			snippets[i].SQL = sql
			snippets[i].CreatedAt = time.Now().Format(time.RFC3339)
			saveSnippetsToFile()
			return nil
		}
	}
	snippets = append(snippets, Snippet{ID: id, Alias: alias, SQL: sql, CreatedAt: time.Now().Format(time.RFC3339)})
	saveSnippetsToFile()
	return nil
}

// DeleteSnippet removes a snippet by id.
func (a *App) DeleteSnippet(id string) error {
	snippetsMu.Lock()
	defer snippetsMu.Unlock()
	if snippets == nil {
		loadSnippets()
	}
	for i, s := range snippets {
		if s.ID == id {
			snippets = append(snippets[:i], snippets[i+1:]...)
			saveSnippetsToFile()
			return nil
		}
	}
	return fmt.Errorf("snippet not found: %s", id)
}

// ImportDataPreview parses and returns preview of import data (first 10 rows)
func (a *App) ImportDataPreview(filePath, format string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return mustMarshalPreview(nil, nil, err.Error())
	}

	var columns []string
	var rows []map[string]interface{}

	switch strings.ToLower(format) {
	case "csv":
		cols, rowsData, err := parseCSV(data)
		if err != nil {
			return mustMarshalPreview(nil, nil, err.Error())
		}
		columns = cols
		// Convert to map format
		rows = make([]map[string]interface{}, 0, len(rowsData))
		for _, row := range rowsData {
			rowMap := make(map[string]interface{})
			for i, col := range columns {
				if i < len(row) {
					rowMap[col] = row[i]
				}
			}
			rows = append(rows, rowMap)
		}
	case "json":
		var jsonData struct {
			Columns []string                 `json:"columns"`
			Rows    []map[string]interface{} `json:"rows"`
		}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			// Try array format
			var arrayData []map[string]interface{}
			if err2 := json.Unmarshal(data, &arrayData); err2 != nil {
				return mustMarshalPreview(nil, nil, err.Error())
			}
			if len(arrayData) > 0 {
				// Extract columns from first row
				columns = make([]string, 0, len(arrayData[0]))
				for k := range arrayData[0] {
					columns = append(columns, k)
				}
				rows = arrayData
			}
		} else {
			columns = jsonData.Columns
			rows = jsonData.Rows
		}
	default:
		return mustMarshalPreview(nil, nil, "unsupported format: "+format)
	}

	// Limit to first 10 rows for preview
	previewRows := rows
	if len(previewRows) > 10 {
		previewRows = previewRows[:10]
	}

	return mustMarshalPreview(columns, previewRows, "")
}

func mustMarshalPreview(cols []string, rows []map[string]interface{}, errMsg string) string {
	result := map[string]interface{}{
		"columns": cols,
		"rows":    rows,
		"error":   errMsg,
	}
	data, _ := json.Marshal(result)
	return string(data)
}

func parseCSV(data []byte) ([]string, [][]string, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) == 0 {
		return nil, nil, fmt.Errorf("empty CSV file")
	}
	columns := records[0]
	rows := records[1:]
	return columns, rows, nil
}

// ImportData imports data into a table. sessionID optional for tab isolation.
func (a *App) ImportData(connectionID, database, tableName, filePath, format string, columnMappingJSON, sessionID string) string {
	g, err := getOrOpenDB(connectionID, sessionID)
	if err != nil {
		return importError(err.Error())
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return importError("connection not found")
	}

	// Parse column mapping
	var columnMapping map[string]string
	if columnMappingJSON != "" {
		if err := json.Unmarshal([]byte(columnMappingJSON), &columnMapping); err != nil {
			return importError("invalid column mapping: " + err.Error())
		}
	}

	// Read and parse file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return importError("failed to read file: " + err.Error())
	}

	var columns []string
	var rows []map[string]interface{}

	switch strings.ToLower(format) {
	case "csv":
		cols, rowsData, err := parseCSV(data)
		if err != nil {
			return importError("failed to parse CSV: " + err.Error())
		}
		columns = cols
		rows = make([]map[string]interface{}, 0, len(rowsData))
		for _, row := range rowsData {
			rowMap := make(map[string]interface{})
			for i, col := range columns {
				if i < len(row) {
					rowMap[col] = row[i]
				}
			}
			rows = append(rows, rowMap)
		}
	case "json":
		var jsonData struct {
			Columns []string                 `json:"columns"`
			Rows    []map[string]interface{} `json:"rows"`
		}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			// Try array format
			var arrayData []map[string]interface{}
			if err2 := json.Unmarshal(data, &arrayData); err2 != nil {
				return importError("failed to parse JSON: " + err.Error())
			}
			if len(arrayData) > 0 {
				columns = make([]string, 0, len(arrayData[0]))
				for k := range arrayData[0] {
					columns = append(columns, k)
				}
				rows = arrayData
			}
		} else {
			columns = jsonData.Columns
			rows = jsonData.Rows
		}
	default:
		return importError("unsupported format: " + format)
	}

	// Apply column mapping if provided
	if len(columnMapping) > 0 {
		for i := range rows {
			newRows := make(map[string]interface{})
			for fileCol, dbCol := range columnMapping {
				if val, ok := rows[i][fileCol]; ok {
					newRows[dbCol] = val
				}
			}
			rows[i] = newRows
		}
	}

	// Get table columns to determine insert columns
	tbl := db.QualTable(conn.Type, database, tableName)
	tableCols, err := getTableColumns(g, conn.Type, database, tableName)
	if err != nil {
		return importError("failed to get table columns: " + err.Error())
	}

	// Build INSERT statements and execute in batches
	batchSize := 100
	inserted := 0
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		// Build INSERT statement
		insertCols := make([]string, 0)
		for _, col := range tableCols {
			// Check if this column exists in the data
			for _, row := range batch {
				if _, ok := row[col]; ok {
					insertCols = append(insertCols, col)
					break
				}
			}
		}

		if len(insertCols) == 0 {
			continue
		}

		// Build VALUES clause
		values := make([]string, 0, len(batch))
		for _, row := range batch {
			rowValues := make([]string, 0, len(insertCols))
			for _, col := range insertCols {
				val := row[col]
				if val == nil {
					rowValues = append(rowValues, "NULL")
				} else {
					valStr := escapeSQLValue(fmt.Sprint(val), conn.Type)
					rowValues = append(rowValues, valStr)
				}
			}
			values = append(values, "("+strings.Join(rowValues, ", ")+")")
		}

		quotedCols := make([]string, len(insertCols))
		for i, col := range insertCols {
			quotedCols[i] = quoteIdent(conn.Type, col)
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			tbl, strings.Join(quotedCols, ", "), strings.Join(values, ", "))

		if err := g.Exec(sql).Error; err != nil {
			return importError(fmt.Sprintf("failed to insert batch: %v", err))
		}
		inserted += len(batch)
	}

	result := map[string]interface{}{
		"success":   true,
		"inserted":  inserted,
		"totalRows": len(rows),
	}
	data2, _ := json.Marshal(result)
	return string(data2)
}

func importError(msg string) string {
	result := map[string]interface{}{
		"success": false,
		"error":   msg,
	}
	data, _ := json.Marshal(result)
	return string(data)
}

func getTableColumns(g *gorm.DB, driver, database, tableName string) ([]string, error) {
	var columns []string
	if driver == "mysql" {
		query := "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		if database == "" {
			query = "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
			if err := g.Raw(query, tableName).Scan(&columns).Error; err != nil {
				return nil, err
			}
		} else {
			if err := g.Raw(query, database, tableName).Scan(&columns).Error; err != nil {
				return nil, err
			}
		}
	} else if driver == "sqlite" {
		query := fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(driver, tableName))
		type ColumnInfo struct {
			Name string `gorm:"column:name"`
		}
		var infos []ColumnInfo
		if err := g.Raw(query).Scan(&infos).Error; err != nil {
			return nil, err
		}
		columns = make([]string, len(infos))
		for i, info := range infos {
			columns[i] = info.Name
		}
	}
	return columns, nil
}

// GenerateCreateTableSQL generates CREATE TABLE SQL from TableSchema
func (a *App) GenerateCreateTableSQL(schemaJSON, driver string) string {
	var schema TableSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return fmt.Sprintf("-- Error: %v", err)
	}

	var sql strings.Builder
	sql.WriteString("CREATE TABLE ")
	if driver == "mysql" && schema.Name != "" {
		sql.WriteString(quoteIdent(driver, schema.Name))
	} else {
		sql.WriteString(quoteIdent(driver, schema.Name))
	}
	sql.WriteString(" (\n")

	// Columns
	columnDefs := make([]string, 0, len(schema.Columns))
	for _, col := range schema.Columns {
		colDef := "  " + quoteIdent(driver, col.Name) + " " + col.Type
		if !col.Nullable {
			colDef += " NOT NULL"
		}
		if col.DefaultValue != "" {
			colDef += " DEFAULT " + col.DefaultValue
		}
		if col.IsPrimaryKey {
			colDef += " PRIMARY KEY"
		}
		if col.IsUnique && !col.IsPrimaryKey {
			colDef += " UNIQUE"
		}
		columnDefs = append(columnDefs, colDef)
	}

	// Primary key constraint (if multiple columns)
	pkCols := make([]string, 0)
	for _, col := range schema.Columns {
		if col.IsPrimaryKey {
			pkCols = append(pkCols, quoteIdent(driver, col.Name))
		}
	}
	if len(pkCols) > 1 {
		columnDefs = append(columnDefs, "  PRIMARY KEY ("+strings.Join(pkCols, ", ")+")")
	}

	sql.WriteString(strings.Join(columnDefs, ",\n"))

	// Foreign keys
	if len(schema.ForeignKeys) > 0 {
		sql.WriteString(",\n")
		fkDefs := make([]string, 0, len(schema.ForeignKeys))
		for _, fk := range schema.ForeignKeys {
			fkCols := make([]string, len(fk.Columns))
			for i, col := range fk.Columns {
				fkCols[i] = quoteIdent(driver, col)
			}
			refCols := make([]string, len(fk.ReferencedColumns))
			for i, col := range fk.ReferencedColumns {
				refCols[i] = quoteIdent(driver, col)
			}
			fkDef := fmt.Sprintf("  FOREIGN KEY (%s) REFERENCES %s (%s)",
				strings.Join(fkCols, ", "),
				quoteIdent(driver, fk.ReferencedTable),
				strings.Join(refCols, ", "))
			if fk.OnDelete != "" {
				fkDef += " ON DELETE " + fk.OnDelete
			}
			if fk.OnUpdate != "" {
				fkDef += " ON UPDATE " + fk.OnUpdate
			}
			fkDefs = append(fkDefs, fkDef)
		}
		sql.WriteString(strings.Join(fkDefs, ",\n"))
	}

	sql.WriteString("\n);\n")

	// Indexes (CREATE INDEX statements)
	if len(schema.Indexes) > 0 {
		sql.WriteString("\n")
		for _, idx := range schema.Indexes {
			idxCols := make([]string, len(idx.Columns))
			for i, col := range idx.Columns {
				idxCols[i] = quoteIdent(driver, col)
			}
			if idx.IsUnique {
				sql.WriteString(fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);\n",
					quoteIdent(driver, idx.Name),
					quoteIdent(driver, schema.Name),
					strings.Join(idxCols, ", ")))
			} else {
				sql.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s (%s);\n",
					quoteIdent(driver, idx.Name),
					quoteIdent(driver, schema.Name),
					strings.Join(idxCols, ", ")))
			}
		}
	}

	return sql.String()
}

// AnalyzeSQL provides basic SQL analysis and optimization suggestions
func (a *App) AnalyzeSQL(sql, driver string) string {
	sqlLower := strings.ToLower(strings.TrimSpace(sql))
	analysis := map[string]interface{}{
		"queryType":   "unknown",
		"suggestions": []string{},
		"warnings":    []string{},
		"performance": map[string]interface{}{},
	}

	// Detect query type
	if strings.HasPrefix(sqlLower, "select") {
		analysis["queryType"] = "SELECT"
		// Check for common issues
		if strings.Contains(sqlLower, "select *") {
			analysis["warnings"] = append(analysis["warnings"].([]string), "使用 SELECT * 可能影响性能，建议明确指定需要的列")
		}
		if !strings.Contains(sqlLower, "where") && !strings.Contains(sqlLower, "limit") {
			analysis["warnings"] = append(analysis["warnings"].([]string), "查询没有 WHERE 条件或 LIMIT，可能返回大量数据")
		}
		if strings.Contains(sqlLower, "like '%") {
			analysis["suggestions"] = append(analysis["suggestions"].([]string), "LIKE '%...' 无法使用索引，考虑使用全文搜索或前缀匹配")
		}
		if strings.Contains(sqlLower, "order by") && !strings.Contains(sqlLower, "limit") {
			analysis["warnings"] = append(analysis["warnings"].([]string), "ORDER BY 没有 LIMIT，可能影响性能")
		}
	} else if strings.HasPrefix(sqlLower, "insert") {
		analysis["queryType"] = "INSERT"
		if strings.Contains(sqlLower, "values") && !strings.Contains(sqlLower, "values") {
			analysis["suggestions"] = append(analysis["suggestions"].([]string), "考虑使用批量插入以提高性能")
		}
	} else if strings.HasPrefix(sqlLower, "update") {
		analysis["queryType"] = "UPDATE"
		if !strings.Contains(sqlLower, "where") {
			analysis["warnings"] = append(analysis["warnings"].([]string), "UPDATE 语句缺少 WHERE 条件，将更新所有行！")
		}
	} else if strings.HasPrefix(sqlLower, "delete") {
		analysis["queryType"] = "DELETE"
		if !strings.Contains(sqlLower, "where") {
			analysis["warnings"] = append(analysis["warnings"].([]string), "DELETE 语句缺少 WHERE 条件，将删除所有行！")
		}
	}

	// Performance tips
	perf := map[string]interface{}{
		"estimatedComplexity": "low",
		"indexUsage":          "unknown",
	}
	if strings.Contains(sqlLower, "join") {
		perf["estimatedComplexity"] = "medium"
		perf["indexUsage"] = "建议确保 JOIN 的列上有索引"
	}
	if strings.Contains(sqlLower, "group by") || strings.Contains(sqlLower, "having") {
		perf["estimatedComplexity"] = "high"
	}
	analysis["performance"] = perf

	data, _ := json.Marshal(analysis)
	return string(data)
}

// GetTableSchema returns table schema. database is optional (MySQL: scope by TABLE_SCHEMA). sessionID optional for tab isolation.
func (a *App) GetTableSchema(connectionID, database, tableName, sessionID string) string {
	g, err := getOrOpenDB(connectionID, sessionID)
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

// ExportData exports data from a table. database is optional (MySQL: qualify db.table). sessionID optional for tab isolation.
func (a *App) ExportData(connectionID, database, tableName, format, sessionID string) string {
	g, err := getOrOpenDB(connectionID, sessionID)
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
