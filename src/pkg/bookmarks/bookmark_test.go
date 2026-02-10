package bookmarks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper function to create a temp directory for testing
func createTempDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "bookmarks-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})
	return tmpDir
}

// Helper function to create a test bookmark
func createTestBookmark(name string) Bookmark {
	now := time.Now()
	return Bookmark{
		Name:              name,
		Description:       "Test bookmark " + name,
		AgentID:           "test-agent-123",
		SessionID:         "test-session-456",
		ProjectPath:       "/test/project",
		OriginalTimestamp: now,
		Hostname:          "test-host",
		BookmarkedBy:      "test-user",
		Scope:             "global",
		Tags:              []string{"test", "bookmark"},
		ResurrectionCount: 0,
	}
}

func TestNewJSONLStorage(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")

	storage, err := NewJSONLStorage(path)
	if err != nil {
		t.Fatalf("NewJSONLStorage failed: %v", err)
	}

	if storage.path != path {
		t.Errorf("expected path %q, got %q", path, storage.path)
	}

	// Verify directory was created
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

func TestAdd(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	bookmark := createTestBookmark("test-bookmark-1")

	// Test adding bookmark
	err := storage.Add(bookmark)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Verify bookmark ID was generated
	retrieved, err := storage.Get("test-bookmark-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("bookmark not found")
	}
	if retrieved.BookmarkID == "" {
		t.Error("bookmark ID was not generated")
	}
	if !retrieved.BookmarkedAt.After(time.Time{}) {
		t.Error("bookmarked_at was not set")
	}

	// Test duplicate name
	err = storage.Add(bookmark)
	if err == nil {
		t.Error("expected error for duplicate name, got nil")
	}
}

func TestGet(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	// Get from empty storage
	bookmark, err := storage.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if bookmark != nil {
		t.Error("expected nil for nonexistent bookmark")
	}

	// Add and get
	testBookmark := createTestBookmark("test-get")
	_ = storage.Add(testBookmark)

	bookmark, err = storage.Get("test-get")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if bookmark == nil {
		t.Fatal("bookmark not found")
	}
	if bookmark.Name != "test-get" {
		t.Errorf("expected name %q, got %q", "test-get", bookmark.Name)
	}
}

func TestList(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	// List empty storage
	bookmarks, err := storage.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Errorf("expected 0 bookmarks, got %d", len(bookmarks))
	}

	// Add multiple bookmarks
	_ = storage.Add(createTestBookmark("bookmark-1"))
	_ = storage.Add(createTestBookmark("bookmark-2"))
	_ = storage.Add(createTestBookmark("bookmark-3"))

	bookmarks, err = storage.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(bookmarks) != 3 {
		t.Errorf("expected 3 bookmarks, got %d", len(bookmarks))
	}

	// Verify sorted by name
	if bookmarks[0].Name != "bookmark-1" || bookmarks[1].Name != "bookmark-2" || bookmarks[2].Name != "bookmark-3" {
		t.Error("bookmarks not sorted by name")
	}
}

