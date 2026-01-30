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
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"topology/internal/backup"
	"topology/internal/db"
	"topology/internal/logger"
	"topology/internal/sshtunnel"

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
	logDir := filepath.Join(getAppDir(), "logs")
	if err := logger.Init(logDir); err != nil {
		// non-fatal; app continues without file logging
		_ = err
	} else {
		logger.Info("topology started; log dir %s", logDir)
	}
	go runBackupScheduler()
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

// Navicat NCX XML structures (connections.ncx)
type navicatConnectionsRoot struct {
	XMLName     xml.Name           `xml:"Connections"`
	Connections []navicatConnEntry `xml:"Connection"`
}

type navicatConnEntry struct {
	ConnectionName   string `xml:"ConnectionName,attr"`
	ConnType         string `xml:"ConnType,attr"`
	Host             string `xml:"Host,attr"`
	Port             string `xml:"Port,attr"`
	Database         string `xml:"Database,attr"`
	DatabaseFileName string `xml:"DatabaseFileName,attr"`
	UserName         string `xml:"UserName,attr"`
	SSL              string `xml:"SSL,attr"`
	SSH              string `xml:"SSH,attr"`
	SSH_Host         string `xml:"SSH_Host,attr"`
	SSH_Port         string `xml:"SSH_Port,attr"`
	SSH_UserName     string `xml:"SSH_UserName,attr"`
	SSH_AuthenMethod string `xml:"SSH_AuthenMethod,attr"`
	SSH_PrivateKey   string `xml:"SSH_PrivateKey,attr"`
}

// ImportNavicatResult is the JSON returned by ImportNavicatConnections.
type ImportNavicatResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
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
	Cached        bool                     `json:"cached,omitempty"`
}

// ExecutionPlanNode represents one step in EXPLAIN result for visualization.
type ExecutionPlanNode struct {
	ID            string  `json:"id"`
	ParentID      *string `json:"parentId,omitempty"`
	Type          string  `json:"type"` // Scan, Join, Sort, Filter, etc.
	Label         string  `json:"label"`
	Detail        string  `json:"detail,omitempty"`
	Rows          int64   `json:"rows,omitempty"`
	Cost          string  `json:"cost,omitempty"`
	Extra         string  `json:"extra,omitempty"`
	FullTableScan bool    `json:"fullTableScan"`
	IndexUsed     bool    `json:"indexUsed"`
}

// ExecutionPlanResult is the JSON returned by GetExecutionPlan.
type ExecutionPlanResult struct {
	Nodes   []ExecutionPlanNode `json:"nodes"`
	Summary struct {
		TotalDurationMs int      `json:"totalDurationMs,omitempty"`
		Warnings        []string `json:"warnings,omitempty"`
	} `json:"summary"`
	Error string `json:"error,omitempty"`
}

// IndexSuggestion is one CREATE INDEX suggestion from GetIndexSuggestions.
type IndexSuggestion struct {
	Table       string   `json:"table"`
	Columns     []string `json:"columns,omitempty"`
	CreateIndex string   `json:"createIndex"`
	Reason      string   `json:"reason"`
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

// ProcessItem represents one row from SHOW FULL PROCESSLIST for live monitor.
type ProcessItem struct {
	ID      string `json:"id"`
	User    string `json:"user"`
	Host    string `json:"host"`
	DB      string `json:"db"`
	Command string `json:"command"`
	Time    int    `json:"time"` // seconds
	State   string `json:"state"`
	Info    string `json:"info"`
}

// LiveStatsPayload is emitted to frontend via "live-stats" event for real-time monitor.
type LiveStatsPayload struct {
	ConnectionID     string        `json:"connectionId"`
	ThreadsConnected int           `json:"threadsConnected"`
	ProcessList      []ProcessItem `json:"processList"`
	Error            string        `json:"error,omitempty"`
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
	connMu              sync.RWMutex
	connections         []Connection
	connectionsLoadOnce sync.Once
	schemaMetaMu        sync.RWMutex
	schemaMetaCache     = make(map[string]SchemaMetadata)
	connFileOnce        sync.Once
	connFilePath        string
	historyMu           sync.RWMutex
	queryHistory        []QueryHistory
	historyFileOnce     sync.Once
	historyFilePath     string
	maxHistorySize      = 100 // Keep last 100 queries
	snippetsMu          sync.RWMutex
	snippets            []Snippet
	snippetsFileOnce    sync.Once
	snippetsFilePath    string
	monitorMu           sync.Mutex
	monitorStop         = make(map[string]chan struct{}) // connectionID -> stop channel
	backupMu            sync.Mutex
	backupRecords       []BackupRecord
	backupsFilePath     string
	scheduleMu          sync.Mutex
	backupSchedules     []BackupSchedule
	schedulesFilePath   string
	queryCacheMu        sync.Mutex
	queryCache          = make(map[string]queryCacheEntry)
	queryCacheOrder     []string
	queryCacheHits      int64
	queryCacheMisses    int64
	txMu                sync.Mutex
	activeTx            = make(map[string]*gorm.DB) // key = txKey(connID, sessionID)
)

type queryCacheEntry struct {
	cols     []string
	rows     []map[string]interface{}
	rowCount int
	execMs   int
	at       time.Time
}

var (
	wsRegex       = regexp.MustCompile(`\s+`)
	fromJoinRegex = regexp.MustCompile(`(?i)(?:FROM|JOIN)\s+(?:[\w.]+\.)?(\w+)`)
	whereColRegex = regexp.MustCompile(`\b(\w+)\s*[=<>]`)
	indexHintSkip = map[string]bool{"AND": true, "OR": true, "ON": true, "IN": true, "AS": true, "SELECT": true, "WHERE": true, "JOIN": true, "LEFT": true, "RIGHT": true, "INNER": true, "OUTER": true, "NULL": true}
)

const (
	queryCacheTTL        = 5 * time.Minute
	queryCacheMaxEntries = 100
)

const (
	connFileName     = "connections.json"
	historyFileName  = "query_history.json"
	snippetsFileName = "snippets.json"
	backupsFileName  = "backups.json"
	maxBackupRecords = 50
	encKey           = "topology-connection-key-2026" // In production, use a proper key management system
)

// BackupRecord holds one backup entry for listing and restore.
type BackupRecord struct {
	ConnectionID string `json:"connectionId"`
	Path         string `json:"path"`
	At           string `json:"at"` // ISO8601
}

// BackupSchedule defines a scheduled backup (daily or weekly).
type BackupSchedule struct {
	ConnectionID string `json:"connectionId"`
	Enabled      bool   `json:"enabled"`
	Schedule     string `json:"schedule"` // "daily" | "weekly"
	Time         string `json:"time"`     // "HH:MM" 24h
	Day          int    `json:"day"`      // 0=Sun..6=Sat for weekly
	OutputDir    string `json:"outputDir,omitempty"`
	LastRun      string `json:"lastRun,omitempty"` // RFC3339
}

const (
	schedulesFileName = "backup_schedules.json"
	defaultBackupDir  = "backups"
)

func getAppDir() string {
	home, _ := os.UserConfigDir()
	if home == "" {
		home = "."
	}
	appDir := filepath.Join(home, "topology")
	_ = os.MkdirAll(appDir, 0o755)
	return appDir
}

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

func getBackupsFilePath() string {
	if backupsFilePath == "" {
		backupsFilePath = filepath.Join(getAppDir(), backupsFileName)
	}
	return backupsFilePath
}

func loadBackupRecords() []BackupRecord {
	data, err := os.ReadFile(getBackupsFilePath())
	if err != nil {
		return nil
	}
	var recs []BackupRecord
	_ = json.Unmarshal(data, &recs)
	return recs
}

func saveBackupRecords(recs []BackupRecord) error {
	data, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getBackupsFilePath(), data, 0o644)
}

