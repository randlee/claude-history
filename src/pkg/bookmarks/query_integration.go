// Package bookmarks provides bookmark management for Claude Code agent sessions.
package bookmarks

import (
	"strings"

	"github.com/randlee/claude-history/pkg/models"
)

// EnrichWithBookmarks takes history query results and enriches them
// with bookmark metadata if the agent is bookmarked.
// Returns enriched results with no duplicates (single result per agent).
func EnrichWithBookmarks(results []models.ConversationEntry, bookmarks []Bookmark) []models.ConversationEntry {
	// Build agent_id → bookmark map
	bookmarkMap := make(map[string]*Bookmark)
	for i := range bookmarks {
		bookmarkMap[bookmarks[i].AgentID] = &bookmarks[i]
	}

	// Enrich results
	enriched := make([]models.ConversationEntry, len(results))
	for i := range results {
		enriched[i] = results[i]
		if bookmark, exists := bookmarkMap[results[i].AgentID]; exists {
			enriched[i].Bookmarked = true
			enriched[i].BookmarkID = bookmark.BookmarkID
			enriched[i].BookmarkName = bookmark.Name
			enriched[i].BookmarkTags = bookmark.Tags
		}
	}

	return enriched
}

// MatchBookmarkTags checks if a bookmark matches the given query string.
// The match is case-insensitive and checks against:
// - Bookmark name
// - Any tag in bookmark.Tags
// - Bookmark description
// Returns true if any of these fields contain the query string.
func MatchBookmarkTags(bookmark Bookmark, query string) bool {
	query = strings.ToLower(query)

	// Check bookmark name
	if strings.Contains(strings.ToLower(bookmark.Name), query) {
		return true
	}

	// Check tags
	for _, tag := range bookmark.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	// Check description
	if strings.Contains(strings.ToLower(bookmark.Description), query) {
		return true
	}

	return false
}