func TestUpdate(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	// Test updating nonexistent bookmark
	err := storage.Update("nonexistent", map[string]interface{}{"description": "new"})
	if err == nil {
		t.Error("expected error for nonexistent bookmark")
	}

	// Add bookmark
	_ = storage.Add(createTestBookmark("test-update"))

	// Update description
	err = storage.Update("test-update", map[string]interface{}{
		"description": "Updated description",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	bookmark, _ := storage.Get("test-update")
	if bookmark.Description != "Updated description" {
		t.Errorf("expected description %q, got %q", "Updated description", bookmark.Description)
	}

	// Update tags
	err = storage.Update("test-update", map[string]interface{}{
		"tags": []string{"new", "tags"},
	})
	if err != nil {
		t.Fatalf("Update tags failed: %v", err)
	}

	bookmark, _ = storage.Get("test-update")
	if len(bookmark.Tags) != 2 || bookmark.Tags[0] != "new" || bookmark.Tags[1] != "tags" {
		t.Errorf("expected tags [new tags], got %v", bookmark.Tags)
	}

	// Update resurrection count
	err = storage.Update("test-update", map[string]interface{}{
		"resurrection_count": 5,
	})
	if err != nil {
		t.Fatalf("Update resurrection_count failed: %v", err)
	}

	bookmark, _ = storage.Get("test-update")
	if bookmark.ResurrectionCount != 5 {
		t.Errorf("expected resurrection_count 5, got %d", bookmark.ResurrectionCount)
	}

	// Update last_resurrected
	now := time.Now()
	err = storage.Update("test-update", map[string]interface{}{
		"last_resurrected": now,
	})
	if err != nil {
		t.Fatalf("Update last_resurrected failed: %v", err)
	}

	bookmark, _ = storage.Get("test-update")
	if bookmark.LastResurrected == nil {
		t.Error("last_resurrected was not set")
	} else if !bookmark.LastResurrected.Equal(now) {
		t.Errorf("expected last_resurrected %v, got %v", now, *bookmark.LastResurrected)
	}

	// Test invalid update field
	err = storage.Update("test-update", map[string]interface{}{
		"invalid_field": "value",
	})
	if err == nil {
		t.Error("expected error for invalid field")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	// Test deleting nonexistent bookmark
	err := storage.Delete("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent bookmark")
	}

	// Add bookmarks
	_ = storage.Add(createTestBookmark("bookmark-1"))
	_ = storage.Add(createTestBookmark("bookmark-2"))
	_ = storage.Add(createTestBookmark("bookmark-3"))

	// Delete one
	err = storage.Delete("bookmark-2")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	bookmarks, _ := storage.List()
	if len(bookmarks) != 2 {
		t.Errorf("expected 2 bookmarks after delete, got %d", len(bookmarks))
	}

	bookmark, _ := storage.Get("bookmark-2")
	if bookmark != nil {
		t.Error("deleted bookmark still exists")
	}
}

func TestSearch(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	// Add test bookmarks
	b1 := createTestBookmark("web-server")
	b1.Description = "Main web server implementation"
	b1.Tags = []string{"backend", "http"}
	_ = storage.Add(b1)

	b2 := createTestBookmark("database-schema")
	b2.Description = "Database schema migrations"
	b2.Tags = []string{"database", "sql"}
	_ = storage.Add(b2)

	b3 := createTestBookmark("api-client")
	b3.Description = "REST API client implementation"
	b3.Tags = []string{"backend", "api"}
	_ = storage.Add(b3)

	tests := []struct {
		query    string
		expected int
		names    []string
	}{
		{
			query:    "",
			expected: 3,
			names:    []string{"api-client", "database-schema", "web-server"},
		},
		{
			query:    "web",
			expected: 1,
			names:    []string{"web-server"},
		},
		{
			query:    "backend",
			expected: 2,
			names:    []string{"api-client", "web-server"},
		},
		{
			query:    "database",
			expected: 1,
			names:    []string{"database-schema"},
		},
		{
			query:    "implementation",
			expected: 2,
			names:    []string{"api-client", "web-server"},
		},
		{
			query:    "nonexistent",
			expected: 0,
			names:    []string{},
		},
	}

	for _, tt := range tests {
		results, err := storage.Search(tt.query)
		if err != nil {
			t.Fatalf("Search(%q) failed: %v", tt.query, err)
		}
		if len(results) != tt.expected {
			t.Errorf("Search(%q): expected %d results, got %d", tt.query, tt.expected, len(results))
		}
		for i, expectedName := range tt.names {
			if i >= len(results) {
				t.Errorf("Search(%q): missing result %q", tt.query, expectedName)
				break
			}
			if results[i].Name != expectedName {
				t.Errorf("Search(%q): expected result[%d] name %q, got %q", tt.query, i, expectedName, results[i].Name)
			}
		}
	}
}

func TestGenerateBookmarkID(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	// Generate first ID
	id1, err := storage.generateBookmarkID()
	if err != nil {
		t.Fatalf("generateBookmarkID failed: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	expectedPrefix := "bmk-" + today + "-"
	if !strings.HasPrefix(id1, expectedPrefix) {
		t.Errorf("expected ID to start with %q, got %q", expectedPrefix, id1)
	}
	if id1 != expectedPrefix+"001" {
		t.Errorf("expected first ID to be %q, got %q", expectedPrefix+"001", id1)
	}

	// Add bookmark with generated ID
	b1 := createTestBookmark("bookmark-1")
	b1.BookmarkID = id1
	_ = storage.Add(b1)

	// Generate second ID
	id2, err := storage.generateBookmarkID()
	if err != nil {
		t.Fatalf("generateBookmarkID failed: %v", err)
	}
	if id2 != expectedPrefix+"002" {
		t.Errorf("expected second ID to be %q, got %q", expectedPrefix+"002", id2)
	}

	// Add multiple bookmarks
	for i := 2; i <= 10; i++ {
		b := createTestBookmark("bookmark-" + string(rune('0'+i)))
		_ = storage.Add(b)
	}

	// Generate 11th ID
	id11, err := storage.generateBookmarkID()
	if err != nil {
		t.Fatalf("generateBookmarkID failed: %v", err)
	}
	if id11 != expectedPrefix+"011" {
		t.Errorf("expected 11th ID to be %q, got %q", expectedPrefix+"011", id11)
	}
}

func TestJSONLPersistence(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")

	// Create storage and add bookmarks
	storage1, _ := NewJSONLStorage(path)
	b1 := createTestBookmark("bookmark-1")
	b2 := createTestBookmark("bookmark-2")
	_ = storage1.Add(b1)
	_ = storage1.Add(b2)

	// Create new storage instance and verify data persists
	storage2, _ := NewJSONLStorage(path)
	bookmarks, err := storage2.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(bookmarks) != 2 {
		t.Errorf("expected 2 bookmarks, got %d", len(bookmarks))
	}

	// Verify content
	if bookmarks[0].Name != "bookmark-1" || bookmarks[1].Name != "bookmark-2" {
		t.Error("bookmark data not persisted correctly")
	}
}

func TestEmptyFile(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")

	// Create empty file
	file, _ := os.Create(path)
	_ = file.Close()

	storage, _ := NewJSONLStorage(path)
	bookmarks, err := storage.List()
	if err != nil {
		t.Fatalf("List on empty file failed: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Errorf("expected 0 bookmarks from empty file, got %d", len(bookmarks))
	}
}

func TestMalformedJSON(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")

	// Write malformed JSON
	file, _ := os.Create(path)
	_, _ = file.WriteString("{invalid json}\n")
	_ = file.Close()

	storage, _ := NewJSONLStorage(path)
	_, err := storage.List()
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestBlankLines(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")

	// Create storage and add bookmark
	storage, _ := NewJSONLStorage(path)
	_ = storage.Add(createTestBookmark("test-bookmark"))

	// Manually add blank lines
	file, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	_, _ = file.WriteString("\n\n")
	_ = file.Close()

	// Verify it still works
	bookmarks, err := storage.List()
	if err != nil {
		t.Fatalf("List with blank lines failed: %v", err)
	}
	if len(bookmarks) != 1 {
		t.Errorf("expected 1 bookmark, got %d", len(bookmarks))
	}
}

func TestUpdateWithInterfaceSlice(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	_ = storage.Add(createTestBookmark("test-update"))

	// Update tags with []interface{} (as would come from JSON unmarshaling)
	err := storage.Update("test-update", map[string]interface{}{
		"tags": []interface{}{"tag1", "tag2", "tag3"},
	})
	if err != nil {
		t.Fatalf("Update with []interface{} failed: %v", err)
	}

	bookmark, _ := storage.Get("test-update")
	if len(bookmark.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(bookmark.Tags))
	}
}

func TestUpdateWithFloat64(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	_ = storage.Add(createTestBookmark("test-update"))

	// Update resurrection_count with float64 (as would come from JSON unmarshaling)
	err := storage.Update("test-update", map[string]interface{}{
		"resurrection_count": float64(42),
	})
	if err != nil {
		t.Fatalf("Update with float64 failed: %v", err)
	}

	bookmark, _ := storage.Get("test-update")
	if bookmark.ResurrectionCount != 42 {
		t.Errorf("expected resurrection_count 42, got %d", bookmark.ResurrectionCount)
	}
}

func TestUpdateWithTimeString(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	_ = storage.Add(createTestBookmark("test-update"))

	// Update last_resurrected with string (as would come from JSON unmarshaling)
	timeStr := "2026-02-09T12:00:00Z"
	err := storage.Update("test-update", map[string]interface{}{
		"last_resurrected": timeStr,
	})
	if err != nil {
		t.Fatalf("Update with time string failed: %v", err)
	}

	bookmark, _ := storage.Get("test-update")
	if bookmark.LastResurrected == nil {
		t.Fatal("last_resurrected was not set")
	}

	expectedTime, _ := time.Parse(time.RFC3339, timeStr)
	if !bookmark.LastResurrected.Equal(expectedTime) {
		t.Errorf("expected last_resurrected %v, got %v", expectedTime, *bookmark.LastResurrected)
	}
}

func TestAtomicWrite(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	// Add bookmark
	_ = storage.Add(createTestBookmark("bookmark-1"))

	// Verify temp file is cleaned up
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); err == nil {
		t.Error("temp file was not cleaned up")
	}

	// Verify original file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("original file does not exist")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	now := time.Now()
	original := Bookmark{
		Name:              "test-roundtrip",
		Description:       "Test JSON roundtrip",
		AgentID:           "agent-123",
		SessionID:         "session-456",
		ProjectPath:       "/test/project",
		OriginalTimestamp: now,
		Hostname:          "test-host",
		BookmarkedBy:      "test-user",
		Scope:             "global",
		Tags:              []string{"test", "roundtrip"},
		ResurrectionCount: 5,
		LastResurrected:   &now,
	}

	// Add bookmark
	_ = storage.Add(original)

	// Retrieve and compare
	retrieved, err := storage.Get("test-roundtrip")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("bookmark not found")
	}

	// Marshal both to JSON and compare
	originalJSON, _ := json.Marshal(original)
	retrievedJSON, _ := json.Marshal(*retrieved)

	var originalMap, retrievedMap map[string]interface{}
	_ = json.Unmarshal(originalJSON, &originalMap)
	_ = json.Unmarshal(retrievedJSON, &retrievedMap)

	// Compare key fields (excluding bookmark_id and bookmarked_at which are generated)
	compareFields := []string{"name", "description", "agent_id", "session_id", "project_path", "hostname", "bookmarked_by", "scope", "resurrection_count"}
	for _, field := range compareFields {
		if originalMap[field] != retrievedMap[field] {
			t.Errorf("field %q: expected %v, got %v", field, originalMap[field], retrievedMap[field])
		}
	}
}

func TestUpdateErrors(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	_ = storage.Add(createTestBookmark("test-bookmark"))

	tests := []struct {
		name      string
		updates   map[string]interface{}
		wantError bool
	}{
		{
			name:      "invalid name type",
			updates:   map[string]interface{}{"name": 123},
			wantError: true,
		},
		{
			name:      "invalid description type",
			updates:   map[string]interface{}{"description": 123},
			wantError: true,
		},
		{
			name:      "invalid tags type",
			updates:   map[string]interface{}{"tags": "not a slice"},
			wantError: true,
		},
		{
			name:      "invalid tags element type",
			updates:   map[string]interface{}{"tags": []interface{}{123, "valid"}},
			wantError: true,
		},
		{
			name:      "invalid resurrection_count type",
			updates:   map[string]interface{}{"resurrection_count": "not a number"},
			wantError: true,
		},
		{
			name:      "invalid last_resurrected type",
			updates:   map[string]interface{}{"last_resurrected": 123},
			wantError: true,
		},
		{
			name:      "invalid last_resurrected time string",
			updates:   map[string]interface{}{"last_resurrected": "invalid time"},
			wantError: true,
		},
		{
			name:      "unsupported field",
			updates:   map[string]interface{}{"unknown_field": "value"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.Update("test-bookmark", tt.updates)
			if (err != nil) != tt.wantError {
				t.Errorf("Update() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestUpdateName(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	_ = storage.Add(createTestBookmark("old-name"))

	err := storage.Update("old-name", map[string]interface{}{"name": "new-name"})
	if err != nil {
		t.Fatalf("Update name failed: %v", err)
	}

	// Old name should not exist
	bookmark, _ := storage.Get("old-name")
	if bookmark != nil {
		t.Error("old name still exists after update")
	}

	// New name should exist
	bookmark, _ = storage.Get("new-name")
	if bookmark == nil {
		t.Fatal("new name not found after update")
	}
	if bookmark.Name != "new-name" {
		t.Errorf("expected name %q, got %q", "new-name", bookmark.Name)
	}
}

func TestUpdateLastResurrectedNull(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	b := createTestBookmark("test-bookmark")
	now := time.Now()
	b.LastResurrected = &now
	_ = storage.Add(b)

	// Set last_resurrected to nil
	err := storage.Update("test-bookmark", map[string]interface{}{"last_resurrected": nil})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	bookmark, _ := storage.Get("test-bookmark")
	if bookmark.LastResurrected != nil {
		t.Error("last_resurrected should be nil after update")
	}
}

func TestAddWithCustomBookmarkID(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	b := createTestBookmark("test-bookmark")
	b.BookmarkID = "custom-id-123"

	err := storage.Add(b)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	retrieved, _ := storage.Get("test-bookmark")
	if retrieved.BookmarkID != "custom-id-123" {
		t.Errorf("expected bookmark_id %q, got %q", "custom-id-123", retrieved.BookmarkID)
	}
}

func TestAddWithCustomBookmarkedAt(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	b := createTestBookmark("test-bookmark")
	customTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	b.BookmarkedAt = customTime

	err := storage.Add(b)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	retrieved, _ := storage.Get("test-bookmark")
	if !retrieved.BookmarkedAt.Equal(customTime) {
		t.Errorf("expected bookmarked_at %v, got %v", customTime, retrieved.BookmarkedAt)
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	tmpDir := createTempDir(t)
	path := filepath.Join(tmpDir, "bookmarks.jsonl")
	storage, _ := NewJSONLStorage(path)

	b := createTestBookmark("Test-Bookmark")
	b.Description = "Test Description"
	b.Tags = []string{"Test-Tag"}
	_ = storage.Add(b)

	// Test case insensitive search
	tests := []string{"test", "TEST", "TeSt", "BOOKMARK", "description", "tag"}
	for _, query := range tests {
		results, err := storage.Search(query)
		if err != nil {
			t.Fatalf("Search(%q) failed: %v", query, err)
		}
		if len(results) != 1 {
			t.Errorf("Search(%q): expected 1 result, got %d", query, len(results))
		}
	}
}
