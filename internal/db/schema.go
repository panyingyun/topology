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

// SchemaForeignKey holds FK metadata for a table.
type SchemaForeignKey struct {
	Name              string   `json:"name"`
	Columns           []string `json:"columns"`
	ReferencedTable   string   `json:"referencedTable"`
	ReferencedColumns []string `json:"referencedColumns"`
	OnDelete          string   `json:"onDelete,omitempty"`
	OnUpdate          string   `json:"onUpdate,omitempty"`
}

// TableSchemaInfo holds schema info for a table.
type TableSchemaInfo struct {
	Name        string             `json:"name"`
	Columns     []SchemaColumn     `json:"columns"`
	ForeignKeys []SchemaForeignKey `json:"foreignKeys"`
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
	fks, _ := mysqlTableForeignKeys(db, database, table)
	info.ForeignKeys = fks
	return info, nil
}

func mysqlTableForeignKeys(db *gorm.DB, database, table string) ([]SchemaForeignKey, error) {
	q := `SELECT kcu.CONSTRAINT_NAME, kcu.COLUMN_NAME, kcu.REFERENCED_TABLE_NAME, kcu.REFERENCED_COLUMN_NAME,
		rc.DELETE_RULE, rc.UPDATE_RULE
		FROM information_schema.KEY_COLUMN_USAGE kcu
		JOIN information_schema.REFERENTIAL_CONSTRAINTS rc
		  ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME AND kcu.CONSTRAINT_SCHEMA = rc.CONSTRAINT_SCHEMA
		WHERE kcu.TABLE_SCHEMA = ? AND kcu.TABLE_NAME = ? AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION`
	var raw []struct {
		ConstraintName   string `gorm:"column:CONSTRAINT_NAME"`
		ColumnName       string `gorm:"column:COLUMN_NAME"`
		ReferencedTable  string `gorm:"column:REFERENCED_TABLE_NAME"`
		ReferencedColumn string `gorm:"column:REFERENCED_COLUMN_NAME"`
		DeleteRule       string `gorm:"column:DELETE_RULE"`
		UpdateRule       string `gorm:"column:UPDATE_RULE"`
	}
	if database == "" {
		q = `SELECT kcu.CONSTRAINT_NAME, kcu.COLUMN_NAME, kcu.REFERENCED_TABLE_NAME, kcu.REFERENCED_COLUMN_NAME,
			rc.DELETE_RULE, rc.UPDATE_RULE
			FROM information_schema.KEY_COLUMN_USAGE kcu
			JOIN information_schema.REFERENTIAL_CONSTRAINTS rc
			  ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME AND kcu.CONSTRAINT_SCHEMA = rc.CONSTRAINT_SCHEMA
			WHERE kcu.TABLE_SCHEMA = DATABASE() AND kcu.TABLE_NAME = ? AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
			ORDER BY kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION`
		if err := db.Raw(q, table).Scan(&raw).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Raw(q, database, table).Scan(&raw).Error; err != nil {
			return nil, err
		}
	}
	byName := make(map[string]*SchemaForeignKey)
	for _, r := range raw {
		fk, ok := byName[r.ConstraintName]
		if !ok {
			fk = &SchemaForeignKey{
				Name:              r.ConstraintName,
				Columns:           nil,
				ReferencedTable:   r.ReferencedTable,
				ReferencedColumns: nil,
				OnDelete:          r.DeleteRule,
				OnUpdate:          r.UpdateRule,
			}
			byName[r.ConstraintName] = fk
		}
		fk.Columns = append(fk.Columns, r.ColumnName)
		fk.ReferencedColumns = append(fk.ReferencedColumns, r.ReferencedColumn)
	}
	out := make([]SchemaForeignKey, 0, len(byName))
	for _, fk := range byName {
		out = append(out, *fk)
	}
	return out, nil
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
	fks, _ := postgresTableForeignKeys(db, schema, table)
	info.ForeignKeys = fks
	return info, nil
}

