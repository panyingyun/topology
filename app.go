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
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	Database     string `json:"database,omitempty"`
	Status       string `json:"status"`
	Group        string `json:"group,omitempty"`
	SavePassword bool   `json:"savePassword,omitempty"`
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
	IsPrimaryKey bool   `json:"isPrimaryKey"`
	IsNotNull    bool   `json:"isNotNull"`
	IsUnique     bool   `json:"isUnique"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Length       int    `json:"length,omitempty"`
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
	Error         string                   `json:"error,omitempty"`
}

type OptimizedSQL struct {
	Original        string   `json:"original"`
	Optimized       string   `json:"optimized"`
	Suggestions     []string `json:"suggestions"`
	PerformanceGain string   `json:"performanceGain,omitempty"`
}

type TableData struct {
	Columns   []string                 `json:"columns"`
	Rows      []map[string]interface{} `json:"rows"`
	TotalRows int                      `json:"totalRows"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"pageSize"`
}

// Mock data storage
var mockConnections = []Connection{
	{
		ID:       "1",
		Name:     "Main Postgres",
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Database: "ecommerce_prod",
		Status:   "connected",
		Group:    "Production",
	},
	{
		ID:       "2",
		Name:     "Analytics Replica",
		Type:     "postgresql",
		Host:     "analytics.example.com",
		Port:     5432,
		Username: "analytics",
		Database: "analytics_db",
		Status:   "connected",
		Group:    "Production",
	},
	{
		ID:       "3",
		Name:     "Staging Cluster",
		Type:     "mysql",
		Host:     "staging.example.com",
		Port:     3306,
		Username: "dev",
		Database: "staging",
		Status:   "disconnected",
		Group:    "Development",
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
		Columns: []string{"id", "user_id", "amount", "status", "created_at"},
		Rows: []map[string]interface{}{
			{
				"id":         "TXN_8841",
				"user_id":    "USR_2291",
				"amount":     "$1,250.00",
				"status":     "COMPLETED",
				"created_at": "2023-10-24 14:22:10",
			},
			{
				"id":         "TXN_8842",
				"user_id":    "USR_4102",
				"amount":     "$745.50",
				"status":     "COMPLETED",
				"created_at": "2023-10-24 14:15:05",
			},
			{
				"id":         "TXN_8843",
				"user_id":    "USR_8812",
				"amount":     "$920.00",
				"status":     "PENDING",
				"created_at": "2023-10-24 14:12:44",
			},
			{
				"id":         "TXN_8844",
				"user_id":    "USR_3301",
				"amount":     "$2,100.00",
				"status":     "COMPLETED",
				"created_at": "2023-10-24 14:10:02",
			},
		},
		RowCount:      1000,
		ExecutionTime: 42,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return `{"columns":[],"rows":[],"rowCount":0,"error":"` + err.Error() + `"}`
	}
	return string(data)
}

// OptimizeSQL optimizes a SQL query using AI
func (a *App) OptimizeSQL(sql string) string {
	// Mock optimization
	optimized := sql
	if len(sql) > 0 {
		// Simple mock optimization
		optimized = sql + "\n-- AI Optimized: Added index hint"
	}

	result := OptimizedSQL{
		Original:  sql,
		Optimized: optimized,
		Suggestions: []string{
			"Added index hint for primary key range",
			"Filtered specific columns instead of SELECT *",
			"Added ORDER BY clause for better performance",
		},
		PerformanceGain: "~35% faster",
	}

	data, err := json.Marshal(result)
	if err != nil {
		return `{"original":"","optimized":"","suggestions":[]}`
	}
	return string(data)
}

// GetTables returns all tables for a connection
func (a *App) GetTables(connectionID string) string {
	tables := []Table{
		{Name: "users", Type: "table", RowCount: 1250},
		{Name: "orders", Type: "table", RowCount: 3420},
		{Name: "transactions", Type: "table", RowCount: 8900},
		{Name: "products", Type: "table", RowCount: 560},
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
			"id":         "TXN_8841",
			"user_id":    "USR_2291",
			"amount":     "$1,250.00",
			"status":     "COMPLETED",
			"created_at": "2023-10-24 14:22:10",
		},
		{
			"id":         "TXN_8842",
			"user_id":    "USR_4102",
			"amount":     "$745.50",
			"status":     "COMPLETED",
			"created_at": "2023-10-24 14:15:05",
		},
		{
			"id":         "TXN_8843",
			"user_id":    "USR_8812",
			"amount":     "$920.00",
			"status":     "PENDING",
			"created_at": "2023-10-24 14:12:44",
		},
		{
			"id":         "TXN_8844",
			"user_id":    "USR_3301",
			"amount":     "$2,100.00",
			"status":     "COMPLETED",
			"created_at": "2023-10-24 14:10:02",
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
		Columns:   []string{"id", "user_id", "amount", "status", "created_at"},
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
				Type:         "BIGINT (AUTO_INC)",
				IsPrimaryKey: true,
				IsNotNull:    true,
				IsUnique:     true,
			},
			{
				Name:         "email",
				Type:         "VARCHAR(255)",
				IsPrimaryKey: false,
				IsNotNull:    true,
				IsUnique:     true,
			},
			{
				Name:         "name",
				Type:         "VARCHAR(255)",
				IsPrimaryKey: false,
				IsNotNull:    false,
				IsUnique:     false,
			},
			{
				Name:         "created_at",
				Type:         "TIMESTAMP",
				IsPrimaryKey: false,
				IsNotNull:    true,
				IsUnique:     false,
				DefaultValue: "CURRENT_TIMESTAMP",
			},
		},
		Indexes: []Index{
			{
				Name:     "PRIMARY_KEY",
				Columns:  []string{"id"},
				IsUnique: true,
				Type:     "PRIMARY",
			},
			{
				Name:     "idx_users_email",
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

// UpdateTableSchema updates table schema
func (a *App) UpdateTableSchema(connectionID, tableName, schemaJSON string) error {
	// Mock update - in real implementation, this would update the database schema
	return nil
}

// ImportData imports data into a table
func (a *App) ImportData(connectionID, tableName, format string, data []byte) error {
	// Mock import - in real implementation, this would import data into the database
	return nil
}

// ExportData exports data from a table
func (a *App) ExportData(connectionID, tableName, format string) string {
	// Mock export - return JSON string of exported data
	result := map[string]interface{}{
		"success":  true,
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
