package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/bookmarks"
	"github.com/spf13/pflag"
)

// setupTestStorage creates a temporary bookmarks file for testing
func setupTestStorage(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	bookmarksFile := filepath.Join(tmpDir, "bookmarks.jsonl")

	// Set claudeDir to the temp directory
	origClaudeDir := claudeDir
	claudeDir = tmpDir

	t.Cleanup(func() {
		claudeDir = origClaudeDir
	})

	return bookmarksFile
}

// createTestBookmark creates a test bookmark in storage
func createTestBookmark(t *testing.T, name, agentID, sessionID string, tags []string) {
	t.Helper()

	storage, err := getStorage()
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	bookmark := bookmarks.Bookmark{
		Name:        name,
		Description: "Test bookmark for " + name,
		AgentID:     agentID,
		SessionID:   sessionID,
		ProjectPath: "",
		Hostname:    "test-host",
		Scope:       "global",
		Tags:        tags,
	}

	if err := storage.Add(bookmark); err != nil {
		t.Fatalf("Failed to add test bookmark: %v", err)
	}
}

func TestBookmarkAdd(t *testing.T) {
	tests := []struct {
		name        string
		bookmarkName string
		agentID     string
		sessionID   string
		projectPath string
		tags        []string
		description string
		wantErr     bool
		errContains string
	}{
		{
			name:         "valid bookmark",
			bookmarkName: "test-bookmark",
			agentID:      "agent-abc123",
			sessionID:    "session-xyz789",
			tags:         []string{"test", "example"},
			description:  "Test bookmark",
			wantErr:      false,
		},
		{
			name:         "bookmark with no tags",
			bookmarkName: "no-tags",
			agentID:      "agent-def456",
			sessionID:    "session-uvw456",
			tags:         []string{},
			description:  "Bookmark without tags",
			wantErr:      false,
		},
		{
			name:         "bookmark with underscore",
			bookmarkName: "test_bookmark",
			agentID:      "agent-ghi789",
			sessionID:    "session-rst123",
			tags:         []string{"test"},
			description:  "Test with underscore",
			wantErr:      false,
		},
		{
			name:         "invalid name - too long",
			bookmarkName: strings.Repeat("a", 65),
			agentID:      "agent-abc123",
			sessionID:    "session-xyz789",
			wantErr:      true,
			errContains:  "exceeds maximum length",
		},
		{
			name:         "invalid name - special chars",
			bookmarkName: "test@bookmark",
			agentID:      "agent-abc123",
			sessionID:    "session-xyz789",
			wantErr:      true,
			errContains:  "alphanumeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestStorage(t)

			// Set flags
			bookmarkName = tt.bookmarkName
			bookmarkAgentID = tt.agentID
			bookmarkSessionID = tt.sessionID
			bookmarkProjectPath = tt.projectPath
			bookmarkTags = tt.tags
			bookmarkDescription = tt.description

			err := runBookmarkAdd(nil, nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// Verify bookmark was created
					storage, err := getStorage()
					if err != nil {
						t.Fatalf("Failed to get storage: %v", err)
					}
					saved, err := storage.Get(tt.bookmarkName)
					if err != nil {
						t.Fatalf("Failed to get saved bookmark: %v", err)
					}
					if saved == nil {
						t.Errorf("Bookmark was not saved")
					} else {
						if saved.Name != tt.bookmarkName {
							t.Errorf("Name mismatch: got %q, want %q", saved.Name, tt.bookmarkName)
						}
						if saved.AgentID != tt.agentID {
							t.Errorf("AgentID mismatch: got %q, want %q", saved.AgentID, tt.agentID)
						}
						if saved.SessionID != tt.sessionID {
							t.Errorf("SessionID mismatch: got %q, want %q", saved.SessionID, tt.sessionID)
						}
						if saved.BookmarkID == "" {
							t.Errorf("BookmarkID was not generated")
						}
						if !strings.HasPrefix(saved.BookmarkID, "bmk-") {
							t.Errorf("BookmarkID has wrong format: %q", saved.BookmarkID)
						}
					}
				}
			}
		})
	}
}

