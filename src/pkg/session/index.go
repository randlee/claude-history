package session

import (
	"encoding/json"
	"os"

	"github.com/randlee/claude-history/pkg/models"
)

// ReadSessionIndex reads and parses a sessions-index.json file.
func ReadSessionIndex(filePath string) (*models.SessionIndex, error) {
	data, err := os.ReadFile(filePath) //nolint:gosec // G304: file path from CLI input is expected
	if err != nil {
		return nil, err
	}

	var index models.SessionIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// GetSessionFromIndex finds a session by ID in an index.
func GetSessionFromIndex(index *models.SessionIndex, sessionID string) *models.SessionIndexEntry {
	for i := range index.Entries {
		if index.Entries[i].SessionID == sessionID {
			return &index.Entries[i]
		}
	}
	return nil
}

// GetProjectPathFromIndex extracts the project path from a session index.
// Returns the projectPath from the first entry, or empty string if no entries.
func GetProjectPathFromIndex(index *models.SessionIndex) string {
	if len(index.Entries) == 0 {
		return ""
	}
	return index.Entries[0].ProjectPath
}