func appendBackupRecord(connID, path string) {
	backupMu.Lock()
	defer backupMu.Unlock()
	if backupRecords == nil {
		backupRecords = loadBackupRecords()
	}
	rec := BackupRecord{ConnectionID: connID, Path: path, At: time.Now().UTC().Format(time.RFC3339)}
	backupRecords = append(backupRecords, rec)
	if len(backupRecords) > maxBackupRecords {
		backupRecords = backupRecords[len(backupRecords)-maxBackupRecords:]
	}
	_ = saveBackupRecords(backupRecords)
}

func removeBackupRecord(path string) bool {
	backupMu.Lock()
	defer backupMu.Unlock()
	if backupRecords == nil {
		backupRecords = loadBackupRecords()
	}
	for i, r := range backupRecords {
		if r.Path == path {
			backupRecords = append(backupRecords[:i], backupRecords[i+1:]...)
			_ = saveBackupRecords(backupRecords)
			return true
		}
	}
	return false
}

func getSchedulesFilePath() string {
	if schedulesFilePath == "" {
		schedulesFilePath = filepath.Join(getAppDir(), schedulesFileName)
	}
	return schedulesFilePath
}

func loadBackupSchedules() []BackupSchedule {
	data, err := os.ReadFile(getSchedulesFilePath())
	if err != nil {
		return nil
	}
	var s []BackupSchedule
	_ = json.Unmarshal(data, &s)
	return s
}

func saveBackupSchedules(s []BackupSchedule) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getSchedulesFilePath(), data, 0o644)
}

// backupToPath runs backup for connectionID to outputPath, appends record. Caller ensures path is absolute.
func backupToPath(connectionID, outputPath string) error {
	conn := getConnByID(connectionID)
	if conn == nil {
		return fmt.Errorf("connection not found")
	}
	ty := conn.Type
	if ty != "mysql" && ty != "postgresql" && ty != "postgres" && ty != "sqlite" {
		return fmt.Errorf("backup only supported for MySQL, PostgreSQL, SQLite")
	}
	pc := &backup.Conn{
		Type:     ty,
		Host:     conn.Host,
		Port:     conn.Port,
		Username: conn.Username,
		Password: conn.Password,
		Database: conn.Database,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := backup.RunBackup(ctx, pc, outputPath); err != nil {
		return err
	}
	appendBackupRecord(connectionID, outputPath)
	return nil
}

// loadConnectionsFromFile returns (connections, fileExisted). When fileExisted is true, use the result
// (even if empty); when false, use empty list so that explicit "no connections" is respected.
func loadConnectionsFromFile() ([]Connection, bool) {
	filePath := getConnectionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false
	}
	var connections []Connection
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, false
	}
	// Decrypt passwords
	for i := range connections {
		if connections[i].Password != "" {
			if decrypted, err := decryptPassword(connections[i].Password); err == nil {
				connections[i].Password = decrypted
			}
		}
	}
	return connections, true
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

// ensureConnectionsLoaded loads connections from file once; if file is missing or invalid, keeps list empty.
func ensureConnectionsLoaded() {
	connectionsLoadOnce.Do(func() {
		connMu.Lock()
		defer connMu.Unlock()
		saved, fileExisted := loadConnectionsFromFile()
		if fileExisted {
			connections = saved
		} else {
			connections = make([]Connection, 0)
		}
	})
}

func getConnByID(id string) *Connection {
	connMu.RLock()
	defer connMu.RUnlock()
	for i := range connections {
		if connections[i].ID == id {
			c := connections[i]
			return &c
		}
	}
	return nil
}

func buildDSN(c *Connection) (string, error) {
	return db.BuildDSN(c.Type, c.Host, c.Port, c.Username, c.Password, c.Database)
}

// effectiveHostPort returns (host, port) for building DSN. When SSH tunnel is enabled for MySQL, starts tunnel and returns 127.0.0.1:localPort.
func effectiveHostPort(connID string, c *Connection) (host string, port int, err error) {
	host, port = c.Host, c.Port
	if c.Type != "mysql" {
		return host, port, nil
	}
	if c.SSHTunnel == nil || !c.SSHTunnel.Enabled {
		return host, port, nil
	}
	sshPort := c.SSHTunnel.Port
	if sshPort <= 0 {
		sshPort = 22
	}
	localPort, err := sshtunnel.GetOrStart(connID, sshtunnel.Config{
		SSHHost:     c.SSHTunnel.Host,
		SSHPort:     sshPort,
		SSHUser:     c.SSHTunnel.Username,
		SSHPassword: c.SSHTunnel.Password,
		SSHKey:      c.SSHTunnel.PrivateKey,
		DBHost:      c.Host,
		DBPort:      c.Port,
	})
	if err != nil {
		return "", 0, fmt.Errorf("ssh tunnel: %w", err)
	}
	return "127.0.0.1", localPort, nil
}

func txKey(connID, sessionID string) string {
	if sessionID == "" {
		return connID
	}
	return connID + "\x00" + sessionID
}