func TestBookmarkList(t *testing.T) {
	setupTestStorage(t)

	// Create test bookmarks
	createTestBookmark(t, "bookmark1", "agent-1", "session-1", []string{"tag1", "tag2"})
	createTestBookmark(t, "bookmark2", "agent-2", "session-2", []string{"tag2", "tag3"})
	createTestBookmark(t, "bookmark3", "agent-3", "session-3", []string{"tag3"})

	tests := []struct {
		name        string
		filterTag   string
		wantCount   int
		wantNames   []string
	}{
		{
			name:      "list all",
			filterTag: "",
			wantCount: 3,
			wantNames: []string{"bookmark1", "bookmark2", "bookmark3"},
		},
		{
			name:      "filter by tag1",
			filterTag: "tag1",
			wantCount: 1,
			wantNames: []string{"bookmark1"},
		},
		{
			name:      "filter by tag2",
			filterTag: "tag2",
			wantCount: 2,
			wantNames: []string{"bookmark1", "bookmark2"},
		},
		{
			name:      "filter by tag3",
			filterTag: "tag3",
			wantCount: 2,
			wantNames: []string{"bookmark2", "bookmark3"},
		},
		{
			name:      "filter by nonexistent tag",
			filterTag: "tag99",
			wantCount: 0,
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarkFilterTag = tt.filterTag
			format = "json"

			// Capture output
			oldFormat := format
			format = "json"
			defer func() { format = oldFormat }()

			// Get bookmarks directly to verify
			storage, err := getStorage()
			if err != nil {
				t.Fatalf("Failed to get storage: %v", err)
			}

			all, err := storage.List()
			if err != nil {
				t.Fatalf("Failed to list bookmarks: %v", err)
			}

			// Filter
			var filtered []bookmarks.Bookmark
			if tt.filterTag == "" {
				filtered = all
			} else {
				for _, b := range all {
					for _, tag := range b.Tags {
						if tag == tt.filterTag {
							filtered = append(filtered, b)
							break
						}
					}
				}
			}

			if len(filtered) != tt.wantCount {
				t.Errorf("Got %d bookmarks, want %d", len(filtered), tt.wantCount)
			}

			// Verify names
			gotNames := make(map[string]bool)
			for _, b := range filtered {
				gotNames[b.Name] = true
			}
			for _, name := range tt.wantNames {
				if !gotNames[name] {
					t.Errorf("Expected bookmark %q not found in results", name)
				}
			}
		})
	}
}

