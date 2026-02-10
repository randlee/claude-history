package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/pkg/bookmarks"
)

var (
	// Bookmark add flags
	bookmarkName        string
	bookmarkAgentID     string
	bookmarkSessionID   string
	bookmarkProjectPath string
	bookmarkTags        []string
	bookmarkDescription string

	// Bookmark list flags
	bookmarkFilterTag string

	// Bookmark update flags
	bookmarkAddTags []string

	// Bookmark delete flags
	bookmarkForce bool
)

// bookmarkCmd represents the bookmark command
var bookmarkCmd = &cobra.Command{
	Use:   "bookmark",
	Short: "Manage bookmarks for Claude Code agents",
	Long: `Manage bookmarks for Claude Code agents.

Bookmarks allow you to save references to specific agents from conversation
history for later retrieval and resurrection.`,
}

// bookmarkAddCmd represents the bookmark add command
var bookmarkAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new bookmark",
	Long: `Add a new bookmark for a Claude Code agent.

Example:
  claude-history bookmark add \
    --name beads-expert \
    --agent agent-abc123 \
    --session session-xyz \
    --project /path/to/beads \
    --tags "architecture,beads" \
    --description "Explored beads architecture"`,
	RunE: runBookmarkAdd,
}

// bookmarkListCmd represents the bookmark list command
var bookmarkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bookmarks",
	Long: `List all bookmarks, optionally filtered by tag.

Example:
  claude-history bookmark list
  claude-history bookmark list --tag architecture
  claude-history bookmark list --format json`,
	RunE: runBookmarkList,
}

// bookmarkGetCmd represents the bookmark get command
var bookmarkGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a specific bookmark",
	Long: `Get detailed information about a specific bookmark.

Example:
  claude-history bookmark get beads-expert
  claude-history bookmark get beads-expert --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runBookmarkGet,
}

// bookmarkUpdateCmd represents the bookmark update command
var bookmarkUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a bookmark",
	Long: `Update a bookmark's description or tags.

Example:
  claude-history bookmark update beads-expert \
    --description "Updated description" \
    --add-tags "python,advanced"`,
	Args: cobra.ExactArgs(1),
	RunE: runBookmarkUpdate,
}

// bookmarkDeleteCmd represents the bookmark delete command
var bookmarkDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a bookmark",
	Long: `Delete a bookmark by name.

Example:
  claude-history bookmark delete beads-expert
  claude-history bookmark delete beads-expert --force`,
	Args: cobra.ExactArgs(1),
	RunE: runBookmarkDelete,
}

// bookmarkSearchCmd represents the bookmark search command
var bookmarkSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search bookmarks",
	Long: `Search bookmarks by name, description, or tags.

