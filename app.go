package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
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
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"`
	Database  string    `json:"database,omitempty"`
	UseSSL    bool      `json:"useSSL,omitempty"`
	SSHTunnel *SSHTunnel `json:"sshTunnel,omitempty"`
	Status   string    `json:"status"`
	CreatedAt string   `json:"createdAt,omitempty"`
}

type SSHTunnel struct {
	Enabled   bool   `json:"enabled"`
	Host      string `json:"host,omitempty"`
	Port      int    `json:"port,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
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
	Name        string      `json:"name"`
	Columns     []Column    `json:"columns"`
	Indexes     []Index     `json:"indexes"`
	ForeignKeys []ForeignKey `json:"foreignKeys"`
}

type QueryResult struct {
	Columns      []string                 `json:"columns"`
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"rowCount"`
	ExecutionTime int                     `json:"executionTime,omitempty"`
	AffectedRows int                      `json:"affectedRows,omitempty"`
	Error        string                   `json:"error,omitempty"`
}

type TableData struct {
	Columns   []string                 `json:"columns"`
	Rows      []map[string]interface{} `json:"rows"`
	TotalRows  int                      `json:"totalRows"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"pageSize"`
}

type UpdateRecord struct {
	RowIndex int                    `json:"rowIndex"`
	Column   string                 `json:"column"`
	OldValue interface{}            `json:"oldValue"`
	NewValue interface{}            `json:"newValue"`
}

// Mock data storage
var mockConnections = []Connection{
	{
		ID:       "1",
		Name:     "Production MySQL",
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Database: "production",
		Status:   "connected",
		CreatedAt: time.Now().Format(time.RFC3339),
	},
	{
		ID:       "2",
		Name:     "Staging PostgreSQL",
		Type:     "postgresql",
		Host:     "staging.example.com",
		Port:     5432,
		Username: "postgres",
		Database: "staging",
		Status:   "disconnected",
		CreatedAt: time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
	},
}

// GetConnections returns all database connections
func (a *App) GetConnections() string {
	data, err := json.Marshal(mockConnections)
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
	mockConnections = append(mockConnections, conn)
	return nil
}

// TestConnection tests a database connection
func (a *App) TestConnection(connJSON string) bool {
	var conn Connection
	if err := json.Unmarshal([]byte(connJSON), &conn); err != nil {
		return false
	}
	// Mock: always succeed for localhost
	return conn.Host == "localhost" || conn.Host == "127.0.0.1"
}