// getOrOpenDB returns a working DB for the connection (and optional session). Uses cache if ping succeeds, otherwise reconnects.
// When an active transaction exists for conn+session, returns that tx instead.
// Empty sessionID uses shared connection per connID; non-empty isolates per tab/session.
// When SSH tunnel is enabled (MySQL only), DB traffic goes through the tunnel.
func getOrOpenDB(connID, sessionID string) (*gorm.DB, error) {
	txMu.Lock()
	if tx := activeTx[txKey(connID, sessionID)]; tx != nil {
		txMu.Unlock()
		return tx, nil
	}
	txMu.Unlock()

	conn := getConnByID(connID)
	if conn == nil {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}
	driver := conn.Type
	if driver != "mysql" && driver != "sqlite" && driver != "postgresql" && driver != "postgres" {
		return nil, fmt.Errorf("unsupported driver: %s (mysql/postgresql/sqlite)", driver)
	}
	host, port, err := effectiveHostPort(connID, conn)
	if err != nil {
		return nil, err
	}
	dsn, err := db.BuildDSN(driver, host, port, conn.Username, conn.Password, conn.Database)
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

// BeginTx starts a transaction for the given connection and session. Fails if one is already active.
func (a *App) BeginTx(connectionID, sessionID string) error {
	g, err := getOrOpenDB(connectionID, sessionID)
	if err != nil {
		return err
	}
	txMu.Lock()
	defer txMu.Unlock()
	key := txKey(connectionID, sessionID)
	if activeTx[key] != nil {
		return fmt.Errorf("transaction already active")
	}
	tx := g.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	activeTx[key] = tx
	return nil
}

// CommitTx commits the active transaction for the connection+session.
func (a *App) CommitTx(connectionID, sessionID string) error {
	txMu.Lock()
	tx := activeTx[txKey(connectionID, sessionID)]
	delete(activeTx, txKey(connectionID, sessionID))
	txMu.Unlock()
	if tx == nil {
		return fmt.Errorf("no active transaction")
	}
	return tx.Commit().Error
}

// RollbackTx rolls back the active transaction for the connection+session.
func (a *App) RollbackTx(connectionID, sessionID string) error {
	txMu.Lock()
	tx := activeTx[txKey(connectionID, sessionID)]
	delete(activeTx, txKey(connectionID, sessionID))
	txMu.Unlock()
	if tx == nil {
		return fmt.Errorf("no active transaction")
	}
	return tx.Rollback().Error
}

// GetTransactionStatus returns JSON {"active": true|false} for the connection+session.
func (a *App) GetTransactionStatus(connectionID, sessionID string) string {
	txMu.Lock()
	active := activeTx[txKey(connectionID, sessionID)] != nil
	txMu.Unlock()
	out := struct {
		Active bool `json:"active"`
	}{Active: active}
	b, _ := json.Marshal(out)
	return string(b)
}

func clearActiveTxForConnection(connID string) {
	txMu.Lock()
	defer txMu.Unlock()
	prefix := connID + "\x00"
	for k := range activeTx {
		if k == connID || strings.HasPrefix(k, prefix) {
			tx := activeTx[k]
			delete(activeTx, k)
			if tx != nil {
				_ = tx.Rollback().Error
			}
		}
	}
}

// GetConnections returns all database connections
func (a *App) GetConnections() string {
	ensureConnectionsLoaded()
	connMu.RLock()
	list := make([]Connection, len(connections))
	copy(list, connections)
	connMu.RUnlock()
	data, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// CreateConnection creates a new database connection
func (a *App) CreateConnection(connJSON string) error {
	ensureConnectionsLoaded()
	var conn Connection
	if err := json.Unmarshal([]byte(connJSON), &conn); err != nil {
		return err
	}
	conn.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	conn.Status = "disconnected"
	conn.CreatedAt = time.Now().Format(time.RFC3339)
	connMu.Lock()
	connections = append(connections, conn)
	connMu.Unlock()
	return saveConnectionsToFile(connections)
}

// ImportNavicatConnectionsFromDialog opens a file dialog for .ncx, then imports and creates connections.
// Returns same JSON as ImportNavicatConnections; if user cancels the dialog, returns imported=0 and no error.
func (a *App) ImportNavicatConnectionsFromDialog() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Navicat 连接文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "Navicat Connections (*.ncx)", Pattern: "*.ncx"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
	if err != nil || path == "" {
		out, _ := json.Marshal(ImportNavicatResult{})
		return string(out)
	}
	return a.ImportNavicatConnections(path)
}

// ImportNavicatConnections reads a Navicat .ncx file and creates connections for MySQL and SQLite.
// Password is not stored in NCX; imported connections have empty password (user can edit later).
// Returns JSON ImportNavicatResult: imported count, skipped count, and any errors.
func (a *App) ImportNavicatConnections(filePath string) string {
	var result ImportNavicatResult
	data, err := os.ReadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, "read file: "+err.Error())
		out, _ := json.Marshal(result)
		return string(out)
	}
	var root navicatConnectionsRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		result.Errors = append(result.Errors, "parse XML: "+err.Error())
		out, _ := json.Marshal(result)
		return string(out)
	}
	ensureConnectionsLoaded()
	for _, n := range root.Connections {
		connType := strings.ToUpper(strings.TrimSpace(n.ConnType))
		var driver string
		switch connType {
		case "MYSQL":
			driver = "mysql"
		case "SQLITE":
			driver = "sqlite"
		default:
			result.Skipped++
			continue
		}
		name := strings.TrimSpace(n.ConnectionName)
		if name == "" {
			name = n.Host + ":" + n.Port
		}
		port := 0
		if driver == "mysql" {
			if n.Port != "" {
				port, _ = strconv.Atoi(n.Port)
			}
			if port <= 0 {
				port = 3306
			}
		}
		conn := Connection{
			Name:     name,
			Type:     driver,
			Host:     strings.TrimSpace(n.Host),
			Port:     port,
			Username: strings.TrimSpace(n.UserName),
			Password: "",
			Database: strings.TrimSpace(n.Database),
			UseSSL:   strings.ToLower(n.SSL) == "true",
			Status:   "disconnected",
		}
		if driver == "sqlite" {
			if n.DatabaseFileName != "" {
				conn.Database = strings.TrimSpace(n.DatabaseFileName)
			}
			conn.Host = ""
			conn.Port = 0
		}
		if strings.ToLower(n.SSH) == "true" && n.SSH_Host != "" && driver == "mysql" {
			sshPort := 22
			if n.SSH_Port != "" {
				if p, _ := strconv.Atoi(n.SSH_Port); p > 0 {
					sshPort = p
				}
			}
			conn.SSHTunnel = &SSHTunnel{
				Enabled:  true,
				Host:     strings.TrimSpace(n.SSH_Host),
				Port:     sshPort,
				Username: strings.TrimSpace(n.SSH_UserName),
				Password: "",
			}
			if strings.ToUpper(n.SSH_AuthenMethod) == "PUBLICKEY" && n.SSH_PrivateKey != "" {
				conn.SSHTunnel.PrivateKey = strings.TrimSpace(n.SSH_PrivateKey)
			}
		}
		connJSON, _ := json.Marshal(conn)
		if err := a.CreateConnection(string(connJSON)); err != nil {
			result.Errors = append(result.Errors, name+": "+err.Error())
			continue
		}
		result.Imported++
	}
	out, _ := json.Marshal(result)
	return string(out)
}

// TestConnection tests a database connection. When SSH tunnel is enabled, starts a temporary tunnel then closes it.
func (a *App) TestConnection(connJSON string) bool {
	var conn Connection
	if err := json.Unmarshal([]byte(connJSON), &conn); err != nil {
		return false
	}
	driver := conn.Type
	if driver != "mysql" && driver != "sqlite" && driver != "postgresql" && driver != "postgres" {
		return false
	}
	var dsn string
	var err error
	if driver == "mysql" && conn.SSHTunnel != nil && conn.SSHTunnel.Enabled {
		testID := fmt.Sprintf("test-%d", time.Now().UnixNano())
		sshPort := conn.SSHTunnel.Port
		if sshPort <= 0 {
			sshPort = 22
		}
		localPort, tunnelErr := sshtunnel.GetOrStart(testID, sshtunnel.Config{
			SSHHost:     conn.SSHTunnel.Host,
			SSHPort:     sshPort,
			SSHUser:     conn.SSHTunnel.Username,
			SSHPassword: conn.SSHTunnel.Password,
			SSHKey:      conn.SSHTunnel.PrivateKey,
			DBHost:      conn.Host,
			DBPort:      conn.Port,
		})
		if tunnelErr != nil {
			return false
		}
		defer sshtunnel.Stop(testID)
		dsn, err = db.BuildDSN(driver, "127.0.0.1", localPort, conn.Username, conn.Password, conn.Database)
		if err != nil {
			return false
		}
	} else {
		dsn, err = buildDSN(&conn)
		if err != nil {
			return false
		}
	}
	return db.Ping(driver, dsn) == nil
}

// UpdateConnection updates an existing connection by ID. ID must exist.
func (a *App) UpdateConnection(connJSON string) error {
	ensureConnectionsLoaded()
	var conn Connection
	if err := json.Unmarshal([]byte(connJSON), &conn); err != nil {
		return err
	}
	if conn.ID == "" {
		return fmt.Errorf("connection ID required")
	}
	clearActiveTxForConnection(conn.ID)
	db.CloseConnection(conn.ID)
	sshtunnel.Stop(conn.ID)
	schemaMetaMu.Lock()
	delete(schemaMetaCache, conn.ID)
	schemaMetaMu.Unlock()
	connMu.Lock()
	defer connMu.Unlock()
	for i, c := range connections {
		if c.ID == conn.ID {
			conn.CreatedAt = c.CreatedAt
			if conn.Status == "" {
				conn.Status = c.Status
			}
			connections[i] = conn
			return saveConnectionsToFile(connections)
		}
	}
	return fmt.Errorf("connection not found")
}

// ReconnectConnection closes cached DB and SSH tunnel for the connection so it reconnects on next use.
func (a *App) ReconnectConnection(id string) error {
	clearActiveTxForConnection(id)
	db.CloseConnection(id)
	sshtunnel.Stop(id)
	schemaMetaMu.Lock()
	delete(schemaMetaCache, id)
	schemaMetaMu.Unlock()
	return nil
}

