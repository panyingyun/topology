package db

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// testdbPath returns path to testdb (project root/testdb/...). Resolves relative to package dir so it works when go test runs from tmp.
func testdbPath(elem ...string) string {
	_, file, _, _ := runtime.Caller(0)
	pkgDir := filepath.Dir(file)
	// internal/db -> project root = ../..
	root := filepath.Join(pkgDir, "..", "..")
	parts := append([]string{root, "testdb"}, elem...)
	return filepath.Join(parts...)
}

func TestLoadMySQLTestConfig(t *testing.T) {
	path := testdbPath("mysql.url")
	cfg, err := LoadMySQLTestConfig(path)
	if err != nil {
		t.Skipf("%s not found or unreadable: %v", path, err)
		return
	}
	// testdb/mysql.url: host=127.0.0.1 port=3306 user=topology password=topology
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected Host 127.0.0.1, got %q", cfg.Host)
	}
	if cfg.Port != 3306 {
		t.Errorf("expected Port 3306, got %d", cfg.Port)
	}
	if cfg.Username != "topology" {
		t.Errorf("expected Username topology, got %q", cfg.Username)
	}
	if cfg.Password != "topology" {
		t.Errorf("expected Password topology, got %q", cfg.Password)
	}
}

func TestBuildDSN(t *testing.T) {
	dsn, err := BuildDSN("mysql", "127.0.0.1", 3306, "root", "secret", "mydb")
	if err != nil {
		t.Fatal(err)
	}
	if dsn == "" {
		t.Fatal("expected non-empty DSN")
	}
	if dsn != "root:secret@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local" {
		t.Logf("DSN: %s", dsn)
	}

	dsn, err = BuildDSN("sqlite", "", 0, "", "", "testdb/realm.db")
	if err != nil {
		t.Fatal(err)
	}
	if dsn != "testdb/realm.db" {
		t.Errorf("expected sqlite path testdb/realm.db, got %q", dsn)
	}
}

func TestSQLiteTestPath(t *testing.T) {
	p := SQLiteTestPath()
	if p == "" {
		t.Fatal("expected non-empty path")
	}
	if filepath.Base(p) != "realm.db" {
		t.Errorf("expected base realm.db, got %q", filepath.Base(p))
	}
}

func TestPingSQLite(t *testing.T) {
	path := testdbPath("realm.db")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s not found: %v", path, err)
		return
	}
	dsn := path
	if err := Ping("sqlite", dsn); err != nil {
		t.Errorf("Ping sqlite %q: %v", dsn, err)
	}
}
