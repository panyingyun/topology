package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Conn holds connection params for backup/restore.
type Conn struct {
	Type     string // mysql, postgresql, sqlite
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// RunBackup runs mysqldump (MySQL), pg_dump (PostgreSQL), or sqlite3 .dump (SQLite). outputPath must be absolute. SSH not supported.
func RunBackup(ctx context.Context, c *Conn, outputPath string) error {
	switch c.Type {
	case "mysql":
		return runMySQLBackup(ctx, c, outputPath)
	case "postgresql", "postgres":
		return runPGBackup(ctx, c, outputPath)
	case "sqlite":
		return runSQLiteBackup(ctx, c, outputPath)
	default:
		return fmt.Errorf("unsupported backup type: %s", c.Type)
	}
}

func runMySQLBackup(ctx context.Context, c *Conn, out string) error {
	args := []string{"-h", c.Host, "-P", fmt.Sprintf("%d", c.Port), "-u", c.Username}
	if c.Password != "" {
		args = append(args, "-p"+c.Password)
	}
	if c.Database != "" {
		args = append(args, "--databases", c.Database)
	} else {
		args = append(args, "--all-databases")
	}
	args = append(args, "--single-transaction", "--routines", "--triggers", "--events")

	cmd := exec.CommandContext(ctx, "mysqldump", args...)
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create backup file: %w", err)
	}
	defer f.Close()
	cmd.Stdout = f
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		_ = os.Remove(out)
		return fmt.Errorf("mysqldump: %w", err)
	}
	return nil
}

func runPGBackup(ctx context.Context, c *Conn, out string) error {
	db := c.Database
	if db == "" {
		db = "postgres"
	}
	args := []string{"-h", c.Host, "-p", fmt.Sprintf("%d", c.Port), "-U", c.Username, "-d", db, "-f", out}
	cmd := exec.CommandContext(ctx, "pg_dump", args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+c.Password)
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		_ = os.Remove(out)
		return fmt.Errorf("pg_dump: %w", err)
	}
	return nil
}

func runSQLiteBackup(ctx context.Context, c *Conn, out string) error {
	dbPath := c.Database
	if dbPath == "" {
		return fmt.Errorf("sqlite backup requires database path")
	}
	if !filepath.IsAbs(dbPath) && !strings.HasPrefix(dbPath, "file:") {
		// treat as relative to cwd
	}
	cmd := exec.CommandContext(ctx, "sqlite3", dbPath, ".dump")
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create backup file: %w", err)
	}
	defer f.Close()
	cmd.Stdout = f
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		_ = os.Remove(out)
		return fmt.Errorf("sqlite3 dump: %w", err)
	}
	return nil
}

// RunRestore runs mysql (MySQL), psql (PostgreSQL), or sqlite3 (SQLite) to restore from backupPath. SSH not supported.
func RunRestore(ctx context.Context, c *Conn, backupPath string) error {
	switch c.Type {
	case "mysql":
		return runMySQLRestore(ctx, c, backupPath)
	case "postgresql", "postgres":
		return runPGRestore(ctx, c, backupPath)
	case "sqlite":
		return runSQLiteRestore(ctx, c, backupPath)
	default:
		return fmt.Errorf("unsupported restore type: %s", c.Type)
	}
}

func runMySQLRestore(ctx context.Context, c *Conn, fpath string) error {
	args := []string{"-h", c.Host, "-P", fmt.Sprintf("%d", c.Port), "-u", c.Username}
	if c.Password != "" {
		args = append(args, "-p"+c.Password)
	}
	// Do not pass default database; dump typically contains CREATE DATABASE / USE
	in, err := os.Open(fpath)
	if err != nil {
		return fmt.Errorf("open backup file: %w", err)
	}
	defer in.Close()
	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Stdin = in
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore: %w", err)
	}
	return nil
}

func runPGRestore(ctx context.Context, c *Conn, fpath string) error {
	db := c.Database
	if db == "" {
		db = "postgres"
	}
	args := []string{"-h", c.Host, "-p", fmt.Sprintf("%d", c.Port), "-U", c.Username, "-d", db, "-f", fpath}
	cmd := exec.CommandContext(ctx, "psql", args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+c.Password)
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("psql restore: %w", err)
	}
	return nil
}

func runSQLiteRestore(ctx context.Context, c *Conn, fpath string) error {
	dbPath := c.Database
	if dbPath == "" {
		return fmt.Errorf("sqlite restore requires database path")
	}
	in, err := os.Open(fpath)
	if err != nil {
		return fmt.Errorf("open backup file: %w", err)
	}
	defer in.Close()
	cmd := exec.CommandContext(ctx, "sqlite3", dbPath)
	cmd.Stdin = in
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sqlite3 restore: %w", err)
	}
	return nil
}