Example:
  claude-history bookmark search beads
  claude-history bookmark search architecture`,
	Args: cobra.ExactArgs(1),
	RunE: runBookmarkSearch,
}

func init() {
	rootCmd.AddCommand(bookmarkCmd)

	// Add subcommands
	bookmarkCmd.AddCommand(bookmarkAddCmd)
	bookmarkCmd.AddCommand(bookmarkListCmd)
	bookmarkCmd.AddCommand(bookmarkGetCmd)
	bookmarkCmd.AddCommand(bookmarkUpdateCmd)
	bookmarkCmd.AddCommand(bookmarkDeleteCmd)
	bookmarkCmd.AddCommand(bookmarkSearchCmd)

	// bookmark add flags
	bookmarkAddCmd.Flags().StringVar(&bookmarkName, "name", "", "Bookmark name (required)")
	bookmarkAddCmd.Flags().StringVar(&bookmarkAgentID, "agent", "", "Agent ID (required)")
	bookmarkAddCmd.Flags().StringVar(&bookmarkSessionID, "session", "", "Session ID (required)")
	bookmarkAddCmd.Flags().StringVar(&bookmarkProjectPath, "project", "", "Project path")
	bookmarkAddCmd.Flags().StringSliceVar(&bookmarkTags, "tags", []string{}, "Comma-separated tags")
	bookmarkAddCmd.Flags().StringVar(&bookmarkDescription, "description", "", "Bookmark description")
	_ = bookmarkAddCmd.MarkFlagRequired("name")
	_ = bookmarkAddCmd.MarkFlagRequired("agent")
	_ = bookmarkAddCmd.MarkFlagRequired("session")

	// bookmark list flags
	bookmarkListCmd.Flags().StringVar(&bookmarkFilterTag, "tag", "", "Filter by tag")

	// bookmark update flags
	bookmarkUpdateCmd.Flags().StringVar(&bookmarkDescription, "description", "", "New description")
	bookmarkUpdateCmd.Flags().StringSliceVar(&bookmarkAddTags, "add-tags", []string{}, "Tags to add")

	// bookmark delete flags
	bookmarkDeleteCmd.Flags().BoolVar(&bookmarkForce, "force", false, "Skip confirmation prompt")
}

// getStorage returns a storage instance for the bookmarks file
func getStorage() (bookmarks.Storage, error) {
	// Determine claude directory
	dir := claudeDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(home, ".claude")
	}

	bookmarksFile := filepath.Join(dir, "bookmarks.jsonl")
	return bookmarks.NewJSONLStorage(bookmarksFile)
}

// runBookmarkAdd handles the bookmark add command
func runBookmarkAdd(cmd *cobra.Command, args []string) error {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// Create bookmark
	bookmark := bookmarks.Bookmark{
		Name:        bookmarkName,
		Description: bookmarkDescription,
		AgentID:     bookmarkAgentID,
		SessionID:   bookmarkSessionID,
		ProjectPath: bookmarkProjectPath,
		Hostname:    hostname,
		Scope:       "global",
		Tags:        bookmarkTags,
	}

	// Validate bookmark
	if err := bookmarks.ValidateBookmark(bookmark); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate agent exists if project path provided
	if bookmarkProjectPath != "" {
		if err := bookmarks.ValidateAgentExists(bookmarkAgentID, bookmarkSessionID, bookmarkProjectPath); err != nil {
			return fmt.Errorf("agent validation failed: %w", err)
		}
	}

	// Get storage
	storage, err := getStorage()
	if err != nil {
		return err
	}

	// Add bookmark
	if err := storage.Add(bookmark); err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}

	// Retrieve the saved bookmark to get the generated ID
	saved, err := storage.Get(bookmarkName)
	if err != nil {
		return fmt.Errorf("failed to retrieve saved bookmark: %w", err)
	}

	fmt.Printf("Bookmark %q created with ID %s\n", bookmarkName, saved.BookmarkID)
	return nil
}

// runBookmarkList handles the bookmark list command
func runBookmarkList(cmd *cobra.Command, args []string) error {
	storage, err := getStorage()
	if err != nil {
		return err
	}

	allBookmarks, err := storage.List()
	if err != nil {
		return fmt.Errorf("failed to list bookmarks: %w", err)
	}

	// Filter by tag if specified
	if bookmarkFilterTag != "" {
		filtered := make([]bookmarks.Bookmark, 0)
		for _, b := range allBookmarks {
			for _, tag := range b.Tags {
				if tag == bookmarkFilterTag {
					filtered = append(filtered, b)
					break
				}
			}
		}
		allBookmarks = filtered
	}

	if len(allBookmarks) == 0 {
		fmt.Println("No bookmarks found")
		return nil
	}

	// Output based on format
	if format == "json" {
		data, err := json.MarshalIndent(allBookmarks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Text format (table)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tID\tSESSION\tTAGS\tDESCRIPTION")
	for _, b := range allBookmarks {
		tags := strings.Join(b.Tags, ",")
		desc := b.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			b.Name,
			b.BookmarkID,
			b.SessionID,
			tags,
			desc)
	}
	_ = w.Flush()

	return nil
}

// runBookmarkGet handles the bookmark get command
func runBookmarkGet(cmd *cobra.Command, args []string) error {
	name := args[0]

	storage, err := getStorage()
	if err != nil {
		return err
	}

	bookmark, err := storage.Get(name)
	if err != nil {
		return fmt.Errorf("failed to get bookmark: %w", err)
	}

	if bookmark == nil {
		return fmt.Errorf("bookmark %q not found", name)
	}

	// Output based on format
	if format == "json" {
		data, err := json.MarshalIndent(bookmark, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Text format
	fmt.Printf("Name:              %s\n", bookmark.Name)
	fmt.Printf("ID:                %s\n", bookmark.BookmarkID)
	fmt.Printf("Description:       %s\n", bookmark.Description)
	fmt.Printf("Agent ID:          %s\n", bookmark.AgentID)
	fmt.Printf("Session ID:        %s\n", bookmark.SessionID)
	fmt.Printf("Project Path:      %s\n", bookmark.ProjectPath)
	fmt.Printf("Hostname:          %s\n", bookmark.Hostname)
	fmt.Printf("Bookmarked At:     %s\n", bookmark.BookmarkedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Bookmarked By:     %s\n", bookmark.BookmarkedBy)
	fmt.Printf("Scope:             %s\n", bookmark.Scope)
	fmt.Printf("Tags:              %s\n", strings.Join(bookmark.Tags, ", "))
	fmt.Printf("Resurrection Count: %d\n", bookmark.ResurrectionCount)
	if bookmark.LastResurrected != nil {
		fmt.Printf("Last Resurrected:  %s\n", bookmark.LastResurrected.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// runBookmarkUpdate handles the bookmark update command
func runBookmarkUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]

	storage, err := getStorage()
	if err != nil {
		return err
	}

	// Check bookmark exists
	existing, err := storage.Get(name)
	if err != nil {
		return fmt.Errorf("failed to get bookmark: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("bookmark %q not found", name)
	}

	// Build updates map
	updates := make(map[string]interface{})

	// Update description if provided
	if cmd.Flags().Changed("description") {
		updates["description"] = bookmarkDescription
	}

	// Add tags if provided
	if cmd.Flags().Changed("add-tags") {
		// Merge with existing tags
		tagSet := make(map[string]bool)
		for _, tag := range existing.Tags {
			tagSet[tag] = true
		}
		for _, tag := range bookmarkAddTags {
			tagSet[tag] = true
		}
		tags := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			tags = append(tags, tag)
		}
		updates["tags"] = tags
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates specified")
	}

	// Apply updates
	if err := storage.Update(name, updates); err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}

	fmt.Printf("Bookmark %q updated\n", name)
	return nil
}

// runBookmarkDelete handles the bookmark delete command
func runBookmarkDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	storage, err := getStorage()
	if err != nil {
		return err
	}

	// Check bookmark exists
	existing, err := storage.Get(name)
	if err != nil {
		return fmt.Errorf("failed to get bookmark: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("bookmark %q not found", name)
	}

	// Confirm deletion unless --force is set
	if !bookmarkForce {
		fmt.Printf("Delete bookmark %q (ID: %s)? [y/N]: ", name, existing.BookmarkID)
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion canceled")
			return nil
		}
	}

	// Delete bookmark
	if err := storage.Delete(name); err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	fmt.Printf("Bookmark %q deleted\n", name)
	return nil
}

// runBookmarkSearch handles the bookmark search command
func runBookmarkSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	storage, err := getStorage()
	if err != nil {
		return err
	}

	results, err := storage.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search bookmarks: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No bookmarks found matching %q\n", query)
		return nil
	}

	// Output based on format
	if format == "json" {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Text format (table)
	fmt.Printf("Found %d bookmark(s) matching %q:\n\n", len(results), query)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tID\tSESSION\tTAGS\tDESCRIPTION")
	for _, b := range results {
		tags := strings.Join(b.Tags, ",")
		desc := b.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			b.Name,
			b.BookmarkID,
			b.SessionID,
			tags,
			desc)
	}
	_ = w.Flush()

	return nil
}
