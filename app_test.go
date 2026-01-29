package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestUserFacingError(t *testing.T) {
	tests := []struct {
		err  error
		code string
		msg  string
	}{
		{nil, "", ""},
		{fmt.Errorf("connection not found: x"), "CONNECTION_NOT_FOUND", "Connection not found"},
		{fmt.Errorf("connect: connection refused"), "CONNECTION_REFUSED", "Cannot connect"},
		{fmt.Errorf("access denied for user 'x'"), "ACCESS_DENIED", "Access denied"},
		{fmt.Errorf("syntax error at or near \"x\""), "SYNTAX_ERROR", "SQL syntax error"},
		{fmt.Errorf("relation \"foo\" does not exist"), "NOT_FOUND", "does not exist"},
		{fmt.Errorf("duplicate key value"), "DUPLICATE_KEY", "Duplicate key"},
		{fmt.Errorf("context deadline exceeded"), "TIMEOUT", "timed out"},
		{fmt.Errorf("something else"), "", "something else"},
	}
	for _, tt := range tests {
		out := userFacingError(tt.err)
		if tt.err == nil {
			if out.Message != "" {
				t.Errorf("nil err: got message %q", out.Message)
			}
			continue
		}
		if tt.code != "" && out.Code != tt.code {
			t.Errorf("err %q: code got %q want %q", tt.err.Error(), out.Code, tt.code)
		}
		if tt.msg != "" && !strings.Contains(out.Message, tt.msg) {
			t.Errorf("err %q: message got %q want substring %q", tt.err.Error(), out.Message, tt.msg)
		}
	}
}

func TestParsePGExplainJSON(t *testing.T) {
	json := `[{"Plan":{"Node Type":"Seq Scan","Relation Name":"foo","Plan Rows":100,"Total Cost":10.5}}]`
	nodes, warnings, err := parsePGExplainJSON(json)
	if err != nil {
		t.Fatalf("parsePGExplainJSON: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n := nodes[0]
	if n.Type != "Scan" || n.Label != "foo" || n.Detail != "Seq Scan" {
		t.Errorf("node: type=%q label=%q detail=%q", n.Type, n.Label, n.Detail)
	}
	if !n.FullTableScan {
		t.Error("expected full table scan")
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "foo") {
		t.Errorf("warnings: %v", warnings)
	}
}

func TestParsePGExplainJSONNested(t *testing.T) {
	json := `[{"Plan":{"Node Type":"Limit","Plan Rows":1,"Plans":[{"Node Type":"Seq Scan","Relation Name":"t","Plan Rows":10}]}}]`
	nodes, _, err := parsePGExplainJSON(json)
	if err != nil {
		t.Fatalf("parsePGExplainJSON: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Type != "Limit" && nodes[1].Type != "Limit" {
		t.Error("expected Limit node")
	}
	if nodes[0].Type != "Scan" && nodes[1].Type != "Scan" {
		t.Error("expected Scan node")
	}
}

func TestParsePGExplainJSONInvalid(t *testing.T) {
	_, _, err := parsePGExplainJSON("not json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	_, _, err = parsePGExplainJSON("[]")
	if err == nil {
		t.Fatal("expected error for empty array")
	}
	_, _, err = parsePGExplainJSON("{}")
	if err == nil {
		t.Fatal("expected error for non-array")
	}
}

func TestExtractPGExplainJSON(t *testing.T) {
	row := map[string]interface{}{"QUERY PLAN": `[{"Plan":{}}]`}
	cols := []string{"QUERY PLAN"}
	s := extractPGExplainJSON(row, cols)
	if s != `[{"Plan":{}}]` {
		t.Errorf("extract: got %q", s)
	}
	row2 := map[string]interface{}{"x": []byte(`[{"Plan":{}}]`)}
	cols2 := []string{"x"}
	s2 := extractPGExplainJSON(row2, cols2)
	if s2 != `[{"Plan":{}}]` {
		t.Errorf("extract bytes: got %q", s2)
	}
}
