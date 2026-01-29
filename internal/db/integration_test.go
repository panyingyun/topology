package db

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func itestPath(elem ...string) string {
	_, file, _, _ := runtime.Caller(0)
	pkgDir := filepath.Dir(file)
	root := filepath.Join(pkgDir, "..", "..")
	parts := append([]string{root, "testdb"}, elem...)
	return filepath.Join(parts...)
}

// MySQL: testdb/mysql.url + database testdb (see mysql.txt)
// SQLite: testdb/realm.db

func mysqlDSN(t *testing.T) (string, bool) {
	path := itestPath("mysql.url")
	cfg, err := LoadMySQLTestConfig(path)
	if err != nil {
		t.Skipf("MySQL config %s: %v", path, err)
		return "", false
	}
	dsn, err := BuildDSN("mysql", cfg.Host, cfg.Port, cfg.Username, cfg.Password, "testdb")
	if err != nil {
		t.Fatalf("BuildDSN mysql: %v", err)
	}
	return dsn, true
}

func sqliteDSN(t *testing.T) (string, bool) {
	path := itestPath("realm.db")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("SQLite %s not found: %v", path, err)
		return "", false
	}
	return path, true
}

func postgresDSN(t *testing.T) (string, bool) {
	path := itestPath("postgresql.url")
	cfg, err := LoadPostgreSQLTestConfig(path)
	if err != nil {
		t.Skipf("PostgreSQL config %s: %v", path, err)
		return "", false
	}
	dsn, err := BuildDSN("postgresql", cfg.Host, cfg.Port, cfg.Username, cfg.Password, "testdb")
	if err != nil {
		t.Fatalf("BuildDSN postgresql: %v", err)
	}
	return dsn, true
}

func TestIntegration_PingMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	if err := Ping("mysql", dsn); err != nil {
		t.Errorf("Ping MySQL: %v", err)
	}
}

func TestIntegration_PingSQLite(t *testing.T) {
	dsn, ok := sqliteDSN(t)
	if !ok {
		return
	}
	if err := Ping("sqlite", dsn); err != nil {
		t.Errorf("Ping SQLite: %v", err)
	}
}

func TestIntegration_OpenMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	connID := "itest-mysql-open"
	defer Close(connID, "")

	db, err := Open(connID, "", "mysql", dsn)
	if err != nil {
		t.Fatalf("Open MySQL: %v", err)
	}
	if db == nil {
		t.Fatal("Open returned nil db")
	}
	if _, ok := Get(connID, ""); !ok {
		t.Error("Get(connID) should find cached connection")
	}
}

func TestIntegration_OpenSQLite(t *testing.T) {
	dsn, ok := sqliteDSN(t)
	if !ok {
		return
	}
	connID := "itest-sqlite-open"
	defer Close(connID, "")

	db, err := Open(connID, "", "sqlite", dsn)
	if err != nil {
		t.Fatalf("Open SQLite: %v", err)
	}
	if db == nil {
		t.Fatal("Open returned nil db")
	}
}