func TestBookmarkGet(t *testing.T) {
	setupTestStorage(t)

	// Create test bookmark
	createTestBookmark(t, "test-bookmark", "agent-abc", "session-xyz", []string{"tag1"})

	tests := []struct {
		name        string
		bookmarkName string
		wantErr     bool
		errContains string
	}{
		{
			name:         "existing bookmark",
			bookmarkName: "test-bookmark",
			wantErr:      false,
		},
		{
			name:         "nonexistent bookmark",
			bookmarkName: "does-not-exist",
			wantErr:      true,
			errContains:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format = "json"

			err := runBookmarkGet(nil, []string{tt.bookmarkName})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBookmarkUpdate(t *testing.T) {
	tests := []struct {
		name            string
		bookmarkName    string
		newDescription  string
		addTags         []string
		setDescription  bool
		setTags         bool
		wantErr         bool
		errContains     string
	}{
		{
			name:           "update description",
			bookmarkName:   "test-bookmark",
			newDescription: "Updated description",
			setDescription: true,
			wantErr:        false,
		},
		{
			name:         "add tags",
			bookmarkName: "test-bookmark",
			addTags:      []string{"new-tag"},
			setTags:      true,
			wantErr:      false,
		},
		{
			name:           "update both",
			bookmarkName:   "test-bookmark",
			newDescription: "New description",
			addTags:        []string{"tag2"},
			setDescription: true,
			setTags:        true,
			wantErr:        false,
		},
		{
			name:         "no updates",
			bookmarkName: "test-bookmark",
			wantErr:      true,
			errContains:  "no updates specified",
		},
		{
			name:           "nonexistent bookmark",
			bookmarkName:   "does-not-exist",
			newDescription: "Description",
			setDescription: true,
			wantErr:        true,
			errContains:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestStorage(t)

			// Create test bookmark
			createTestBookmark(t, "test-bookmark", "agent-abc", "session-xyz", []string{"tag1"})

			// Reset flags before each test
			cmd := bookmarkUpdateCmd
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				f.Changed = false
			})
			bookmarkDescription = ""
			bookmarkAddTags = []string{}

			// Mock cobra command for flag checking
			if tt.setDescription {
				bookmarkDescription = tt.newDescription
				cmd.Flags().Set("description", tt.newDescription)
			}
			if tt.setTags {
				bookmarkAddTags = tt.addTags
				cmd.Flags().Set("add-tags", strings.Join(tt.addTags, ","))
			}

			err := runBookmarkUpdate(cmd, []string{tt.bookmarkName})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// Verify update
					storage, err := getStorage()
					if err != nil {
						t.Fatalf("Failed to get storage: %v", err)
					}
					updated, err := storage.Get(tt.bookmarkName)
					if err != nil {
						t.Fatalf("Failed to get updated bookmark: %v", err)
					}

					if tt.setDescription && updated.Description != tt.newDescription {
						t.Errorf("Description not updated: got %q, want %q", updated.Description, tt.newDescription)
					}
					if tt.setTags {
						hasTag := false
						for _, tag := range updated.Tags {
							for _, newTag := range tt.addTags {
								if tag == newTag {
									hasTag = true
									break
								}
							}
						}
						if !hasTag {
							t.Errorf("Tags not updated: got %v, expected to contain %v", updated.Tags, tt.addTags)
						}
					}
				}
			}

			// Reset flags
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				f.Changed = false
			})
			bookmarkDescription = ""
			bookmarkAddTags = []string{}
		})
	}
}

func TestBookmarkDelete(t *testing.T) {
	tests := []struct {
		name         string
		bookmarkName string
		force        bool
		wantErr      bool
		errContains  string
	}{
		{
			name:         "delete with force",
			bookmarkName: "test-bookmark",
			force:        true,
			wantErr:      false,
		},
		{
			name:         "delete nonexistent",
			bookmarkName: "does-not-exist",
			force:        true,
			wantErr:      true,
			errContains:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestStorage(t)

			// Create test bookmark
			createTestBookmark(t, "test-bookmark", "agent-abc", "session-xyz", []string{"tag1"})

			bookmarkForce = tt.force

			err := runBookmarkDelete(nil, []string{tt.bookmarkName})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// Verify deletion
					storage, err := getStorage()
					if err != nil {
						t.Fatalf("Failed to get storage: %v", err)
					}
					deleted, err := storage.Get(tt.bookmarkName)
					if err != nil {
						t.Fatalf("Failed to check deleted bookmark: %v", err)
					}
					if deleted != nil {
						t.Errorf("Bookmark was not deleted")
					}
				}
			}
		})
	}
}

