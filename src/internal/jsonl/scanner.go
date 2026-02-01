// Package jsonl provides streaming JSONL file parsing with support for large files.
package jsonl

import (
	"bufio"
	"encoding/json"
	"os"
)

// Scanner reads JSONL files line by line with streaming support.
type Scanner struct {
	// MaxLineSize is the maximum size of a single line in bytes.
	// Defaults to 10MB if not set.
	MaxLineSize int
}

// NewScanner creates a new JSONL scanner with default settings.
func NewScanner() *Scanner {
	return &Scanner{
		MaxLineSize: 10 * 1024 * 1024, // 10MB default
	}
}

// Scan reads a JSONL file and calls fn for each successfully parsed line.
// Lines that fail to parse as JSON are silently skipped.
// If fn returns an error, scanning stops and that error is returned.
func (s *Scanner) Scan(filePath string, fn func(line json.RawMessage) error) error {
	file, err := os.Open(filePath) //nolint:gosec // G304: file path from CLI input is expected
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	// Handle large lines - Claude sessions can have very large message entries
	maxSize := s.MaxLineSize
	if maxSize == 0 {
		maxSize = 10 * 1024 * 1024
	}
	buf := make([]byte, 0, 64*1024) // 64KB initial buffer
	scanner.Buffer(buf, maxSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Quick JSON validation - must start with { or [
		trimmed := line
		for len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t') {
			trimmed = trimmed[1:]
		}
		if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
			continue
		}

		// Make a copy since scanner reuses the buffer
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)

		if err := fn(json.RawMessage(lineCopy)); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// ScanInto reads a JSONL file and unmarshals each line into type T.
// Lines that fail to unmarshal are silently skipped.
func ScanInto[T any](filePath string, fn func(entry T) error) error {
	s := NewScanner()
	return s.Scan(filePath, func(line json.RawMessage) error {
		var entry T
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip malformed entries
			return nil
		}
		return fn(entry)
	})
}

// ReadAll reads all entries from a JSONL file into a slice.
// This loads the entire file into memory - use Scan for large files.
func ReadAll[T any](filePath string) ([]T, error) {
	var results []T
	err := ScanInto(filePath, func(entry T) error {
		results = append(results, entry)
		return nil
	})
	return results, err
}

// CountLines counts the number of valid JSON lines in a JSONL file.
func CountLines(filePath string) (int, error) {
	count := 0
	s := NewScanner()
	err := s.Scan(filePath, func(line json.RawMessage) error {
		count++
		return nil
	})
	return count, err
}
