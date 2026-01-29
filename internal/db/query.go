package db

import (
	"database/sql"
	"encoding/json"
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
		s := string(v)
		dt := strings.ToUpper(dbType)
		if (strings.Contains(dt, "JSON") || strings.Contains(dt, "JSONB")) && len(v) > 0 {
			var x interface{}
			if json.Unmarshal(v, &x) == nil {
				b, err := json.MarshalIndent(x, "", "  ")
				if err == nil {
					return string(b)
				}
			}
		}
		return s
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

// SchemaNames returns schema names for the current PostgreSQL database (e.g. public, user schemas). Only for driver "postgresql"/"postgres".
func SchemaNames(db *gorm.DB) ([]string, error) {
	cols, rows, err := RawSelect(db, `SELECT schema_name FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog','information_schema') AND schema_name NOT LIKE 'pg_toast%'
		ORDER BY schema_name`)
	if err != nil {
		return nil, err
	}
	col := "schema_name"
	if len(cols) > 0 {
		col = cols[0]
	}
	var names []string
	for _, r := range rows {
		if v, ok := r[col]; ok && v != nil {
			names = append(names, fmt.Sprint(v))
		}
	}
	return names, nil
}

// DatabaseNames returns database names for the given driver. MySQL: SHOW DATABASES; PostgreSQL: pg_database (or use SchemaNames for tree); SQLite: ["main"].
func DatabaseNames(db *gorm.DB, driver string) ([]string, error) {
	switch driver {
	case "mysql":
		cols, rows, err := RawSelect(db, "SHOW DATABASES")
		if err != nil {
			return nil, err
		}
		col := "Database"
		if len(cols) > 0 {
			col = cols[0]
		}
		var names []string
		for _, r := range rows {
			if v, ok := r[col]; ok && v != nil {
				names = append(names, fmt.Sprint(v))
			}
		}
		return names, nil
	case "postgresql", "postgres":
		cols, rows, err := RawSelect(db, "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
		if err != nil {
			return nil, err
		}
		col := "datname"
		if len(cols) > 0 {
			col = cols[0]
		}
		var names []string
		for _, r := range rows {
			if v, ok := r[col]; ok && v != nil {
				names = append(names, fmt.Sprint(v))
			}
		}
		return names, nil
	case "sqlite":
		return []string{"main"}, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

// TableNames returns table names for the given driver and database. For SQLite, database is ignored. For PostgreSQL, database is schema (default "public").
func TableNames(db *gorm.DB, driver, database string) ([]string, error) {
	var q string
	switch driver {
	case "mysql":
		if database != "" {
			q = "SHOW TABLES FROM " + quoteIdent(driver, database)
		} else {
			q = "SHOW TABLES"
		}
	case "postgresql", "postgres":
		schema := "public"
		if database != "" {
			schema = strings.ReplaceAll(database, "'", "''")
		}
		q = "SELECT tablename FROM pg_tables WHERE schemaname = '" + schema + "' ORDER BY tablename"
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
	col := "Tables_in_"
	if database != "" && (driver == "mysql") {
		col = "Tables_in_" + database
	}
	if driver == "postgresql" || driver == "postgres" {
		col = "tablename"
	}
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

// qualTable returns qualified table for queries: MySQL "`db`.`table`"; PostgreSQL "schema"."table" (default "public"); SQLite "table".
func qualTable(driver, database, table string) string {
	tbl := quoteIdent(driver, table)
	if driver == "mysql" && database != "" {
		return quoteIdent(driver, database) + "." + tbl
	}
	if driver == "postgresql" || driver == "postgres" {
		schema := database
		if schema == "" {
			schema = "public"
		}
		return quoteIdent(driver, schema) + "." + tbl
	}
	return tbl
}

// QualTable is the exported version of qualTable for use by app layer.
func QualTable(driver, database, table string) string {
	return qualTable(driver, database, table)
}

// TableRowCount returns total row count for a table. database is optional (MySQL: qualify db.table).
func TableRowCount(db *gorm.DB, driver, database, table string) (int, error) {
	q := "SELECT COUNT(*) FROM " + qualTable(driver, database, table)
	var n int64
	err := db.Raw(q).Scan(&n).Error
	return int(n), err
}

// TableData returns columns, rows (for limit/offset), and total count. database is optional.
func TableData(db *gorm.DB, driver, database, table string, limit, offset int) (cols []string, rows []map[string]interface{}, total int, err error) {
	total, err = TableRowCount(db, driver, database, table)
	if err != nil {
		return nil, nil, 0, err
	}
	qt := qualTable(driver, database, table)
	q := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", qt, limit, offset)
	cols, rows, err = RawSelect(db, q)
	return cols, rows, total, err
}

func quoteIdent(driver, name string) string {
	switch driver {
	case "mysql":
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case "sqlite", "postgresql", "postgres":
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	default:
		return name
	}
}