func TestIntegration_RawSelectMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	connID := "itest-mysql-raw"
	db, err := Open(connID, "", "mysql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	cols, rows, err := RawSelect(db, "SELECT 1 AS one, 2 AS two")
	if err != nil {
		t.Fatalf("RawSelect: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
	if len(rows) > 0 {
		if v, ok := rows[0]["one"]; !ok {
			t.Error("row missing key one")
		} else if n, ok := v.(int64); !ok || n != 1 {
			t.Errorf("one: expected 1, got %v", v)
		}
	}
}

func TestIntegration_RawSelectSQLite(t *testing.T) {
	dsn, ok := sqliteDSN(t)
	if !ok {
		return
	}
	connID := "itest-sqlite-raw"
	db, err := Open(connID, "", "sqlite", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	cols, rows, err := RawSelect(db, "SELECT 1 AS one, 2 AS two")
	if err != nil {
		t.Fatalf("RawSelect: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestIntegration_DatabaseNamesMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	connID := "itest-mysql-dbs"
	db, err := Open(connID, "", "mysql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	names, err := DatabaseNames(db, "mysql")
	if err != nil {
		t.Fatalf("DatabaseNames: %v", err)
	}
	if len(names) == 0 {
		t.Error("expected at least one database")
	}
	found := false
	for _, n := range names {
		if n == "testdb" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected testdb in %v", names)
	}
}

func TestIntegration_TableNamesMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	connID := "itest-mysql-tables"
	db, err := Open(connID, "", "mysql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, err = RawExec(db, "CREATE TABLE IF NOT EXISTS _topology_itest (id INT PRIMARY KEY, x INT)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	names, err := TableNames(db, "mysql", "testdb")
	if err != nil {
		t.Fatalf("TableNames: %v", err)
	}
	found := false
	for _, n := range names {
		if n == "_topology_itest" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected _topology_itest in %v", names)
	}
}

func TestIntegration_TableNamesSQLite(t *testing.T) {
	dsn, ok := sqliteDSN(t)
	if !ok {
		return
	}
	connID := "itest-sqlite-tables"
	db, err := Open(connID, "", "sqlite", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, err = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id INTEGER PRIMARY KEY, x INTEGER)`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	names, err := TableNames(db, "sqlite", "")
	if err != nil {
		t.Fatalf("TableNames: %v", err)
	}
	found := false
	for _, n := range names {
		if n == "_topology_itest" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected _topology_itest in %v", names)
	}
}

func TestIntegration_TableSchemaMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	connID := "itest-mysql-schema"
	db, err := Open(connID, "", "mysql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, "CREATE TABLE IF NOT EXISTS _topology_itest (id INT PRIMARY KEY, x INT)")
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	info, err := TableSchema(db, "mysql", "testdb", "_topology_itest")
	if err != nil {
		t.Fatalf("TableSchema: %v", err)
	}
	if info.Name != "_topology_itest" {
		t.Errorf("Name: expected _topology_itest, got %q", info.Name)
	}
	if len(info.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(info.Columns))
	}
}

func TestIntegration_TableSchemaSQLite(t *testing.T) {
	dsn, ok := sqliteDSN(t)
	if !ok {
		return
	}
	connID := "itest-sqlite-schema"
	db, err := Open(connID, "", "sqlite", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id INTEGER PRIMARY KEY, x INTEGER)`)
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	info, err := TableSchema(db, "sqlite", "", "_topology_itest")
	if err != nil {
		t.Fatalf("TableSchema: %v", err)
	}
	if info.Name != "_topology_itest" {
		t.Errorf("Name: expected _topology_itest, got %q", info.Name)
	}
	if len(info.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(info.Columns))
	}
}

func TestIntegration_TableDataMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	connID := "itest-mysql-data"
	db, err := Open(connID, "", "mysql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, "CREATE TABLE IF NOT EXISTS _topology_itest (id INT PRIMARY KEY, x INT)")
	_, _ = RawExec(db, "INSERT INTO _topology_itest (id, x) VALUES (1, 10), (2, 20)")
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	total, err := TableRowCount(db, "mysql", "testdb", "_topology_itest")
	if err != nil {
		t.Fatalf("TableRowCount: %v", err)
	}
	if total != 2 {
		t.Errorf("TableRowCount: expected 2, got %d", total)
	}

	cols, rows, total2, err := TableData(db, "mysql", "testdb", "_topology_itest", 10, 0)
	if err != nil {
		t.Fatalf("TableData: %v", err)
	}
	if total2 != 2 {
		t.Errorf("TableData total: expected 2, got %d", total2)
	}
	if len(cols) != 2 {
		t.Errorf("TableData cols: expected 2, got %d", len(cols))
	}
	if len(rows) != 2 {
		t.Errorf("TableData rows: expected 2, got %d", len(rows))
	}
}

func TestIntegration_TableDataSQLite(t *testing.T) {
	dsn, ok := sqliteDSN(t)
	if !ok {
		return
	}
	connID := "itest-sqlite-data"
	db, err := Open(connID, "", "sqlite", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id INTEGER PRIMARY KEY, x INTEGER)`)
	_, _ = RawExec(db, "INSERT INTO _topology_itest (id, x) VALUES (1, 10), (2, 20)")
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	total, err := TableRowCount(db, "sqlite", "", "_topology_itest")
	if err != nil {
		t.Fatalf("TableRowCount: %v", err)
	}
	if total != 2 {
		t.Errorf("TableRowCount: expected 2, got %d", total)
	}

	cols, rows, total2, err := TableData(db, "sqlite", "", "_topology_itest", 10, 0)
	if err != nil {
		t.Fatalf("TableData: %v", err)
	}
	if total2 != 2 {
		t.Errorf("TableData total: expected 2, got %d", total2)
	}
	if len(cols) != 2 {
		t.Errorf("TableData cols: expected 2, got %d", len(cols))
	}
	if len(rows) != 2 {
		t.Errorf("TableData rows: expected 2, got %d", len(rows))
	}
}

func TestIntegration_RawExecMySQL(t *testing.T) {
	dsn, ok := mysqlDSN(t)
	if !ok {
		return
	}
	connID := "itest-mysql-exec"
	db, err := Open(connID, "", "mysql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, "CREATE TABLE IF NOT EXISTS _topology_itest (id INT PRIMARY KEY)")
	n, err := RawExec(db, "INSERT INTO _topology_itest (id) VALUES (1)")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if n != 1 {
		t.Errorf("INSERT RowsAffected: expected 1, got %d", n)
	}
	n, err = RawExec(db, "UPDATE _topology_itest SET id = 2 WHERE id = 1")
	if err != nil {
		t.Fatalf("UPDATE: %v", err)
	}
	if n != 1 {
		t.Errorf("UPDATE RowsAffected: expected 1, got %d", n)
	}
	n, err = RawExec(db, "DELETE FROM _topology_itest WHERE id = 2")
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	if n != 1 {
		t.Errorf("DELETE RowsAffected: expected 1, got %d", n)
	}
	_, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest")
}

func TestIntegration_RawExecSQLite(t *testing.T) {
	dsn, ok := sqliteDSN(t)
	if !ok {
		return
	}
	connID := "itest-sqlite-exec"
	db, err := Open(connID, "", "sqlite", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id INTEGER PRIMARY KEY)`)
	n, err := RawExec(db, "INSERT INTO _topology_itest (id) VALUES (1)")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if n != 1 {
		t.Errorf("INSERT RowsAffected: expected 1, got %d", n)
	}
	_, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest")
}

func TestIntegration_CloseAll(t *testing.T) {
	CloseAll()
}

// TestIntegration_PostgreSQLPlaceholder: db 层尚未支持 PostgreSQL（见 testdb/postgresql.txt）。
// 支持后可将此用例改为实际 Ping/Open 等测试并去掉 t.Skip。
func TestIntegration_PingPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	if err := Ping("postgresql", dsn); err != nil {
		t.Errorf("Ping PostgreSQL: %v", err)
	}
}

func TestIntegration_OpenPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-open"
	defer Close(connID, "")

	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open PostgreSQL: %v", err)
	}
	if db == nil {
		t.Fatal("Open returned nil db")
	}
}