// DeleteConnection deletes a connection by ID
func (a *App) DeleteConnection(id string) error {
	ensureConnectionsLoaded()
	clearActiveTxForConnection(id)
	db.CloseConnection(id)
	sshtunnel.Stop(id)
	schemaMetaMu.Lock()
	delete(schemaMetaCache, id)
	schemaMetaMu.Unlock()
	connMu.Lock()
	defer connMu.Unlock()
	for i, c := range connections {
		if c.ID == id {
			connections = append(connections[:i], connections[i+1:]...)
			return saveConnectionsToFile(connections)
		}
	}
	return fmt.Errorf("connection not found")
}

// ExecuteQuery executes a SQL query. sessionID optionally isolates this tab's DB session (e.g. tab id).
// SELECT results are cached by connectionID + normalized SQL; TTL and size limits apply.
func (a *App) ExecuteQuery(connectionID, sessionID, sql string) string {
	conn := getConnByID(connectionID)
	if conn == nil {
		return mustMarshalResult(nil, nil, 0, 0, userFacingError(fmt.Errorf("connection not found: %s", connectionID)).Message)
	}

	if db.IsSelect(sql) {
		key := queryCacheKey(connectionID, sql)
		if ent, hit := queryCacheGet(key); hit {
			queryCacheRecordHit()
			return marshalQueryResultCached(ent.cols, ent.rows, ent.rowCount, ent.execMs, true)
		}
		queryCacheRecordMiss()
	}

	g, err := getOrOpenDB(connectionID, sessionID)
	if err != nil {
		return mustMarshalResult(nil, nil, 0, 0, userFacingError(err).Message)
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
			result = mustMarshalResult(nil, nil, 0, elapsed, userFacingError(err).Message)
			success = false
		} else {
			rowCount = len(rows)
			result = mustMarshalResult(cols, rows, rowCount, elapsed, "")
			success = true
			key := queryCacheKey(connectionID, sql)
			queryCacheSet(key, queryCacheEntry{cols: cols, rows: rows, rowCount: rowCount, execMs: elapsed})
		}
	} else {
		affected, err := db.RawExec(g, sql)
		elapsed = int(time.Since(start).Milliseconds())
		if err != nil {
			result = mustMarshalResult(nil, nil, 0, elapsed, userFacingError(err).Message)
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

// StartMonitor starts a background goroutine that polls MySQL live stats every 5s and emits "live-stats" events.
// Only MySQL is supported. Returns JSON object with "error" key on failure, or "{}" on success.
func (a *App) StartMonitor(connectionID string) string {
	conn := getConnByID(connectionID)
	if conn == nil {
		return `{"error":"connection not found"}`
	}
	if conn.Type != "mysql" {
		return `{"error":"live monitor is only supported for MySQL"}`
	}
	monitorMu.Lock()
	if monitorStop == nil {
		monitorStop = make(map[string]chan struct{})
	}
	if _, running := monitorStop[connectionID]; running {
		monitorMu.Unlock()
		return `{}`
	}
	stopCh := make(chan struct{})
	monitorStop[connectionID] = stopCh
	monitorMu.Unlock()
	go a.liveMonitorWorker(connectionID, stopCh)
	return `{}`
}

// StopMonitor stops the background monitor for the given connection.
func (a *App) StopMonitor(connectionID string) {
	monitorMu.Lock()
	ch, ok := monitorStop[connectionID]
	if ok {
		delete(monitorStop, connectionID)
		close(ch)
	}
	monitorMu.Unlock()
}

const liveMonitorInterval = 5 * time.Second

// liveMonitorWorker polls MySQL for Threads_connected and PROCESSLIST, then emits "live-stats".
func (a *App) liveMonitorWorker(connectionID string, stopCh <-chan struct{}) {
	ticker := time.NewTicker(liveMonitorInterval)
	defer ticker.Stop()
	emit := func(payload LiveStatsPayload) {
		data, _ := json.Marshal(payload)
		runtime.EventsEmit(a.ctx, "live-stats", string(data))
	}
	for {
		payload := LiveStatsPayload{ConnectionID: connectionID}
		conn := getConnByID(connectionID)
		if conn == nil {
			payload.Error = "connection not found"
			emit(payload)
			return
		}
		g, err := getOrOpenDB(connectionID, "")
		if err != nil {
			payload.Error = err.Error()
			emit(payload)
		} else {
			// Threads_connected
			_, rows, err := db.RawSelect(g, "SHOW GLOBAL STATUS LIKE 'Threads_connected'")
			if err == nil && len(rows) > 0 {
				for _, r := range rows {
					for k, v := range r {
						if strings.EqualFold(k, "Value") && v != nil {
							if n, ok := v.(int64); ok {
								payload.ThreadsConnected = int(n)
							} else {
								fmt.Sscanf(fmt.Sprint(v), "%d", &payload.ThreadsConnected)
							}
							break
						}
					}
				}
			}
			// SHOW FULL PROCESSLIST
			_, plRows, err := db.RawSelect(g, "SHOW FULL PROCESSLIST")
			if err == nil {
				getVal := func(row map[string]interface{}, keys ...string) string {
					for _, key := range keys {
						for k, v := range row {
							if strings.EqualFold(k, key) && v != nil {
								return fmt.Sprint(v)
							}
						}
					}
					return ""
				}
				getInt := func(row map[string]interface{}, keys ...string) int {
					s := getVal(row, keys...)
					var n int
					fmt.Sscanf(s, "%d", &n)
					return n
				}
				for _, row := range plRows {
					payload.ProcessList = append(payload.ProcessList, ProcessItem{
						ID:      getVal(row, "Id", "ID"),
						User:    getVal(row, "User", "USER"),
						Host:    getVal(row, "Host", "HOST"),
						DB:      getVal(row, "db", "DB"),
						Command: getVal(row, "Command", "COMMAND"),
						Time:    getInt(row, "Time", "TIME"),
						State:   getVal(row, "State", "STATE"),
						Info:    getVal(row, "Info", "INFO"),
					})
				}
			}
			emit(payload)
		}
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			// next loop
		}
	}
}

// GetExecutionPlan runs EXPLAIN on the given SQL (SELECT only) and returns a structured plan for visualization.
// Only MySQL is supported; SQLite returns error in summary.
func (a *App) GetExecutionPlan(connectionID, sessionID, sql string) string {
	var out ExecutionPlanResult
	conn := getConnByID(connectionID)
	if conn == nil {
		out.Error = "connection not found"
		data, _ := json.Marshal(out)
		return string(data)
	}
	sql = strings.TrimSpace(sql)
	if !db.IsSelect(sql) || strings.HasPrefix(strings.ToUpper(sql), "EXPLAIN") {
		out.Error = "only SELECT queries can be explained"
		data, _ := json.Marshal(out)
		return string(data)
	}
	g, err := getOrOpenDB(connectionID, sessionID)
	if err != nil {
		out.Error = userFacingError(err).Message
		data, _ := json.Marshal(out)
		return string(data)
	}

	switch conn.Type {
	case "mysql":
		explainSQL := "EXPLAIN " + sql
		cols, rows, err := db.RawSelect(g, explainSQL)
		if err != nil {
			out.Error = userFacingError(err).Message
			data, _ := json.Marshal(out)
			return string(data)
		}
		_ = cols
		getVal := func(row map[string]interface{}, keys ...string) string {
			for _, k := range keys {
				for mapK, v := range row {
					if strings.EqualFold(mapK, k) && v != nil {
						return fmt.Sprint(v)
					}
				}
			}
			return ""
		}
		getInt64 := func(row map[string]interface{}, key string) int64 {
			s := getVal(row, key)
			if s == "" {
				return 0
			}
			var n int64
			_, _ = fmt.Sscanf(s, "%d", &n)
			return n
		}
		var warnings []string
		nodes := make([]ExecutionPlanNode, 0, len(rows))
		var lastID *string
		for i, row := range rows {
			id := fmt.Sprintf("%d", i+1)
			typeVal := getVal(row, "type", "Type")
			tableVal := getVal(row, "table", "Table")
			keyVal := getVal(row, "key", "Key")
			extraVal := getVal(row, "extra", "Extra")
			selectType := getVal(row, "select_type", "select_type")
			rowsEst := getInt64(row, "rows")
			fullScan := typeVal == "ALL" || typeVal == "index"
			indexUsed := keyVal != "" && keyVal != "NULL"
			nodeType := "Table"
			if strings.Contains(strings.ToLower(extraVal), "where") {
				nodeType = "Filter"
			}
			if selectType == "SIMPLE" && tableVal != "" {
				nodeType = "Scan"
			}
			label := tableVal
			if label == "" {
				label = typeVal
			}
			node := ExecutionPlanNode{
				ID:            id,
				ParentID:      lastID,
				Type:          nodeType,
				Label:         label,
				Detail:        typeVal,
				Rows:          rowsEst,
				Extra:         extraVal,
				FullTableScan: fullScan,
				IndexUsed:     indexUsed,
			}
			nodes = append(nodes, node)
			lastID = &id
			if fullScan && !indexUsed && tableVal != "" {
				warnings = append(warnings, "Full table scan on '"+tableVal+"'; consider adding an index")
			}
		}
		out.Nodes = nodes
		out.Summary.Warnings = warnings
	case "postgresql", "postgres":
		explainSQL := "EXPLAIN (ANALYZE, VERBOSE, FORMAT JSON) " + sql
		cols, rows, err := db.RawSelect(g, explainSQL)
		if err != nil {
			out.Error = userFacingError(err).Message
			data, _ := json.Marshal(out)
			return string(data)
		}
		if len(rows) == 0 {
			out.Error = "PostgreSQL EXPLAIN returned no rows"
			data, _ := json.Marshal(out)
			return string(data)
		}
		jsonStr := extractPGExplainJSON(rows[0], cols)
		if jsonStr == "" {
			out.Error = "could not extract EXPLAIN JSON from PostgreSQL result"
			data, _ := json.Marshal(out)
			return string(data)
		}
		nodes, warnings, parseErr := parsePGExplainJSON(jsonStr)
		if parseErr != nil {
			out.Error = userFacingError(parseErr).Message
			data, _ := json.Marshal(out)
			return string(data)
		}
		out.Nodes = nodes
		out.Summary.Warnings = warnings
	default:
		out.Error = "execution plan is supported for MySQL and PostgreSQL only"
	}
	data, _ := json.Marshal(out)
	return string(data)
}

// extractPGExplainJSON gets the JSON string from EXPLAIN (FORMAT JSON) result (one row, one column).
func extractPGExplainJSON(row map[string]interface{}, cols []string) string {
	for _, c := range cols {
		if v, ok := row[c]; ok && v != nil {
			switch x := v.(type) {
			case string:
				return x
			case []byte:
				return string(x)
			}
		}
	}
	for _, v := range row {
		if v == nil {
			continue
		}
		switch x := v.(type) {
		case string:
			if strings.HasPrefix(strings.TrimSpace(x), "[") {
				return x
			}
		case []byte:
			s := string(x)
			if strings.HasPrefix(strings.TrimSpace(s), "[") {
				return s
			}
		}
	}
	return ""
}

// parsePGExplainJSON parses PostgreSQL EXPLAIN (FORMAT JSON) output into ExecutionPlanNode list and warnings.
func parsePGExplainJSON(jsonStr string) (nodes []ExecutionPlanNode, warnings []string, err error) {
	var arr []interface{}
	if e := json.Unmarshal([]byte(jsonStr), &arr); e != nil {
		return nil, nil, fmt.Errorf("invalid EXPLAIN JSON: %w", e)
	}
	if len(arr) == 0 {
		return nil, nil, fmt.Errorf("EXPLAIN JSON empty array")
	}
	top, _ := arr[0].(map[string]interface{})
	if top == nil {
		return nil, nil, fmt.Errorf("EXPLAIN JSON invalid structure")
	}
	plan, _ := top["Plan"].(map[string]interface{})
	if plan == nil {
		return nil, nil, fmt.Errorf("EXPLAIN JSON missing Plan")
	}

	nodes = make([]ExecutionPlanNode, 0)
	warnings = make([]string, 0)
	var lastID *string
	idSeq := 0

	var walk func(m map[string]interface{})
	walk = func(m map[string]interface{}) {
		nodeType := getStr(m, "Node Type")
		rel := getStr(m, "Relation Name")
		alias := getStr(m, "Alias")
		planRows := getFloat(m, "Plan Rows")
		actualRows := getFloat(m, "Actual Rows")
		totalCost := getFloat(m, "Total Cost")
		indexName := getStr(m, "Index Name")

		rowsEst := int64(planRows)
		if actualRows > 0 {
			rowsEst = int64(actualRows)
		}
		costStr := ""
		if totalCost > 0 {
			costStr = fmt.Sprintf("%.2f", totalCost)
		}

		fullScan := nodeType == "Seq Scan"
		indexUsed := indexName != "" || strings.Contains(nodeType, "Index")

		ourType := "Table"
		switch {
		case strings.Contains(nodeType, "Scan"):
			ourType = "Scan"
		case strings.Contains(nodeType, "Join") || strings.Contains(nodeType, "Loop"):
			ourType = "Join"
		case strings.Contains(nodeType, "Sort"):
			ourType = "Sort"
		case strings.Contains(nodeType, "Aggregate"):
			ourType = "Aggregate"
		case strings.Contains(nodeType, "Limit"):
			ourType = "Limit"
		}

		label := rel
		if label == "" {
			label = alias
		}
		if label == "" {
			label = nodeType
		}

		idSeq++
		id := fmt.Sprintf("%d", idSeq)
		node := ExecutionPlanNode{
			ID:            id,
			ParentID:      lastID,
			Type:          ourType,
			Label:         label,
			Detail:        nodeType,
			Rows:          rowsEst,
			Cost:          costStr,
			FullTableScan: fullScan,
			IndexUsed:     indexUsed,
		}
		if indexName != "" {
			node.Extra = "Index: " + indexName
		}
		nodes = append(nodes, node)
		lastID = &id

		if fullScan && rel != "" {
			warnings = append(warnings, "Full table scan on '"+rel+"'; consider adding an index")
		}

		subPlans, _ := m["Plans"].([]interface{})
		for _, sp := range subPlans {
			if sub, _ := sp.(map[string]interface{}); sub != nil {
				walk(sub)
			}
		}
	}
	walk(plan)
	return nodes, warnings, nil
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			return x
		case int:
			return float64(x)
		case int64:
			return float64(x)
		case string:
			var f float64
			_, _ = fmt.Sscanf(x, "%f", &f)
			return f
		}
	}
	return 0
}

