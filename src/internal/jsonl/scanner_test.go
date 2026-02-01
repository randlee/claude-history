package jsonl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestScanner_Scan(t *testing.T) {
	// Create a temporary JSONL file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id": 1, "name": "first"}
{"id": 2, "name": "second"}
{"id": 3, "name": "third"}
`
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	var lines []json.RawMessage

	err := s.Scan(testFile, func(line json.RawMessage) error {
		lines = append(lines, line)
		return nil
	})

	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
}

func TestScanner_SkipMalformed(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id": 1}
not valid json
{"id": 2}
also not valid
{"id": 3}
`
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	s := NewScanner()
	var count int

	err := s.Scan(testFile, func(line json.RawMessage) error {
		count++
		return nil
	})

	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 valid lines, got %d", count)
	}
}

func TestScanInto(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id": 1, "name": "first"}
{"id": 2, "name": "second"}
`
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	type Entry struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	var entries []Entry
	err := ScanInto(testFile, func(entry Entry) error {
		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		t.Fatalf("ScanInto failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Name != "first" {
		t.Errorf("Expected first entry name 'first', got %q", entries[0].Name)
	}
}

func TestReadAll(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"value": 10}
{"value": 20}
{"value": 30}
`
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	type Entry struct {
		Value int `json:"value"`
	}

	entries, err := ReadAll[Entry](testFile)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	sum := 0
	for _, e := range entries {
		sum += e.Value
	}
	if sum != 60 {
		t.Errorf("Expected sum 60, got %d", sum)
	}
}

func TestCountLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"a": 1}
{"b": 2}
invalid
{"c": 3}
{"d": 4}
`
	if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	count, err := CountLines(testFile)
	if err != nil {
		t.Fatalf("CountLines failed: %v", err)
	}

	if count != 4 {
		t.Errorf("Expected 4 valid lines, got %d", count)
	}
}
