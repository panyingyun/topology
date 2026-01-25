package db

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// RawSelect runs a SELECT query and returns columns and rows as []map[string]interface{}.
func RawSelect(db *gorm.DB, q string) (cols []string, rows []map[string]interface{}, err error) {
	var rs *sql.Rows
	rs, err = db.Raw(q).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rs.Close()

	cols, err = rs.Columns()
	if err != nil {
		return nil, nil, err
	}
	types, _ := rs.ColumnTypes()
	scanners := make([]interface{}, len(cols))
	for i := range cols {
		var v interface{}
		scanners[i] = &v
	}

	for rs.Next() {
		if err = rs.Scan(scanners...); err != nil {
			return nil, nil, err
		}
		row := make(map[string]interface{})
		for i, c := range cols {
			val := *(scanners[i].(*interface{}))
			if val != nil && types != nil && i < len(types) {
				row[c] = formatColumnValue(val, types[i].DatabaseTypeName())
			} else if val != nil {
				row[c] = val
			} else {
				row[c] = nil
			}
		}
		rows = append(rows, row)
	}
	return cols, rows, rs.Err()
}

func formatColumnValue(val interface{}, dbType string) interface{} {
	switch v := val.(type) {
	case []byte:
		return string(v)
	default:
		return v
	}
}

// RawExec runs INSERT/UPDATE/DELETE and returns rows affected.
func RawExec(db *gorm.DB, q string) (int64, error) {
	tx := db.Exec(q)
	return tx.RowsAffected, tx.Error
}

// IsSelect returns true if the trimmed, upper-cased query looks like a SELECT.
func IsSelect(q string) bool {
	q = strings.TrimSpace(q)
	// Remove leading comments and blanks
	for len(q) > 0 {
		if strings.HasPrefix(q, "--") {
			i := strings.Index(q, "\n")
			if i < 0 {
				return false
			}
			q = strings.TrimSpace(q[i+1:])
			continue
		}
		if strings.HasPrefix(q, "/*") {
			i := strings.Index(q, "*/")
			if i < 0 {
				return false
			}
			q = strings.TrimSpace(q[i+2:])
			continue
		}
		break
	}
	upper := strings.ToUpper(q)
	return strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "DESCRIBE") || strings.HasPrefix(upper, "DESC") ||
		strings.HasPrefix(upper, "EXPLAIN") || strings.HasPrefix(upper, "PRAGMA")
}

// TableNames returns table names for the given driver.
func TableNames(db *gorm.DB, driver string) ([]string, error) {
	var q string
	switch driver {
	case "mysql":
		q = "SHOW TABLES"
	case "sqlite":
		q = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	cols, rows, err := RawSelect(db, q)
	if err != nil {
		return nil, err
	}
	var names []string
	col := "Tables_in_" // MySQL SHOW TABLES
	if len(cols) > 0 {
		col = cols[0]
	}
	for _, r := range rows {
		if v, ok := r[col]; ok && v != nil {
			names = append(names, fmt.Sprint(v))
		}
	}
	return names, nil
}

// TableRowCount returns total row count for a table.
func TableRowCount(db *gorm.DB, driver, table string) (int, error) {
	table = quoteIdent(driver, table)
	q := "SELECT COUNT(*) FROM " + table
	var n int64
	err := db.Raw(q).Scan(&n).Error
	return int(n), err
}

// TableData returns columns, rows (for limit/offset), and total count.
func TableData(db *gorm.DB, driver, table string, limit, offset int) (cols []string, rows []map[string]interface{}, total int, err error) {
	total, err = TableRowCount(db, driver, table)
	if err != nil {
		return nil, nil, 0, err
	}
	tbl := quoteIdent(driver, table)
	q := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", tbl, limit, offset)
	cols, rows, err = RawSelect(db, q)
	return cols, rows, total, err
}

func quoteIdent(driver, name string) string {
	switch driver {
	case "mysql":
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case "sqlite":
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	default:
		return name
	}
}

