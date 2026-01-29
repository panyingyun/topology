package db

import "testing"

func TestIsSelect(t *testing.T) {
	tests := []struct {
		sql    string
		expect bool
	}{
		{"SELECT 1", true},
		{"select * from t", true},
		{"  SELECT 1", true},
		{"\n\tSELECT 1", true},
		{"-- comment\nSELECT 1", true},
		{"/* comment */ SELECT 1", true},
		{"SHOW TABLES", true},
		{"DESCRIBE t", true},
		{"DESC t", true},
		{"EXPLAIN SELECT 1", true},
		{"PRAGMA table_info(t)", true},
		{"INSERT INTO t VALUES (1)", false},
		{"UPDATE t SET x = 1", false},
		{"DELETE FROM t", false},
		{"CREATE TABLE t (id int)", false},
		{"DROP TABLE t", false},
		{"", false},
		{"-- only comment", false},
	}
	for _, tt := range tests {
		got := IsSelect(tt.sql)
		if got != tt.expect {
			t.Errorf("IsSelect(%q) = %v, want %v", tt.sql, got, tt.expect)
		}
	}
}
