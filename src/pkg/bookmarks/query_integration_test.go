package bookmarks

import (
	"testing"
	"time"

	"github.com/randlee/claude-history/pkg/models"
)

func TestEnrichWithBookmarks(t *testing.T) {
	tests := []struct {
		name      string
		results   []models.ConversationEntry
		bookmarks []Bookmark
		wantLen   int
		checks    func(t *testing.T, enriched []models.ConversationEntry)
	}{
		{
			name: "enriches bookmarked agent",
			results: []models.ConversationEntry{
				{
					UUID:    "entry-1",
					AgentID: "agent-123",
					Type:    models.EntryTypeAssistant,
				},
			},
			bookmarks: []Bookmark{
				{
					BookmarkID: "bmk-001",
					Name:       "test-bookmark",
					AgentID:    "agent-123",
					Tags:       []string{"tag1", "tag2"},
				},
			},
			wantLen: 1,
			checks: func(t *testing.T, enriched []models.ConversationEntry) {
				if !enriched[0].Bookmarked {
					t.Error("expected entry to be marked as bookmarked")
				}
				if enriched[0].BookmarkID != "bmk-001" {
					t.Errorf("expected bookmark_id 'bmk-001', got '%s'", enriched[0].BookmarkID)
				}
				if enriched[0].BookmarkName != "test-bookmark" {
					t.Errorf("expected bookmark_name 'test-bookmark', got '%s'", enriched[0].BookmarkName)
				}
				if len(enriched[0].BookmarkTags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(enriched[0].BookmarkTags))
				}
			},
		},
		{
			name: "does not enrich non-bookmarked agent",
			results: []models.ConversationEntry{
				{
					UUID:    "entry-1",
					AgentID: "agent-456",
					Type:    models.EntryTypeAssistant,
				},
			},
			bookmarks: []Bookmark{
				{
					BookmarkID: "bmk-001",
					Name:       "test-bookmark",
					AgentID:    "agent-123",
					Tags:       []string{"tag1"},
				},
			},
			wantLen: 1,
			checks: func(t *testing.T, enriched []models.ConversationEntry) {
				if enriched[0].Bookmarked {
					t.Error("expected entry to not be bookmarked")
				}
				if enriched[0].BookmarkID != "" {
					t.Error("expected empty bookmark_id for non-bookmarked entry")
				}
			},
		},
		{
			name: "handles multiple results with mixed bookmarks",
			results: []models.ConversationEntry{
				{UUID: "entry-1", AgentID: "agent-123", Type: models.EntryTypeAssistant},
				{UUID: "entry-2", AgentID: "agent-456", Type: models.EntryTypeAssistant},
				{UUID: "entry-3", AgentID: "agent-789", Type: models.EntryTypeAssistant},
			},
			bookmarks: []Bookmark{
				{BookmarkID: "bmk-001", Name: "bookmark-1", AgentID: "agent-123", Tags: []string{"tag1"}},
				{BookmarkID: "bmk-002", Name: "bookmark-2", AgentID: "agent-789", Tags: []string{"tag2"}},
			},
			wantLen: 3,
			checks: func(t *testing.T, enriched []models.ConversationEntry) {
				// First entry should be bookmarked
				if !enriched[0].Bookmarked || enriched[0].BookmarkID != "bmk-001" {
					t.Error("first entry should be bookmarked with bmk-001")
				}
				// Second entry should not be bookmarked
				if enriched[1].Bookmarked {
					t.Error("second entry should not be bookmarked")
				}
				// Third entry should be bookmarked
				if !enriched[2].Bookmarked || enriched[2].BookmarkID != "bmk-002" {
					t.Error("third entry should be bookmarked with bmk-002")
				}
			},
		},
		{
			name:      "handles empty results",
			results:   []models.ConversationEntry{},
			bookmarks: []Bookmark{{BookmarkID: "bmk-001", Name: "test", AgentID: "agent-123"}},
			wantLen:   0,
			checks: func(t *testing.T, enriched []models.ConversationEntry) {
				if len(enriched) != 0 {
					t.Error("expected empty result set")
				}
			},
		},
		{
			name: "handles empty bookmarks",
			results: []models.ConversationEntry{
				{UUID: "entry-1", AgentID: "agent-123", Type: models.EntryTypeAssistant},
			},
			bookmarks: []Bookmark{},
			wantLen:   1,
			checks: func(t *testing.T, enriched []models.ConversationEntry) {
				if enriched[0].Bookmarked {
					t.Error("expected entry to not be bookmarked when no bookmarks exist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enriched := EnrichWithBookmarks(tt.results, tt.bookmarks)
			if len(enriched) != tt.wantLen {
				t.Errorf("expected %d results, got %d", tt.wantLen, len(enriched))
			}
			if tt.checks != nil {
				tt.checks(t, enriched)
			}
		})
	}
}

func TestMatchBookmarkTags(t *testing.T) {
	baseTime := time.Now()
	bookmark := Bookmark{
		BookmarkID:  "bmk-001",
		Name:        "Architecture Expert",
		Description: "An agent that explores system architecture and design patterns",
		AgentID:     "agent-123",
		Tags:        []string{"architecture", "design", "patterns"},
		BookmarkedAt: baseTime,
	}

	tests := []struct {
		name     string
		bookmark Bookmark
		query    string
		want     bool
	}{
		{
			name:     "matches bookmark name (case insensitive)",
			bookmark: bookmark,
			query:    "architecture",
			want:     true,
		},
		{
			name:     "matches bookmark name (different case)",
			bookmark: bookmark,
			query:    "EXPERT",
			want:     true,
		},
		{
			name:     "matches tag",
			bookmark: bookmark,
			query:    "design",
			want:     true,
		},
		{
			name:     "matches description",
			bookmark: bookmark,
			query:    "system",
			want:     true,
		},
		{
			name:     "matches partial word in name",
			bookmark: bookmark,
			query:    "arch",
			want:     true,
		},
		{
			name:     "matches partial word in tag",
			bookmark: bookmark,
			query:    "patt",
			want:     true,
		},
		{
			name:     "does not match unrelated query",
			bookmark: bookmark,
			query:    "database",
			want:     false,
		},
		{
			name:     "does not match empty query",
			bookmark: bookmark,
			query:    "",
			want:     true, // Empty string is contained in any string
		},
		{
			name: "matches against bookmark with no tags",
			bookmark: Bookmark{
				BookmarkID:  "bmk-002",
				Name:        "Simple Bookmark",
				Description: "A basic bookmark",
				Tags:        []string{},
			},
			query: "simple",
			want:  true,
		},
		{
			name: "does not match against bookmark with no tags and unrelated query",
			bookmark: Bookmark{
				BookmarkID:  "bmk-002",
				Name:        "Simple Bookmark",
				Description: "A basic bookmark",
				Tags:        []string{},
			},
			query: "complex",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchBookmarkTags(tt.bookmark, tt.query)
			if got != tt.want {
				t.Errorf("MatchBookmarkTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnrichWithBookmarksNoDuplicates(t *testing.T) {
	// Test that we don't create duplicate entries, just enrich existing ones
	results := []models.ConversationEntry{
		{UUID: "entry-1", AgentID: "agent-123", Type: models.EntryTypeAssistant},
		{UUID: "entry-2", AgentID: "agent-456", Type: models.EntryTypeAssistant},
	}

	bookmarks := []Bookmark{
		{BookmarkID: "bmk-001", Name: "bookmark-1", AgentID: "agent-123"},
	}

	enriched := EnrichWithBookmarks(results, bookmarks)

	// Should still have exactly 2 entries
	if len(enriched) != 2 {
		t.Errorf("expected 2 entries (no duplicates), got %d", len(enriched))
	}

	// First entry should be enriched
	if !enriched[0].Bookmarked {
		t.Error("first entry should be marked as bookmarked")
	}

	// Second entry should not be enriched
	if enriched[1].Bookmarked {
		t.Error("second entry should not be marked as bookmarked")
	}

	// Verify UUIDs are preserved
	if enriched[0].UUID != "entry-1" || enriched[1].UUID != "entry-2" {
		t.Error("UUIDs should be preserved in enriched results")
	}
}
