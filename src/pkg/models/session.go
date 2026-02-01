package models

import (
	"time"
)

// Session represents a Claude Code session.
type Session struct {
	ID           string    `json:"sessionId"`
	ProjectPath  string    `json:"projectPath"`
	FilePath     string    `json:"fullPath"`
	FirstPrompt  string    `json:"firstPrompt,omitempty"`
	Summary      string    `json:"summary,omitempty"`
	MessageCount int       `json:"messageCount"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
	GitBranch    string    `json:"gitBranch,omitempty"`
	IsSidechain  bool      `json:"isSidechain"`
}

// SessionIndexEntry represents an entry in sessions-index.json.
type SessionIndexEntry struct {
	SessionID    string `json:"sessionId"`
	FullPath     string `json:"fullPath"`
	ProjectPath  string `json:"projectPath"`
	FirstPrompt  string `json:"firstPrompt,omitempty"`
	Summary      string `json:"summary,omitempty"`
	MessageCount int    `json:"messageCount"`
	Created      string `json:"created"`
	Modified     string `json:"modified"`
	GitBranch    string `json:"gitBranch,omitempty"`
	IsSidechain  bool   `json:"isSidechain"`
}

// ToSession converts a SessionIndexEntry to a Session.
func (e *SessionIndexEntry) ToSession() Session {
	created, _ := time.Parse(time.RFC3339Nano, e.Created)
	modified, _ := time.Parse(time.RFC3339Nano, e.Modified)

	return Session{
		ID:           e.SessionID,
		ProjectPath:  e.ProjectPath,
		FilePath:     e.FullPath,
		FirstPrompt:  e.FirstPrompt,
		Summary:      e.Summary,
		MessageCount: e.MessageCount,
		Created:      created,
		Modified:     modified,
		GitBranch:    e.GitBranch,
		IsSidechain:  e.IsSidechain,
	}
}

// SessionIndex represents the sessions-index.json file structure.
type SessionIndex struct {
	Version int                 `json:"version"`
	Entries []SessionIndexEntry `json:"entries"`
}

// Agent represents a spawned agent within a session.
type Agent struct {
	ID         string  `json:"agentId"`
	SessionID  string  `json:"sessionId"`
	FilePath   string  `json:"filePath"`
	EntryCount int     `json:"entryCount"`
	SpawnedBy  *string `json:"spawnedBy,omitempty"` // parentUuid of spawning queue-operation
	AgentType  string  `json:"agentType,omitempty"` // e.g., "prompt_suggestion", "explore"
}

// Project represents a Claude Code project directory.
type Project struct {
	Name        string `json:"name"`        // Encoded name (directory name)
	Path        string `json:"path"`        // Full path to project directory
	ProjectPath string `json:"projectPath"` // Original filesystem path (decoded or from index)
}
