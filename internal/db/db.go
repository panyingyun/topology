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

var (
	connCache = make(map[string]*gorm.DB)
	mu        sync.RWMutex
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

// Open opens a DB and caches it by connID.
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
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
