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

// cacheKey returns the map key for connection cache. Empty sessionID means shared connection per connID.
func cacheKey(connID, sessionID string) string {
	if sessionID == "" {
		return connID
	}
	return connID + "\x00" + sessionID
}

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

// Open opens a DB and caches it by connID and optional sessionID. Uses retry with backoff on transient failure.
// When sessionID is non-empty, the connection is isolated per tab/session.
func Open(connID, sessionID, driver, dsn string) (*gorm.DB, error) {
	key := cacheKey(connID, sessionID)
	mu.Lock()
	defer mu.Unlock()
	if cached, ok := connCache[key]; ok {
		sqlDB, _ := cached.DB()
		if sqlDB != nil && sqlDB.Ping() == nil {
			return cached, nil
		}
		delete(connCache, key)
	}

	var lastErr error
	backoff := OpenRetryDelay
	for attempt := 0; attempt < OpenRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2
		}
		db, err := openOnce(key, driver, dsn)
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

// openOnce opens a single connection and configures the pool; caller holds mu. key is the cache map key.
func openOnce(key, driver, dsn string) (*gorm.DB, error) {
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

	connCache[key] = db
	return db, nil
}

// Get returns cached DB for connID and optional sessionID, or nil if not found.
func Get(connID, sessionID string) (*gorm.DB, bool) {
	key := cacheKey(connID, sessionID)
	mu.RLock()
	defer mu.RUnlock()
	db, ok := connCache[key]
	return db, ok
}

// Close closes and removes cached DB for the given connID and sessionID.
func Close(connID, sessionID string) {
	key := cacheKey(connID, sessionID)
	mu.Lock()
	defer mu.Unlock()
	if db, ok := connCache[key]; ok {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		delete(connCache, key)
	}
}

// CloseConnection closes all cached DBs for this connection (all sessions). Used when connection is deleted or updated.
func CloseConnection(connID string) {
	mu.Lock()
	defer mu.Unlock()
	var toDelete []string
	for k := range connCache {
		if k == connID || (len(k) > len(connID) && k[len(connID)] == '\x00' && k[:len(connID)] == connID) {
			toDelete = append(toDelete, k)
		}
	}
	for _, k := range toDelete {
		if db, ok := connCache[k]; ok {
			if sqlDB, err := db.DB(); err == nil {
				_ = sqlDB.Close()
			}
			delete(connCache, k)
		}
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