func postgresTableForeignKeys(db *gorm.DB, schema, table string) ([]SchemaForeignKey, error) {
	q := `SELECT kcu.constraint_name, kcu.column_name, kcu.ordinal_position,
		rel_tco.table_name AS ref_table,
		rc.delete_rule, rc.update_rule
		FROM information_schema.table_constraints tco
		JOIN information_schema.key_column_usage kcu
		  ON tco.constraint_schema = kcu.constraint_schema AND tco.constraint_name = kcu.constraint_name
		JOIN information_schema.referential_constraints rc
		  ON tco.constraint_schema = rc.constraint_schema AND tco.constraint_name = rc.constraint_name
		JOIN information_schema.table_constraints rel_tco
		  ON rc.unique_constraint_schema = rel_tco.constraint_schema
		  AND rc.unique_constraint_name = rel_tco.constraint_name
		WHERE tco.constraint_type = 'FOREIGN KEY' AND tco.table_schema = ? AND tco.table_name = ?
		ORDER BY kcu.constraint_name, kcu.ordinal_position`
	var raw []struct {
		ConstraintName string `gorm:"column:constraint_name"`
		ColumnName     string `gorm:"column:column_name"`
		OrdinalPos     int    `gorm:"column:ordinal_position"`
		RefTable       string `gorm:"column:ref_table"`
		DeleteRule     string `gorm:"column:delete_rule"`
		UpdateRule     string `gorm:"column:update_rule"`
	}
	if err := db.Raw(q, schema, table).Scan(&raw).Error; err != nil {
		return nil, err
	}
	byName := make(map[string]*SchemaForeignKey)
	for _, r := range raw {
		fk, ok := byName[r.ConstraintName]
		if !ok {
			fk = &SchemaForeignKey{
				Name:              r.ConstraintName,
				Columns:           nil,
				ReferencedTable:   r.RefTable,
				ReferencedColumns: nil,
				OnDelete:          r.DeleteRule,
				OnUpdate:          r.UpdateRule,
			}
			byName[r.ConstraintName] = fk
		}
		fk.Columns = append(fk.Columns, r.ColumnName)
		// ReferencedColumns: not available from this query; leave empty
	}
	out := make([]SchemaForeignKey, 0, len(byName))
	for _, fk := range byName {
		out = append(out, *fk)
	}
	return out, nil
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
	fks, _ := sqliteTableForeignKeys(db, table)
	info.ForeignKeys = fks
	return info, nil
}

func sqliteTableForeignKeys(db *gorm.DB, table string) ([]SchemaForeignKey, error) {
	q := "PRAGMA foreign_key_list(" + quoteIdent("sqlite", table) + ")"
	var raw []struct {
		ID       int     `gorm:"column:id"`
		Seq      int     `gorm:"column:seq"`
		Table    string  `gorm:"column:table"`
		From     string  `gorm:"column:from"`
		To       string  `gorm:"column:to"`
		OnUpdate *string `gorm:"column:on_update"`
		OnDelete *string `gorm:"column:on_delete"`
	}
	if err := db.Raw(q).Scan(&raw).Error; err != nil {
		return nil, err
	}
	byID := make(map[int]*SchemaForeignKey)
	for _, r := range raw {
		fk, ok := byID[r.ID]
		if !ok {
			del, upd := "", ""
			if r.OnDelete != nil {
				del = *r.OnDelete
			}
			if r.OnUpdate != nil {
				upd = *r.OnUpdate
			}
			fk = &SchemaForeignKey{
				Name:              fmt.Sprintf("fk_%s_%d", table, r.ID),
				Columns:           nil,
				ReferencedTable:   r.Table,
				ReferencedColumns: nil,
				OnDelete:          del,
				OnUpdate:          upd,
			}
			byID[r.ID] = fk
		}
		fk.Columns = append(fk.Columns, r.From)
		fk.ReferencedColumns = append(fk.ReferencedColumns, r.To)
	}
	out := make([]SchemaForeignKey, 0, len(byID))
	for _, fk := range byID {
		out = append(out, *fk)
	}
	return out, nil
}
