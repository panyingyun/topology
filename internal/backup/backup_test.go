package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunBackupUnsupported(t *testing.T) {
	ctx := context.Background()
	c := &Conn{Type: "invalid"}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.sql")
	err := RunBackup(ctx, c, path)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestRunBackupSQLiteNoDB(t *testing.T) {
	ctx := context.Background()
	c := &Conn{Type: "sqlite", Database: ""}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.sql")
	err := RunBackup(ctx, c, path)
	if err == nil {
		t.Fatal("expected error when database path empty")
	}
}

func TestRunRestoreUnsupported(t *testing.T) {
	ctx := context.Background()
	c := &Conn{Type: "invalid"}
	err := RunRestore(ctx, c, "/tmp/nonexistent.sql")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestRunRestoreFileNotFound(t *testing.T) {
	ctx := context.Background()
	c := &Conn{Type: "mysql", Host: "127.0.0.1", Port: 3306, Username: "u", Password: "p"}
	err := RunRestore(ctx, c, "/nonexistent/path/backup.sql")
	if err == nil {
		t.Fatal("expected error when file not found")
	}
}

func TestRunBackupSQLiteDump(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	_ = os.WriteFile(dbPath, []byte("not a real db"), 0o644)
	c := &Conn{Type: "sqlite", Database: dbPath}
	outPath := filepath.Join(t.TempDir(), "dump.sql")
	err := RunBackup(ctx, c, outPath)
	// sqlite3 .dump will likely fail on invalid db; we primarily check we don't panic
	if err != nil {
		t.Logf("RunBackup (invalid sqlite) err: %v", err)
		return
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("backup file missing: %v", err)
	}
}