// DeleteConnection deletes a connection by ID
func (a *App) DeleteConnection(id string) error {
	for i, conn := range mockConnections {
		if conn.ID == id {
			mockConnections = append(mockConnections[:i], mockConnections[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("connection not found")
}

// ExecuteQuery executes a SQL query
func (a *App) ExecuteQuery(connectionID, sql string) string {
	// Mock query results
	result := QueryResult{
		Columns: []string{"id", "name", "email", "role", "status", "created_at"},
		Rows: []map[string]interface{}{
			{
				"id":         1,
				"name":       "John Doe",
				"email":      "john@example.com",
				"role":       "Admin",
				"status":    "Active",
				"created_at": "2026-01-20 10:00:00",
			},
			{
				"id":         2,
				"name":       "Jane Smith",
				"email":      "jane@example.com",
				"role":       "Developer",
				"status":    "Active",
				"created_at": "2026-01-21 14:30:00",
			},
			{
				"id":         3,
				"name":       "Bob Wilson",
				"email":      "bob@example.com",
				"role":       "User",
				"status":    "Idle",
				"created_at": "2026-01-22 09:15:00",
			},
		},
		RowCount:      1000,
		ExecutionTime: 42,
		AffectedRows:  3,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return `{"columns":[],"rows":[],"rowCount":0,"error":"` + err.Error() + `"}`
	}
	return string(data)
}

// FormatSQL formats a SQL query
func (a *App) FormatSQL(sql string) string {
	// Mock SQL formatting - in real implementation, use a SQL formatter library
	return sql
}

// GetTables returns all tables for a connection
func (a *App) GetTables(connectionID string) string {
	tables := []Table{
		{Name: "users", Type: "table", RowCount: 1250},
		{Name: "orders", Type: "table", RowCount: 3420},
		{Name: "products", Type: "table", RowCount: 560},
		{Name: "logs", Type: "table", RowCount: 8900},
	}

	data, err := json.Marshal(tables)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// GetTableData returns table data with pagination
func (a *App) GetTableData(connectionID, tableName string, limit, offset int) string {
	// Mock table data
	rows := []map[string]interface{}{
		{
			"id":         1,
			"name":       "John Doe",
			"email":      "john@example.com",
			"role":       "Admin",
			"status":    "Active",
			"created_at": "2026-01-20 10:00:00",
		},
		{
			"id":         2,
			"name":       "Jane Smith",
			"email":      "jane@example.com",
			"role":       "Developer",
			"status":    "Active",
			"created_at": "2026-01-21 14:30:00",
		},
		{
			"id":         3,
			"name":       "Bob Wilson",
			"email":      "bob@example.com",
			"role":       "User",
			"status":    "Idle",
			"created_at": "2026-01-22 09:15:00",
		},
		{
			"id":         4,
			"name":       "Alice Brown",
			"email":      "alice@example.com",
			"role":       "User",
			"status":    "Active",
			"created_at": "2026-01-23 11:20:00",
		},
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(rows) {
		start = len(rows)
	}
	if end > len(rows) {
		end = len(rows)
	}

	result := TableData{
		Columns:   []string{"id", "name", "email", "role", "status", "created_at"},
		Rows:      rows[start:end],
		TotalRows: 1000,
		Page:      offset/limit + 1,
		PageSize:  limit,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return `{"columns":[],"rows":[],"totalRows":0,"page":1,"pageSize":100}`
	}
	return string(data)
}

// UpdateTableData updates table data
func (a *App) UpdateTableData(connectionID, tableName, updatesJSON string) error {
	// Mock update - in real implementation, this would update the database
	return nil
}

// GetTableSchema returns table schema
func (a *App) GetTableSchema(connectionID, tableName string) string {
	schema := TableSchema{
		Name: tableName,
		Columns: []Column{
			{
				Name:         "id",
				Type:         "INT",
				Nullable:     false,
				IsPrimaryKey: true,
				IsUnique:     true,
			},
			{
				Name:         "name",
				Type:         "VARCHAR(255)",
				Nullable:     false,
				IsPrimaryKey: false,
				IsUnique:     false,
			},
			{
				Name:         "email",
				Type:         "VARCHAR(255)",
				Nullable:     false,
				IsPrimaryKey: false,
				IsUnique:     true,
			},
			{
				Name:         "role",
				Type:         "VARCHAR(50)",
				Nullable:     true,
				IsPrimaryKey: false,
				IsUnique:     false,
			},
			{
				Name:         "status",
				Type:         "VARCHAR(20)",
				Nullable:     false,
				IsPrimaryKey: false,
				IsUnique:     false,
				DefaultValue: "Active",
			},
			{
				Name:         "created_at",
				Type:         "TIMESTAMP",
				Nullable:     false,
				IsPrimaryKey: false,
				IsUnique:     false,
				DefaultValue: "CURRENT_TIMESTAMP",
			},
		},
		Indexes: []Index{
			{
				Name:     "PRIMARY",
				Columns:  []string{"id"},
				IsUnique: true,
				Type:     "PRIMARY",
			},
			{
				Name:     "idx_email",
				Columns:  []string{"email"},
				IsUnique: true,
				Type:     "UNIQUE",
			},
		},
		ForeignKeys: []ForeignKey{},
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return `{"name":"","columns":[],"indexes":[],"foreignKeys":[]}`
	}
	return string(data)
}

// ExportData exports data from a table
func (a *App) ExportData(connectionID, tableName, format string) string {
	// Mock export - return JSON string of exported data
	result := map[string]interface{}{
		"success":  true,
		"format":   format,
		"filename": tableName + "_export." + format,
		"data":     "Mock exported data",
	}

	data, err := json.Marshal(result)
	if err != nil {
		return `{"success":false,"error":"` + err.Error() + `"}`
	}
	return string(data)
}

// Greet returns a greeting for the given name (kept for compatibility)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
