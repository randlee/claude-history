// Package bookmarks provides bookmark management for Claude Code agent sessions.
package bookmarks

import (
	"os"
	"path/filepath"
)

// DefaultBookmarkPath returns the default path to the bookmarks file.
// Returns ~/.claude/bookmarks.jsonl
func DefaultBookmarkPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "bookmarks.jsonl"), nil
}

// LoadDefaultBookmarks loads bookmarks from the default location.
// Returns an empty slice if the file doesn't exist.
func LoadDefaultBookmarks() ([]Bookmark, error) {
	path, err := DefaultBookmarkPath()
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Bookmark{}, nil
	}

	// Use existing JSONLStorage to read bookmarks
	storage, err := NewJSONLStorage(path)
	if err != nil {
		return nil, err
	}

	return storage.List()
}
