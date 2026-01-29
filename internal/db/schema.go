package db

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// SchemaColumn represents a column in table schema.
type SchemaColumn struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"defaultValue,omitempty"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
	IsUnique     bool   `json:"isUnique"`
}

// TableSchemaInfo holds schema info for a table.
type TableSchemaInfo struct {
	Name    string         `json:"name"`
	Columns []SchemaColumn `json:"columns"`
}

// TableSchema returns schema (columns) for the given table. database is optional (MySQL: TABLE_SCHEMA; PostgreSQL: schema, default "public").
func TableSchema(db *gorm.DB, driver, database, table string) (*TableSchemaInfo, error) {
	info := &TableSchemaInfo{Name: table}
	switch driver {
	case "mysql":
		return mysqlTableSchema(db, database, table, info)
	case "postgresql", "postgres":
		return postgresTableSchema(db, database, table, info)
	case "sqlite":
		return sqliteTableSchema(db, table, info)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

func mysqlTableSchema(db *gorm.DB, database, table string, info *TableSchemaInfo) (*TableSchemaInfo, error) {
	q := "SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_KEY, EXTRA FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
	var raw []struct {
		COLUMN_NAME    string
		COLUMN_TYPE    string
		IS_NULLABLE    string
		COLUMN_DEFAULT *string
		COLUMN_KEY     string
		EXTRA          string
	}
	if database != "" {
		if err := db.Raw(q, database, table).Scan(&raw).Error; err != nil {
			return nil, err
		}
	} else {
		// use DATABASE() for current connection DB
		q2 := "SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_KEY, EXTRA FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		if err := db.Raw(q2, table).Scan(&raw).Error; err != nil {
			return nil, err
		}
	}
	for _, r := range raw {
		def := ""
		if r.COLUMN_DEFAULT != nil {
			def = *r.COLUMN_DEFAULT
		}
		info.Columns = append(info.Columns, SchemaColumn{
			Name:         r.COLUMN_NAME,
			Type:         r.COLUMN_TYPE,
			Nullable:     strings.ToUpper(r.IS_NULLABLE) == "YES",
			DefaultValue: def,
			IsPrimaryKey: strings.ToUpper(r.COLUMN_KEY) == "PRI",
			IsUnique:     strings.ToUpper(r.COLUMN_KEY) == "UNI",
		})
	}
	return info, nil
}

func postgresTableSchema(db *gorm.DB, database, table string, info *TableSchemaInfo) (*TableSchemaInfo, error) {
	schema := "public"
	if database != "" {
		schema = database
	}
	q := `SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position`
	var raw []struct {
		ColumnName    string
		DataType      string
		IsNullable    string
		ColumnDefault *string
	}
	if err := db.Raw(q, schema, table).Scan(&raw).Error; err != nil {
		return nil, err
	}
	// primary key: check pg_constraint
	pkCols := make(map[string]bool)
	var pkCheck []struct {
		Attname string
	}
	_ = db.Raw(`SELECT a.attname FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) AND a.attisdropped = false
		JOIN pg_class c ON c.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = ? AND c.relname = ? AND i.indisprimary`,
		schema, table).Scan(&pkCheck)
	for _, r := range pkCheck {
		pkCols[r.Attname] = true
	}
	for _, r := range raw {
		def := ""
		if r.ColumnDefault != nil {
			def = *r.ColumnDefault
		}
		info.Columns = append(info.Columns, SchemaColumn{
			Name:         r.ColumnName,
			Type:         r.DataType,
			Nullable:     strings.ToUpper(r.IsNullable) == "YES",
			DefaultValue: def,
			IsPrimaryKey: pkCols[r.ColumnName],
			IsUnique:     false,
		})
	}
	return info, nil
}

func sqliteTableSchema(db *gorm.DB, table string, info *TableSchemaInfo) (*TableSchemaInfo, error) {
	q := "PRAGMA table_info(" + quoteIdent("sqlite", table) + ")"
	var raw []struct {
		CID     int
		Name    string
		Type    string
		Notnull int
		Dflt    *string
		PK      int
	}
	if err := db.Raw(q).Scan(&raw).Error; err != nil {
		return nil, err
	}
	for _, r := range raw {
		def := ""
		if r.Dflt != nil {
			def = *r.Dflt
		}
		info.Columns = append(info.Columns, SchemaColumn{
			Name:         r.Name,
			Type:         r.Type,
			Nullable:     r.Notnull == 0,
			DefaultValue: def,
			IsPrimaryKey: r.PK > 0,
			IsUnique:     false, // would need PRAGMA index_list
		})
	}
	return info, nil
}
