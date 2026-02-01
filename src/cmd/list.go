package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/internal/output"
	"github.com/randlee/claude-history/pkg/encoding"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/session"
)

var (
	listProjectID string
)

var listCmd = &cobra.Command{
	Use:   "list [project-path]",
	Short: "List projects or sessions",
	Long: `List all Claude Code projects, or sessions within a specific project.

Examples:
  # List all projects
  claude-history list

  # List sessions in a project (by filesystem path)
  claude-history list /Users/randlee/Documents/github/project

  # List sessions in a project (by encoded ID)
  claude-history list --project-id -Users-randlee-Documents-github`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listProjectID, "project-id", "", "Encoded project ID (alternative to path)")
}

func runList(cmd *cobra.Command, args []string) error {
	outputFormat := output.ParseFormat(format)

	// Determine if we're listing projects or sessions
	var projectPath string
	var projectDir string

	if len(args) > 0 {
		projectPath = args[0]
	}

	// Handle --project-id flag
	if listProjectID != "" {
		projectsDir, err := paths.ProjectsDir(claudeDir)
		if err != nil {
			return err
		}
		projectDir = filepath.Join(projectsDir, listProjectID)
		if !paths.Exists(projectDir) {
			return fmt.Errorf("project not found: %s", listProjectID)
		}
	} else if projectPath != "" {
		// Convert filesystem path to project directory
		var err error
		projectDir, err = paths.ProjectDir(claudeDir, projectPath)
		if err != nil {
			return err
		}
	}

	// If we have a project, list sessions
	if projectDir != "" {
		return listSessions(projectDir, outputFormat)
	}

	// Otherwise, list all projects
	return listProjects(outputFormat)
}

func listProjects(format output.Format) error {
	projectsMap, err := paths.ListProjects(claudeDir)
	if err != nil {
		return err
	}

	if len(projectsMap) == 0 {
		fmt.Fprintln(os.Stderr, "No projects found")
		return nil
	}

	var projects []models.Project
	for name, path := range projectsMap {
		project := models.Project{
			Name: name,
			Path: path,
		}

		// Try to get the original project path from sessions-index.json
		indexPath := filepath.Join(path, "sessions-index.json")
		if paths.Exists(indexPath) {
			if index, err := session.ReadSessionIndex(indexPath); err == nil {
				project.ProjectPath = session.GetProjectPathFromIndex(index)
			}
		}

		// Fallback: decode the path
		if project.ProjectPath == "" {
			project.ProjectPath = encoding.DecodePath(name, "")
		}

		projects = append(projects, project)
	}

	return output.WriteProjects(os.Stdout, projects, format)
}

func listSessions(projectDir string, format output.Format) error {
	if !paths.Exists(projectDir) {
		return fmt.Errorf("project directory not found: %s", projectDir)
	}

	sessions, err := session.ListSessions(projectDir)
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Fprintln(os.Stderr, "No sessions found")
		return nil
	}

	return output.WriteSessions(os.Stdout, sessions, format)
}
