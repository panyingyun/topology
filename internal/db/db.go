package db

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// PoolConfig holds connection pool settings (defaults used when opening).
var (
	connCache = make(map[string]*gorm.DB)
	mu        sync.RWMutex

	// Default pool settings: balanced for desktop app with multiple connections.
	MaxIdleConns    = 5
	MaxOpenConns    = 20
	ConnMaxLifetime = 30 * time.Minute // close connections older than 30m
	ConnMaxIdleTime = 5 * time.Minute  // close idle connections after 5m (helps with server-side idle timeout)
	OpenRetries     = 4                // total attempts (1 initial + 3 retries)
	OpenRetryDelay  = time.Second      // backoff base: 1s, 2s, 4s
)

// BuildDSN builds DSN for mysql or sqlite. For sqlite, host is unused; database is the file path.
func BuildDSN(driver, host string, port int, user, pass, database string) (string, error) {
	switch driver {
	case "mysql":
		db := database
		if db == "" {
			db = "mysql"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, pass, host, port, db), nil
	case "sqlite":
		path := database
		if path == "" {
			path = filepath.Join("testdb", "realm.db")
		}
		if !strings.HasPrefix(path, "file:") && !strings.HasSuffix(path, ".db") {
			if !strings.Contains(path, ".") {
				path = path + ".db"
			}
		}
		return path, nil
	default:
		return "", fmt.Errorf("unsupported driver: %s", driver)
	}
}

// Open opens a DB and caches it by connID. Uses retry with backoff on transient failure.
func Open(connID, driver, dsn string) (*gorm.DB, error) {
	mu.Lock()
	defer mu.Unlock()
	if cached, ok := connCache[connID]; ok {
		sqlDB, _ := cached.DB()
		if sqlDB != nil && sqlDB.Ping() == nil {
			return cached, nil
		}
		delete(connCache, connID)
	}

	var lastErr error
	backoff := OpenRetryDelay
	for attempt := 0; attempt < OpenRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2
		}
		db, err := openOnce(connID, driver, dsn)
		if err == nil {
			return db, nil
		}
		lastErr = err
		// SQLite file errors usually don't benefit from retry
		if driver == "sqlite" {
			return nil, err
		}
	}
	return nil, lastErr
}

// openOnce opens a single connection and configures the pool; caller holds mu.
func openOnce(connID, driver, dsn string) (*gorm.DB, error) {
	var dial gorm.Dialector
	switch driver {
	case "mysql":
		dial = mysql.Open(dsn)
	case "sqlite":
		dial = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(MaxIdleConns)
	sqlDB.SetMaxOpenConns(MaxOpenConns)
	sqlDB.SetConnMaxLifetime(ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(ConnMaxIdleTime)

	connCache[connID] = db
	return db, nil
}

// Get returns cached DB for connID, or nil if not found.
func Get(connID string) (*gorm.DB, bool) {
	mu.RLock()
	defer mu.RUnlock()
	db, ok := connCache[connID]
	return db, ok
}

// Close closes and removes cached DB for connID.
func Close(connID string) {
	mu.Lock()
	defer mu.Unlock()
	if db, ok := connCache[connID]; ok {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		delete(connCache, connID)
	}
}

// CloseAll closes all cached connections.
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()
	for id, db := range connCache {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		delete(connCache, id)
	}
}

// Ping opens a temporary DB with the given DSN, pings, then closes. Used for TestConnection.
func Ping(driver, dsn string) error {
	db, err := openTemp(driver, dsn)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return sqlDB.Ping()
}

func openTemp(driver, dsn string) (*gorm.DB, error) {
	var dial gorm.Dialector
	switch driver {
	case "mysql":
		dial = mysql.Open(dsn)
	case "sqlite":
		dial = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	return gorm.Open(dial, &gorm.Config{})
}
