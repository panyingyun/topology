package db

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MySQLTestConfig holds MySQL config parsed from testdb/mysql.url.
type MySQLTestConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// LoadMySQLTestConfig reads testdb/mysql.url and returns MySQL config.
// Format: 访问地址, 访问端口, 用户名, 密码 (key: value or key：value).
func LoadMySQLTestConfig(path string) (*MySQLTestConfig, error) {
	if path == "" {
		path = "testdb/mysql.url"
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := &MySQLTestConfig{Port: 3306}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		sep := ":"
		if idx := strings.Index(line, "："); idx >= 0 {
			sep = "："
		} else if idx := strings.Index(line, ":"); idx >= 0 {
			sep = ":"
		}
		parts := strings.SplitN(line, sep, 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch {
		case strings.Contains(k, "地址") || strings.Contains(k, "host"):
			c.Host = v
		case strings.Contains(k, "端口") || strings.Contains(k, "port"):
			if p, e := strconv.Atoi(v); e == nil {
				c.Port = p
			}
		case strings.Contains(k, "用户名") || strings.Contains(k, "user"):
			c.Username = v
		case strings.Contains(k, "密码") || strings.Contains(k, "password"):
			c.Password = v
		}
	}
	return c, sc.Err()
}

// SQLiteTestPath returns the path to test SQLite DB (testdb/realm.db).
func SQLiteTestPath() string {
	return filepath.Join("testdb", "realm.db")
}
