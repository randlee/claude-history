// Package session handles Claude Code session operations.
package session

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/randlee/claude-history/internal/jsonl"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
)

// ReadSession reads all entries from a session JSONL file.
func ReadSession(filePath string) ([]models.ConversationEntry, error) {
	return jsonl.ReadAll[models.ConversationEntry](filePath)
}

// ScanSession streams through a session JSONL file, calling fn for each entry.
// If fn returns StopScan, scanning stops early without error.
func ScanSession(filePath string, fn func(entry models.ConversationEntry) error) error {
	err := jsonl.ScanInto(filePath, func(entry models.ConversationEntry) error {
		if err := fn(entry); err != nil {
			if err == StopScan {
				return err // Propagate to stop iteration
			}
			return err
		}
		return nil
	})
	if err == StopScan {
		return nil // StopScan is not an error
	}
	return err
}

// GetSessionInfo extracts session metadata by scanning a session file.
func GetSessionInfo(filePath string) (*models.Session, error) {
	var session models.Session
	var firstEntry, lastEntry *models.ConversationEntry
	var messageCount int
	var firstPrompt string

	err := ScanSession(filePath, func(entry models.ConversationEntry) error {
		messageCount++

		if firstEntry == nil {
			entryCopy := entry
			firstEntry = &entryCopy
		}
		entryCopy := entry
		lastEntry = &entryCopy

		// Capture first user message as the prompt
		if firstPrompt == "" && entry.IsUser() {
			firstPrompt = entry.GetTextContent()
			if len(firstPrompt) > 200 {
				firstPrompt = firstPrompt[:200] + "..."
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if firstEntry != nil {
		session.ID = firstEntry.SessionID
		session.IsSidechain = firstEntry.IsSidechain

		if ts, err := firstEntry.GetTimestamp(); err == nil {
			session.Created = ts
		}
	}

	if lastEntry != nil {
		if ts, err := lastEntry.GetTimestamp(); err == nil {
			session.Modified = ts
		}
	}

	session.FilePath = filePath
	session.MessageCount = messageCount
	session.FirstPrompt = firstPrompt

	return &session, nil
}

// ListSessions returns all sessions in a project directory.
// It scans all JSONL files and enriches with index data when available.
// Empty sessions (no user/assistant messages) are filtered out.
func ListSessions(projectDir string) ([]models.Session, error) {
	// Build index lookup map for enrichment
	indexMap := make(map[string]*models.SessionIndexEntry)
	indexPath := filepath.Join(projectDir, "sessions-index.json")
	if paths.Exists(indexPath) {
		if index, err := ReadSessionIndex(indexPath); err == nil {
			for i := range index.Entries {
				indexMap[index.Entries[i].SessionID] = &index.Entries[i]
			}
		}
	}

	// Always scan directory for all session files
	sessionFiles, err := paths.ListSessionFiles(projectDir)
	if err != nil {
		return nil, err
	}

	var sessions []models.Session
	for sessionID, filePath := range sessionFiles {
		// Check if we have index data for this session
		if indexEntry, ok := indexMap[sessionID]; ok {
			// Use index data (faster, has summary)
			sessions = append(sessions, indexEntry.ToSession())
		} else {
			// Scan file to get session info
			info, err := GetSessionInfo(filePath)
			if err != nil {
				// Skip files that can't be read
				continue
			}
			info.ID = sessionID

			// Filter out empty sessions (no actual conversation)
			if !hasConversation(filePath) {
				continue
			}

			sessions = append(sessions, *info)
		}
	}

	// Sort by modified time (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Modified.After(sessions[j].Modified)
	})

	return sessions, nil
}

// hasConversation checks if a session file has at least one user or assistant message.
func hasConversation(filePath string) bool {
	hasContent := false
	_ = ScanSession(filePath, func(entry models.ConversationEntry) error {
		if entry.Type == models.EntryTypeUser || entry.Type == models.EntryTypeAssistant {
			hasContent = true
			return StopScan // Stop scanning once we find one
		}
		return nil
	})
	return hasContent
}

// StopScan is a sentinel error to stop scanning early.
var StopScan = &stopScanError{}

type stopScanError struct{}

func (e *stopScanError) Error() string { return "stop scan" }

// FindSession finds a session by ID in a project directory.
func FindSession(projectDir string, sessionID string) (*models.Session, error) {
	filePath := filepath.Join(projectDir, sessionID+".jsonl")
	if !paths.Exists(filePath) {
		return nil, os.ErrNotExist
	}

	return GetSessionInfo(filePath)
}

// FilterOptions specifies criteria for filtering session entries.
type FilterOptions struct {
	StartTime *time.Time
	EndTime   *time.Time
	Types     []models.EntryType
	AgentID   string

	// Tool filtering
	ToolTypes []string // Filter by tool names (case-insensitive)
	ToolMatch string   // Regex pattern to match tool inputs
}

// FilterEntries filters session entries based on the given options.
func FilterEntries(entries []models.ConversationEntry, opts FilterOptions) []models.ConversationEntry {
	var result []models.ConversationEntry

	typeSet := make(map[models.EntryType]bool)
	for _, t := range opts.Types {
		typeSet[t] = true
	}

	for _, entry := range entries {
		// Filter by type
		if len(typeSet) > 0 && !typeSet[entry.Type] {
			continue
		}

		// Filter by agent ID
		if opts.AgentID != "" && entry.AgentID != opts.AgentID {
			continue
		}

		// Filter by time range
		if opts.StartTime != nil || opts.EndTime != nil {
			ts, err := entry.GetTimestamp()
			if err != nil {
				continue
			}
			if opts.StartTime != nil && ts.Before(*opts.StartTime) {
				continue
			}
			if opts.EndTime != nil && ts.After(*opts.EndTime) {
				continue
			}
		}

		// Filter by tool types (only applies to entries with tool calls)
		if len(opts.ToolTypes) > 0 {
			hasMatchingTool := false
			for _, toolName := range opts.ToolTypes {
				if entry.HasToolCall(toolName) {
					hasMatchingTool = true
					break
				}
			}
			if !hasMatchingTool {
				continue
			}
		}

		// Filter by tool input pattern
		if opts.ToolMatch != "" {
			if !entry.MatchesToolInput(opts.ToolMatch) {
				continue
			}
		}

		result = append(result, entry)
	}

	return result
}

// CountEntriesByType counts entries grouped by type.
func CountEntriesByType(entries []models.ConversationEntry) map[models.EntryType]int {
	counts := make(map[models.EntryType]int)
	for _, entry := range entries {
		counts[entry.Type]++
	}
	return counts
}