func TestBookmarkSearch(t *testing.T) {
	setupTestStorage(t)

	// Create test bookmarks with different attributes
	storage, err := getStorage()
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	testBookmarks := []bookmarks.Bookmark{
		{
			Name:        "python-expert",
			Description: "Expert in Python programming",
			AgentID:     "agent-1",
			SessionID:   "session-1",
			Hostname:    "test-host",
			Scope:       "global",
			Tags:        []string{"python", "programming"},
		},
		{
			Name:        "go-expert",
			Description: "Expert in Go development",
			AgentID:     "agent-2",
			SessionID:   "session-2",
			Hostname:    "test-host",
			Scope:       "global",
			Tags:        []string{"go", "programming"},
		},
		{
			Name:        "architecture-review",
			Description: "Architecture review specialist",
			AgentID:     "agent-3",
			SessionID:   "session-3",
			Hostname:    "test-host",
			Scope:       "global",
			Tags:        []string{"architecture", "design"},
		},
	}

	for _, b := range testBookmarks {
		if err := storage.Add(b); err != nil {
			t.Fatalf("Failed to add bookmark: %v", err)
		}
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantNames []string
	}{
		{
			name:      "search by name",
			query:     "python",
			wantCount: 1,
			wantNames: []string{"python-expert"},
		},
		{
			name:      "search by description",
			query:     "Expert",
			wantCount: 2,
			wantNames: []string{"python-expert", "go-expert"},
		},
		{
			name:      "search by tag",
			query:     "programming",
			wantCount: 2,
			wantNames: []string{"python-expert", "go-expert"},
		},
		{
			name:      "search no match",
			query:     "javascript",
			wantCount: 0,
			wantNames: []string{},
		},
		{
			name:      "search architecture",
			query:     "architecture",
			wantCount: 1,
			wantNames: []string{"architecture-review"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format = "json"

			// Get results directly
			results, err := storage.Search(tt.query)
			if err != nil {
				t.Fatalf("Failed to search: %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("Got %d results, want %d", len(results), tt.wantCount)
			}

			gotNames := make(map[string]bool)
			for _, b := range results {
				gotNames[b.Name] = true
			}
			for _, name := range tt.wantNames {
				if !gotNames[name] {
					t.Errorf("Expected result %q not found", name)
				}
			}
		})
	}
}

func TestBookmarkIDGeneration(t *testing.T) {
	setupTestStorage(t)

	storage, err := getStorage()
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	// Add multiple bookmarks on the same day
	today := time.Now().Format("2006-01-02")
	expectedPrefix := "bmk-" + today + "-"

	for i := 1; i <= 3; i++ {
		bookmark := bookmarks.Bookmark{
			Name:      fmt.Sprintf("bookmark-%d", i),
			AgentID:   fmt.Sprintf("agent-%d", i),
			SessionID: fmt.Sprintf("session-%d", i),
			Hostname:  "test-host",
			Scope:     "global",
		}

		if err := storage.Add(bookmark); err != nil {
			t.Fatalf("Failed to add bookmark: %v", err)
		}

		saved, err := storage.Get(bookmark.Name)
		if err != nil {
			t.Fatalf("Failed to get saved bookmark: %v", err)
		}

		if !strings.HasPrefix(saved.BookmarkID, expectedPrefix) {
			t.Errorf("Bookmark ID has wrong prefix: got %q, want prefix %q", saved.BookmarkID, expectedPrefix)
		}

		// Verify counter
		expectedID := fmt.Sprintf("%s%03d", expectedPrefix, i)
		if saved.BookmarkID != expectedID {
			t.Errorf("Bookmark ID mismatch: got %q, want %q", saved.BookmarkID, expectedID)
		}
	}
}

func TestBookmarkJSONOutput(t *testing.T) {
	setupTestStorage(t)

	createTestBookmark(t, "json-test", "agent-abc", "session-xyz", []string{"tag1"})

	format = "json"

	storage, err := getStorage()
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	bookmark, err := storage.Get("json-test")
	if err != nil {
		t.Fatalf("Failed to get bookmark: %v", err)
	}

	// Verify JSON marshaling works
	data, err := json.MarshalIndent(bookmark, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Verify we can unmarshal back
	var unmarshaled bookmarks.Bookmark
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unmarshaled.Name != bookmark.Name {
		t.Errorf("Name mismatch after unmarshal: got %q, want %q", unmarshaled.Name, bookmark.Name)
	}
}