func TestIntegration_RawSelectPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-raw"
	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	cols, rows, err := RawSelect(db, "SELECT 1 AS one, 2 AS two")
	if err != nil {
		t.Fatalf("RawSelect: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("expected 2 columns, got %d", len(cols))
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestIntegration_DatabaseNamesPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-dbs"
	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	names, err := DatabaseNames(db, "postgresql")
	if err != nil {
		t.Fatalf("DatabaseNames: %v", err)
	}
	if len(names) == 0 {
		t.Error("expected at least one database")
	}
	found := false
	for _, n := range names {
		if n == "testdb" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected testdb in %v", names)
	}
}

func TestIntegration_TableNamesPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-tables"
	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id SERIAL PRIMARY KEY, x INT)`)
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	names, err := TableNames(db, "postgresql", "public")
	if err != nil {
		t.Fatalf("TableNames: %v", err)
	}
	found := false
	for _, n := range names {
		if n == "_topology_itest" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected _topology_itest in %v", names)
	}
}

func TestIntegration_TableSchemaPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-schema"
	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id SERIAL PRIMARY KEY, x INT)`)
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	info, err := TableSchema(db, "postgresql", "public", "_topology_itest")
	if err != nil {
		t.Fatalf("TableSchema: %v", err)
	}
	if info.Name != "_topology_itest" {
		t.Errorf("Name: expected _topology_itest, got %q", info.Name)
	}
	if len(info.Columns) < 2 {
		t.Errorf("expected at least 2 columns, got %d", len(info.Columns))
	}
}

func TestIntegration_TableDataPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-data"
	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id SERIAL PRIMARY KEY, x INT)`)
	_, _ = RawExec(db, "INSERT INTO _topology_itest (x) VALUES (10), (20)")
	defer func() { _, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest") }()

	total, err := TableRowCount(db, "postgresql", "public", "_topology_itest")
	if err != nil {
		t.Fatalf("TableRowCount: %v", err)
	}
	if total != 2 {
		t.Errorf("TableRowCount: expected 2, got %d", total)
	}

	cols, rows, total2, err := TableData(db, "postgresql", "public", "_topology_itest", 10, 0)
	if err != nil {
		t.Fatalf("TableData: %v", err)
	}
	if total2 != 2 {
		t.Errorf("TableData total: expected 2, got %d", total2)
	}
	if len(cols) < 2 {
		t.Errorf("TableData cols: expected at least 2, got %d", len(cols))
	}
	if len(rows) != 2 {
		t.Errorf("TableData rows: expected 2, got %d", len(rows))
	}
}

func TestIntegration_RawExecPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-exec"
	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	_, _ = RawExec(db, `CREATE TABLE IF NOT EXISTS _topology_itest (id SERIAL PRIMARY KEY)`)
	n, err := RawExec(db, "INSERT INTO _topology_itest DEFAULT VALUES")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	if n != 1 {
		t.Errorf("INSERT RowsAffected: expected 1, got %d", n)
	}
	_, _ = RawExec(db, "DROP TABLE IF EXISTS _topology_itest")
}

func TestIntegration_ExplainPostgreSQL(t *testing.T) {
	dsn, ok := postgresDSN(t)
	if !ok {
		return
	}
	connID := "itest-pg-explain"
	db, err := Open(connID, "", "postgresql", dsn)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer Close(connID, "")

	cols, rows, err := RawSelect(db, "EXPLAIN (FORMAT JSON) SELECT 1")
	if err != nil {
		t.Fatalf("EXPLAIN: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
	if len(cols) == 0 {
		t.Error("expected at least one column")
	}
	firstCol := cols[0]
	v, ok := rows[0][firstCol]
	if !ok || v == nil {
		t.Error("expected non-nil EXPLAIN JSON value")
	}
	var js string
	switch x := v.(type) {
	case string:
		js = x
	case []byte:
		js = string(x)
	}
	if js == "" || len(js) < 10 {
		t.Errorf("EXPLAIN JSON too short: %q", js)
	}
	if !strings.Contains(js, "Plan") {
		t.Errorf("EXPLAIN JSON missing Plan")
	}
}