// ApiError holds a user-facing error code and message for API responses.
type ApiError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

func userFacingError(err error) ApiError {
	if err == nil {
		return ApiError{}
	}
	msg := err.Error()
	low := strings.ToLower(msg)
	switch {
	case strings.Contains(low, "connection not found"):
		return ApiError{Code: "CONNECTION_NOT_FOUND", Message: "Connection not found. It may have been deleted."}
	case strings.Contains(low, "connection refused") || strings.Contains(low, "connect: connection refused") || strings.Contains(low, "connection reset"):
		return ApiError{Code: "CONNECTION_REFUSED", Message: "Cannot connect to database: connection refused. Check host, port, and that the server is running."}
	case strings.Contains(low, "access denied") || (strings.Contains(low, "password") && strings.Contains(low, "failed")) || strings.Contains(low, "authentication failed"):
		return ApiError{Code: "ACCESS_DENIED", Message: "Access denied. Check username and password."}
	case strings.Contains(low, "syntax error") || strings.Contains(low, "syntaxerror") || strings.Contains(low, "unexpected token"):
		return ApiError{Code: "SYNTAX_ERROR", Message: "SQL syntax error. Check your query."}
	case strings.Contains(low, "does not exist") || strings.Contains(low, "relation ") && strings.Contains(low, " does not exist"):
		return ApiError{Code: "NOT_FOUND", Message: msg}
	case strings.Contains(low, "duplicate key") || strings.Contains(low, "unique constraint"):
		return ApiError{Code: "DUPLICATE_KEY", Message: "Duplicate key or unique constraint violation."}
	case strings.Contains(low, "timeout") || strings.Contains(low, "deadline exceeded"):
		return ApiError{Code: "TIMEOUT", Message: "Operation timed out. Try again or simplify the query."}
	default:
		return ApiError{Message: msg}
	}
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

func marshalQueryResultCached(cols []string, rows []map[string]interface{}, rowCount, execMs int, cached bool) string {
	r := QueryResult{Columns: cols, Rows: rows, RowCount: rowCount, ExecutionTime: execMs, Cached: cached}
	data, _ := json.Marshal(r)
	return string(data)
}

func normalizeSQL(sql string) string {
	s := strings.TrimSpace(sql)
	return wsRegex.ReplaceAllString(s, " ")
}

func queryCacheKey(connID, sql string) string {
	return connID + "\x00" + normalizeSQL(sql)
}

func queryCacheGet(key string) (queryCacheEntry, bool) {
	queryCacheMu.Lock()
	defer queryCacheMu.Unlock()
	e, ok := queryCache[key]
	if !ok {
		return queryCacheEntry{}, false
	}
	if time.Since(e.at) > queryCacheTTL {
		delete(queryCache, key)
		for i, k := range queryCacheOrder {
			if k == key {
				queryCacheOrder = append(queryCacheOrder[:i], queryCacheOrder[i+1:]...)
				break
			}
		}
		return queryCacheEntry{}, false
	}
	return e, true
}

func queryCacheSet(key string, e queryCacheEntry) {
	queryCacheMu.Lock()
	defer queryCacheMu.Unlock()
	e.at = time.Now()
	if _, exists := queryCache[key]; exists {
		for i, k := range queryCacheOrder {
			if k == key {
				queryCacheOrder = append(queryCacheOrder[:i], queryCacheOrder[i+1:]...)
				break
			}
		}
	}
	for len(queryCache) >= queryCacheMaxEntries && len(queryCacheOrder) > 0 {
		evict := queryCacheOrder[0]
		queryCacheOrder = queryCacheOrder[1:]
		delete(queryCache, evict)
	}
	queryCache[key] = e
	queryCacheOrder = append(queryCacheOrder, key)
}

func queryCacheStats() (hits, misses int64) {
	queryCacheMu.Lock()
	defer queryCacheMu.Unlock()
	return queryCacheHits, queryCacheMisses
}

func queryCacheRecordHit() {
	queryCacheMu.Lock()
	queryCacheHits++
	queryCacheMu.Unlock()
}

func queryCacheRecordMiss() {
	queryCacheMu.Lock()
	queryCacheMisses++
	queryCacheMu.Unlock()
}

// GetQueryCacheStats returns JSON { "hits": N, "misses": M } for cache hit-rate visibility.
func (a *App) GetQueryCacheStats() string {
	h, m := queryCacheStats()
	out := struct {
		Hits   int64 `json:"hits"`
		Misses int64 `json:"misses"`
	}{Hits: h, Misses: m}
	b, _ := json.Marshal(out)
	return string(b)
}

// ExtractIndexHintTablesAndCols parses SQL for table (FROM/JOIN) and column (WHERE/ON) hints. Used by index suggestions.
func ExtractIndexHintTablesAndCols(sql string) (tables []string, cols []string) {
	norm := wsRegex.ReplaceAllString(strings.TrimSpace(sql), " ")
	for _, m := range fromJoinRegex.FindAllStringSubmatch(norm, -1) {
		if len(m) > 1 && m[1] != "" && !indexHintSkip[strings.ToUpper(m[1])] {
			tables = append(tables, m[1])
		}
	}
	seenCol := make(map[string]bool)
	for _, m := range whereColRegex.FindAllStringSubmatch(norm, -1) {
		if len(m) > 1 && m[1] != "" && !indexHintSkip[strings.ToUpper(m[1])] && !seenCol[m[1]] {
			seenCol[m[1]] = true
			cols = append(cols, m[1])
		}
	}
	return tables, cols
}

func extractIndexHintTablesAndCols(sql string) (tables []string, cols []string) {
	return ExtractIndexHintTablesAndCols(sql)
}

// GetIndexSuggestions runs EXPLAIN on the given SELECT, detects full-table scans, and returns CREATE INDEX suggestions.
// MySQL and PostgreSQL supported. Uses simple SQL parsing to infer tables and WHERE/JOIN columns.
func (a *App) GetIndexSuggestions(connectionID, sessionID, sql string) string {
	var out struct {
		Suggestions []IndexSuggestion `json:"suggestions"`
		Error       string            `json:"error,omitempty"`
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		out.Error = "connection not found"
		b, _ := json.Marshal(out)
		return string(b)
	}
	sql = strings.TrimSpace(sql)
	if !db.IsSelect(sql) || strings.HasPrefix(strings.ToUpper(sql), "EXPLAIN") {
		out.Error = "only SELECT queries can be analyzed for index suggestions"
		b, _ := json.Marshal(out)
		return string(b)
	}
	g, err := getOrOpenDB(connectionID, sessionID)
	if err != nil {
		out.Error = userFacingError(err).Message
		b, _ := json.Marshal(out)
		return string(b)
	}
	_, colsFromSQL := extractIndexHintTablesAndCols(sql)
	driver := conn.Type
	if driver == "postgres" {
		driver = "postgresql"
	}
	quote := func(s string) string { return quoteIdent(driver, s) }

	var fullScanTables []string
	switch conn.Type {
	case "mysql":
		explainSQL := "EXPLAIN " + sql
		_, rows, err := db.RawSelect(g, explainSQL)
		if err != nil {
			out.Error = userFacingError(err).Message
			b, _ := json.Marshal(out)
			return string(b)
		}
		getVal := func(row map[string]interface{}, keys ...string) string {
			for _, k := range keys {
				for mapK, v := range row {
					if strings.EqualFold(mapK, k) && v != nil {
						return strings.TrimSpace(fmt.Sprint(v))
					}
				}
			}
			return ""
		}
		seen := make(map[string]bool)
		for _, row := range rows {
			typeVal := getVal(row, "type", "Type")
			tableVal := getVal(row, "table", "Table")
			keyVal := getVal(row, "key", "Key")
			if (typeVal == "ALL" || typeVal == "index") && (keyVal == "" || strings.EqualFold(keyVal, "NULL")) && tableVal != "" && !seen[tableVal] {
				seen[tableVal] = true
				fullScanTables = append(fullScanTables, tableVal)
			}
		}
	case "postgresql", "postgres":
		explainSQL := "EXPLAIN (ANALYZE, VERBOSE, FORMAT JSON) " + sql
		cols, rows, err := db.RawSelect(g, explainSQL)
		if err != nil {
			out.Error = userFacingError(err).Message
			b, _ := json.Marshal(out)
			return string(b)
		}
		if len(rows) == 0 {
			b, _ := json.Marshal(out)
			return string(b)
		}
		jsonStr := extractPGExplainJSON(rows[0], cols)
		if jsonStr == "" {
			b, _ := json.Marshal(out)
			return string(b)
		}
		nodes, _, parseErr := parsePGExplainJSON(jsonStr)
		if parseErr != nil {
			out.Error = userFacingError(parseErr).Message
			b, _ := json.Marshal(out)
			return string(b)
		}
		seen := make(map[string]bool)
		for _, n := range nodes {
			if n.FullTableScan && n.Label != "" && !seen[n.Label] {
				seen[n.Label] = true
				fullScanTables = append(fullScanTables, n.Label)
			}
		}
	default:
		out.Error = "index suggestions are supported for MySQL and PostgreSQL only"
		b, _ := json.Marshal(out)
		return string(b)
	}

	for _, t := range fullScanTables {
		reason := "Full table scan on '" + t + "'"
		var cols []string
		for _, c := range colsFromSQL {
			cols = append(cols, c)
		}
		var createIndex string
		if len(cols) > 0 {
			var idxCols []string
			for _, c := range cols {
				idxCols = append(idxCols, quote(c))
			}
			idxName := "idx_" + t
			if len(idxName) > 50 {
				idxName = idxName[:50]
			}
			createIndex = fmt.Sprintf("CREATE INDEX %s ON %s (%s);", quote(idxName), quote(t), strings.Join(idxCols, ", "))
		} else {
			createIndex = "-- Consider adding an index on table " + quote(t) + ". Add columns from WHERE/JOIN. Example: CREATE INDEX " + quote("idx_"+t) + " ON " + quote(t) + "(col1, col2);"
		}
		out.Suggestions = append(out.Suggestions, IndexSuggestion{
			Table:       t,
			Columns:     cols,
			CreateIndex: createIndex,
			Reason:      reason,
		})
	}
	b, _ := json.Marshal(out)
	return string(b)
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
	var dbNames []string
	if conn.Type == "postgresql" || conn.Type == "postgres" {
		dbNames, _ = db.SchemaNames(g)
	} else {
		dbNames, _ = db.DatabaseNames(g, conn.Type)
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

// BackupResult is JSON returned by BackupNow.
type BackupResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RestoreResult is JSON returned by RestoreBackup.
type RestoreResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BackupNow opens a save-file dialog, runs mysqldump/pg_dump/sqlite3 .dump, saves to the chosen path, and records the backup. Returns BackupResult JSON.
// SSH tunnel is not supported for backup.
func (a *App) BackupNow(connectionID string) string {
	var out BackupResult
	conn := getConnByID(connectionID)
	if conn == nil {
		out.Error = "connection not found"
		data, _ := json.Marshal(out)
		return string(data)
	}
	ty := conn.Type
	if ty != "mysql" && ty != "postgresql" && ty != "postgres" && ty != "sqlite" {
		out.Error = "backup only supported for MySQL, PostgreSQL, SQLite"
		data, _ := json.Marshal(out)
		return string(data)
	}
	ext := ".sql"
	if ty == "sqlite" {
		ext = ".sql"
	}
	defName := fmt.Sprintf("topology-backup-%s-%s%s", conn.Name, time.Now().Format("20060102-150405"), ext)
	safeName := strings.Map(func(r rune) rune {
		if r == ' ' || r == '/' || r == '\\' || r == ':' {
			return '-'
		}
		return r
	}, defName)

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "保存备份文件",
		DefaultFilename:  safeName,
		DefaultDirectory: getAppDir(),
		Filters: []runtime.FileFilter{
			{DisplayName: "SQL (*.sql)", Pattern: "*.sql"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
	if err != nil || path == "" {
		if path == "" {
			data, _ := json.Marshal(out)
			return string(data)
		}
		out.Error = err.Error()
		data, _ := json.Marshal(out)
		return string(data)
	}
	if err := backupToPath(connectionID, path); err != nil {
		out.Error = userFacingError(err).Message
		data, _ := json.Marshal(out)
		return string(data)
	}
	out.Success = true
	out.Path = path
	data, _ := json.Marshal(out)
	return string(data)
}

// RestoreBackup restores from backupPath using mysql/psql/sqlite3. Call only after user confirmation. Returns RestoreResult JSON.
func (a *App) RestoreBackup(connectionID, backupPath string) string {
	var out RestoreResult
	conn := getConnByID(connectionID)
	if conn == nil {
		out.Error = "connection not found"
		data, _ := json.Marshal(out)
		return string(data)
	}
	ty := conn.Type
	if ty != "mysql" && ty != "postgresql" && ty != "postgres" && ty != "sqlite" {
		out.Error = "restore only supported for MySQL, PostgreSQL, SQLite"
		data, _ := json.Marshal(out)
		return string(data)
	}
	if _, err := os.Stat(backupPath); err != nil {
		out.Error = "backup file not found"
		data, _ := json.Marshal(out)
		return string(data)
	}
	pc := &backup.Conn{
		Type:     ty,
		Host:     conn.Host,
		Port:     conn.Port,
		Username: conn.Username,
		Password: conn.Password,
		Database: conn.Database,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	if err := backup.RunRestore(ctx, pc, backupPath); err != nil {
		out.Error = userFacingError(err).Message
		data, _ := json.Marshal(out)
		return string(data)
	}
	out.Success = true
	data, _ := json.Marshal(out)
	return string(data)
}

// ListBackups returns JSON array of recent backup records for the connection (or all if connectionID is empty). Newest first.
func (a *App) ListBackups(connectionID string) string {
	backupMu.Lock()
	if backupRecords == nil {
		backupRecords = loadBackupRecords()
	}
	recs := make([]BackupRecord, len(backupRecords))
	copy(recs, backupRecords)
	backupMu.Unlock()

	if connectionID != "" {
		filtered := make([]BackupRecord, 0, len(recs))
		for _, r := range recs {
			if r.ConnectionID == connectionID {
				filtered = append(filtered, r)
			}
		}
		recs = filtered
	}
	// reverse so newest first (recs is a copy)
	for i, j := 0, len(recs)-1; i < j; i, j = i+1, j-1 {
		recs[i], recs[j] = recs[j], recs[i]
	}
	data, _ := json.Marshal(recs)
	return string(data)
}

// PickBackupFile opens a file dialog for *.sql and returns the chosen path, or empty if cancelled.
func (a *App) PickBackupFile() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择备份文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQL (*.sql)", Pattern: "*.sql"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
	if err != nil || path == "" {
		return ""
	}
	return path
}

func nextRun(s *BackupSchedule, base time.Time) time.Time {
	parts := strings.SplitN(s.Time, ":", 2)
	if len(parts) != 2 {
		return base
	}
	h, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	m, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return base
	}
	if base.IsZero() {
		base = time.Now()
	}
	candidate := time.Date(base.Year(), base.Month(), base.Day(), h, m, 0, 0, base.Location())
	if s.Schedule == "daily" {
		if base.Before(candidate) {
			return candidate
		}
		return candidate.AddDate(0, 0, 1)
	}
	if s.Schedule == "weekly" {
		for i := 0; i < 8; i++ {
			if int(candidate.Weekday()) == s.Day && !candidate.Before(base) {
				return candidate
			}
			candidate = candidate.AddDate(0, 0, 1)
		}
		return candidate
	}
	return base
}

func runBackupScheduler() {
	tick := time.NewTicker(1 * time.Minute)
	defer tick.Stop()
	for range tick.C {
		scheduleMu.Lock()
		if backupSchedules == nil {
			backupSchedules = loadBackupSchedules()
		}
		schedules := make([]BackupSchedule, len(backupSchedules))
		copy(schedules, backupSchedules)
		scheduleMu.Unlock()

		now := time.Now()
		for i := range schedules {
			s := &schedules[i]
			if !s.Enabled {
				continue
			}
			var lastRun time.Time
			if s.LastRun != "" {
				lastRun, _ = time.Parse(time.RFC3339, s.LastRun)
			}
			nr := nextRun(s, lastRun)
			if !now.Before(nr) && (lastRun.IsZero() || now.Sub(lastRun) > 2*time.Minute) {
				conn := getConnByID(s.ConnectionID)
				if conn == nil {
					continue
				}
				outDir := s.OutputDir
				if outDir == "" {
					outDir = filepath.Join(getAppDir(), defaultBackupDir)
				}
				_ = os.MkdirAll(outDir, 0o755)
				safeName := strings.Map(func(r rune) rune {
					if r == ' ' || r == '/' || r == '\\' || r == ':' {
						return '-'
					}
					return r
				}, conn.Name)
				fname := fmt.Sprintf("%s-%s.sql", safeName, now.Format("20060102-150405"))
				path := filepath.Join(outDir, fname)
				if err := backupToPath(s.ConnectionID, path); err != nil {
					logger.Warn("scheduled backup failed: %v", err)
				} else {
					logger.Info("scheduled backup ok: %s", path)
				}
				schedules[i].LastRun = now.Format(time.RFC3339)
				scheduleMu.Lock()
				backupSchedules = schedules
				_ = saveBackupSchedules(backupSchedules)
				scheduleMu.Unlock()
			}
		}
	}
}

// GetBackupSchedules returns JSON array of backup schedules.
func (a *App) GetBackupSchedules() string {
	scheduleMu.Lock()
	if backupSchedules == nil {
		backupSchedules = loadBackupSchedules()
	}
	out := make([]BackupSchedule, len(backupSchedules))
	copy(out, backupSchedules)
	scheduleMu.Unlock()
	data, _ := json.Marshal(out)
	return string(data)
}

// SetBackupSchedules saves backup schedules from JSON array.
func (a *App) SetBackupSchedules(jsonSchedules string) error {
	var s []BackupSchedule
	if err := json.Unmarshal([]byte(jsonSchedules), &s); err != nil {
		return err
	}
	scheduleMu.Lock()
	backupSchedules = s
	scheduleMu.Unlock()
	return saveBackupSchedules(s)
}

// DeleteBackup removes a backup record and deletes the file. path must match a stored record.
func (a *App) DeleteBackup(path string) string {
	if path == "" {
		out, _ := json.Marshal(map[string]interface{}{"success": false, "error": "path required"})
		return string(out)
	}
	// Check record exists before deleting file
	backupMu.Lock()
	if backupRecords == nil {
		backupRecords = loadBackupRecords()
	}
	found := false
	for _, r := range backupRecords {
		if r.Path == path {
			found = true
			break
		}
	}
	backupMu.Unlock()
	if !found {
		out, _ := json.Marshal(map[string]interface{}{"success": false, "error": "backup not found"})
		return string(out)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		out, _ := json.Marshal(map[string]interface{}{"success": false, "error": err.Error()})
		return string(out)
	}
	if !removeBackupRecord(path) {
		out, _ := json.Marshal(map[string]interface{}{"success": false, "error": "backup not found"})
		return string(out)
	}
	out, _ := json.Marshal(map[string]interface{}{"success": true})
	return string(out)
}

// VerifyBackup returns JSON { "exists": bool, "size": int64 } for the given path.
func (a *App) VerifyBackup(path string) string {
	if path == "" {
		data, _ := json.Marshal(map[string]interface{}{"exists": false, "size": int64(0)})
		return string(data)
	}
	fi, err := os.Stat(path)
	if err != nil {
		data, _ := json.Marshal(map[string]interface{}{"exists": false, "size": int64(0)})
		return string(data)
	}
	data, _ := json.Marshal(map[string]interface{}{"exists": true, "size": fi.Size()})
	return string(data)
}

// GetDatabases returns database names for a connection (MySQL: SHOW DATABASES; PostgreSQL: schema names of current DB; SQLite: ["main"]). sessionID optional for tab isolation.
func (a *App) GetDatabases(connectionID, sessionID string) string {
	g, err := getOrOpenDB(connectionID, sessionID)
	if err != nil {
		return "[]"
	}
	conn := getConnByID(connectionID)
	if conn == nil {
		return "[]"
	}
	var names []string
	if conn.Type == "postgresql" || conn.Type == "postgres" {
		names, err = db.SchemaNames(g)
	} else {
		names, err = db.DatabaseNames(g, conn.Type)
	}
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

// DeleteTableRows deletes rows by matching all columns (or PK columns when available). rowsJSON: []map[string]interface{}.
func (a *App) DeleteTableRows(connectionID, database, tableName, rowsJSON, sessionID string) error {
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(rowsJSON), &rows); err != nil {
		return err
	}
	if len(rows) == 0 {
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
	info, err := db.TableSchema(g, conn.Type, database, tableName)
	if err != nil {
		return err
	}
	var keyCols []string
	for _, c := range info.Columns {
		if c.IsPrimaryKey {
			keyCols = append(keyCols, c.Name)
		}
	}
	if len(keyCols) == 0 {
		for _, c := range info.Columns {
			keyCols = append(keyCols, c.Name)
		}
	}
	tbl := db.QualTable(conn.Type, database, tableName)
	return g.Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			var args []interface{}
			var preds []string
			for _, col := range keyCols {
				v, ok := row[col]
				if !ok {
					return fmt.Errorf("row missing key column %q", col)
				}
				qc := quoteIdent(conn.Type, col)
				preds = append(preds, qc+" = ?")
				args = append(args, v)
			}
			q := fmt.Sprintf("DELETE FROM %s WHERE %s", tbl, strings.Join(preds, " AND "))
			if res := tx.Exec(q, args...); res.Error != nil {
				return res.Error
			}
		}
		return nil
	})
}

// InsertTableRows inserts rows. rowsJSON: []map[string]interface{}. Uses table columns to build INSERT.
func (a *App) InsertTableRows(connectionID, database, tableName, rowsJSON, sessionID string) error {
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(rowsJSON), &rows); err != nil {
		return err
	}
	if len(rows) == 0 {
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
	tableCols, err := getTableColumns(g, conn.Type, database, tableName)
	if err != nil {
		return err
	}
	tbl := db.QualTable(conn.Type, database, tableName)
	batchSize := 100
	return g.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(rows); i += batchSize {
			end := i + batchSize
			if end > len(rows) {
				end = len(rows)
			}
			batch := rows[i:end]
			var insertCols []string
			for _, col := range tableCols {
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
			var values []string
			for _, row := range batch {
				var parts []string
				for _, col := range insertCols {
					v := row[col]
					if v == nil {
						parts = append(parts, "NULL")
					} else {
						parts = append(parts, escapeSQLValue(fmt.Sprint(v), conn.Type))
					}
				}
				values = append(values, "("+strings.Join(parts, ", ")+")")
			}
			quoted := make([]string, len(insertCols))
			for j, c := range insertCols {
				quoted[j] = quoteIdent(conn.Type, c)
			}
			sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tbl, strings.Join(quoted, ", "), strings.Join(values, ", "))
			if err := tx.Exec(sql).Error; err != nil {
				return err
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
	} else if driver == "postgresql" || driver == "postgres" {
		schema := "public"
		if database != "" {
			schema = database
		}
		query := "SELECT column_name FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ordinal_position"
		var raw []struct {
			ColumnName string `gorm:"column:column_name"`
		}
		if err := g.Raw(query, schema, tableName).Scan(&raw).Error; err != nil {
			return nil, err
		}
		columns = make([]string, len(raw))
		for i, r := range raw {
			columns[i] = r.ColumnName
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
