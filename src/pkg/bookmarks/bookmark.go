// Package bookmarks provides bookmark management for Claude Code agents.
package bookmarks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Bookmark represents a saved reference to a Claude Code agent conversation.
type Bookmark struct {
	BookmarkID        string     `json:"bookmark_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	AgentID           string     `json:"agent_id"`
	SessionID         string     `json:"session_id"`
	ProjectPath       string     `json:"project_path"`
	OriginalTimestamp time.Time  `json:"original_timestamp"`
	Hostname          string     `json:"hostname"`
	BookmarkedAt      time.Time  `json:"bookmarked_at"`
	BookmarkedBy      string     `json:"bookmarked_by"`
	Scope             string     `json:"scope"`
	Tags              []string   `json:"tags"`
	ResurrectionCount int        `json:"resurrection_count"`
	LastResurrected   *time.Time `json:"last_resurrected,omitempty"`
}

// Storage defines the interface for bookmark persistence.
type Storage interface {
	// Add creates a new bookmark
	Add(bookmark Bookmark) error

	// Get retrieves a bookmark by name
	Get(name string) (*Bookmark, error)

	// List returns all bookmarks
	List() ([]Bookmark, error)

	// Update modifies an existing bookmark
	Update(name string, updates map[string]interface{}) error

	// Delete removes a bookmark by name
	Delete(name string) error

	// Search finds bookmarks matching a query string
	Search(query string) ([]Bookmark, error)
}

// JSONLStorage implements Storage using JSONL files.
type JSONLStorage struct {
	path string
}

// NewJSONLStorage creates a new JSONLStorage instance.
func NewJSONLStorage(path string) (*JSONLStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &JSONLStorage{path: path}, nil
}

// Add creates a new bookmark.
func (s *JSONLStorage) Add(bookmark Bookmark) error {
	// Check for duplicate name
	existing, err := s.Get(bookmark.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("bookmark with name %q already exists", bookmark.Name)
	}

	// Generate bookmark ID if not set
	if bookmark.BookmarkID == "" {
		id, err := s.generateBookmarkID()
		if err != nil {
			return fmt.Errorf("failed to generate bookmark ID: %w", err)
		}
		bookmark.BookmarkID = id
	}

	// Set bookmarked_at if not set
	if bookmark.BookmarkedAt.IsZero() {
		bookmark.BookmarkedAt = time.Now()
	}

	// Read all existing bookmarks
	bookmarks, err := s.readAll()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing bookmarks: %w", err)
	}

	// Append new bookmark
	bookmarks = append(bookmarks, bookmark)

	// Write back
	return s.writeAll(bookmarks)
}

// Get retrieves a bookmark by name.
func (s *JSONLStorage) Get(name string) (*Bookmark, error) {
	bookmarks, err := s.readAll()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, b := range bookmarks {
		if b.Name == name {
			return &b, nil
		}
	}
	return nil, nil
}

// List returns all bookmarks.
func (s *JSONLStorage) List() ([]Bookmark, error) {
	bookmarks, err := s.readAll()
	if err != nil && os.IsNotExist(err) {
		return []Bookmark{}, nil
	}
	return bookmarks, err
}

// Update modifies an existing bookmark.
func (s *JSONLStorage) Update(name string, updates map[string]interface{}) error {
	bookmarks, err := s.readAll()
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("bookmark %q not found", name)
		}
		return err
	}

	found := false
	for i := range bookmarks {
		if bookmarks[i].Name == name {
			found = true
			if err := applyUpdates(&bookmarks[i], updates); err != nil {
				return err
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("bookmark %q not found", name)
	}

	return s.writeAll(bookmarks)
}

// Delete removes a bookmark by name.
func (s *JSONLStorage) Delete(name string) error {
	bookmarks, err := s.readAll()
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("bookmark %q not found", name)
		}
		return err
	}

	newBookmarks := make([]Bookmark, 0, len(bookmarks))
	found := false
	for _, b := range bookmarks {
		if b.Name == name {
			found = true
			continue
		}
		newBookmarks = append(newBookmarks, b)
	}

	if !found {
		return fmt.Errorf("bookmark %q not found", name)
	}

	return s.writeAll(newBookmarks)
}

// Search finds bookmarks matching a query string.
// Searches in name, description, and tags.
func (s *JSONLStorage) Search(query string) ([]Bookmark, error) {
	bookmarks, err := s.List()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return bookmarks, nil
	}

	query = strings.ToLower(query)
	results := make([]Bookmark, 0)

	for _, b := range bookmarks {
		// Search in name
		if strings.Contains(strings.ToLower(b.Name), query) {
			results = append(results, b)
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(b.Description), query) {
			results = append(results, b)
			continue
		}

		// Search in tags
		for _, tag := range b.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, b)
				break
			}
		}
	}

	return results, nil
}

// generateBookmarkID generates a new bookmark ID in format bmk-YYYY-MM-DD-NNN
func (s *JSONLStorage) generateBookmarkID() (string, error) {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	prefix := fmt.Sprintf("bmk-%s-", dateStr)

	bookmarks, err := s.readAll()
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// Find max counter for today
	maxCounter := 0
	for _, b := range bookmarks {
		if strings.HasPrefix(b.BookmarkID, prefix) {
			var counter int
			_, err := fmt.Sscanf(b.BookmarkID, prefix+"%d", &counter)
			if err == nil && counter > maxCounter {
				maxCounter = counter
			}
		}
	}

	return fmt.Sprintf("%s%03d", prefix, maxCounter+1), nil
}

// readAll reads all bookmarks from the JSONL file
func (s *JSONLStorage) readAll() ([]Bookmark, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var bookmarks []Bookmark
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var bookmark Bookmark
		if err := json.Unmarshal([]byte(line), &bookmark); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		bookmarks = append(bookmarks, bookmark)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return bookmarks, nil
}

// writeAll writes all bookmarks to the JSONL file
func (s *JSONLStorage) writeAll(bookmarks []Bookmark) error {
	// Sort bookmarks by name for consistent output
	sort.Slice(bookmarks, func(i, j int) bool {
		return bookmarks[i].Name < bookmarks[j].Name
	})

	// Write to temp file first
	tmpPath := s.path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	writer := bufio.NewWriter(file)
	for _, b := range bookmarks {
		data, err := json.Marshal(b)
		if err != nil {
			_ = file.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to marshal bookmark: %w", err)
		}

		if _, err := writer.Write(data); err != nil {
			_ = file.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to write bookmark: %w", err)
		}

		if err := writer.WriteByte('\n'); err != nil {
			_ = file.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		_ = file.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// applyUpdates applies a map of updates to a bookmark
func applyUpdates(bookmark *Bookmark, updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "name":
			if v, ok := value.(string); ok {
				bookmark.Name = v
			} else {
				return fmt.Errorf("invalid type for name: expected string")
			}
		case "description":
			if v, ok := value.(string); ok {
				bookmark.Description = v
			} else {
				return fmt.Errorf("invalid type for description: expected string")
			}
		case "tags":
			if v, ok := value.([]string); ok {
				bookmark.Tags = v
			} else if v, ok := value.([]interface{}); ok {
				tags := make([]string, len(v))
				for i, item := range v {
					if s, ok := item.(string); ok {
						tags[i] = s
					} else {
						return fmt.Errorf("invalid type for tags: expected []string")
					}
				}
				bookmark.Tags = tags
			} else {
				return fmt.Errorf("invalid type for tags: expected []string")
			}
		case "resurrection_count":
			if v, ok := value.(int); ok {
				bookmark.ResurrectionCount = v
			} else if v, ok := value.(float64); ok {
				bookmark.ResurrectionCount = int(v)
			} else {
				return fmt.Errorf("invalid type for resurrection_count: expected int")
			}
		case "last_resurrected":
			if v, ok := value.(time.Time); ok {
				bookmark.LastResurrected = &v
			} else if v, ok := value.(string); ok {
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return fmt.Errorf("invalid time format for last_resurrected: %w", err)
				}
				bookmark.LastResurrected = &t
			} else if value == nil {
				bookmark.LastResurrected = nil
			} else {
				return fmt.Errorf("invalid type for last_resurrected: expected time.Time or string")
			}
		default:
			return fmt.Errorf("unsupported update field: %s", key)
		}
	}
	return nil
}
