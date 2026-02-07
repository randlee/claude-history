package export

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed templates/*
var templatesFS embed.FS

// GetStyleCSS returns the contents of the embedded CSS file.
func GetStyleCSS() string {
	data, err := templatesFS.ReadFile("templates/style.css")
	if err != nil {
		return ""
	}
	return string(data)
}

// GetScriptJS returns the contents of the embedded JavaScript file.
func GetScriptJS() string {
	data, err := templatesFS.ReadFile("templates/script.js")
	if err != nil {
		return ""
	}
	return string(data)
}

// GetClipboardJS returns the contents of the embedded clipboard JavaScript file.
func GetClipboardJS() string {
	data, err := templatesFS.ReadFile("templates/clipboard.js")
	if err != nil {
		return ""
	}
	return string(data)
}

// GetControlsJS returns the contents of the embedded controls JavaScript file.
func GetControlsJS() string {
	data, err := templatesFS.ReadFile("templates/controls.js")
	if err != nil {
		return ""
	}
	return string(data)
}

// GetNavigationJS returns the contents of the embedded navigation JavaScript file.
func GetNavigationJS() string {
	data, err := templatesFS.ReadFile("templates/navigation.js")
	if err != nil {
		return ""
	}
	return string(data)
}

// GetAgentTooltipJS returns the contents of the embedded agent-tooltip JavaScript file.
func GetAgentTooltipJS() string {
	data, err := templatesFS.ReadFile("templates/agent-tooltip.js")
	if err != nil {
		return ""
	}
	return string(data)
}

// WriteStaticAssets writes all static assets to the output directory.
// Creates a 'static' subdirectory containing style.css and script.js.
func WriteStaticAssets(outputDir string) error {
	staticDir := filepath.Join(outputDir, "static")

	// Create static directory
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return err
	}

	// Write CSS file
	cssContent := GetStyleCSS()
	if cssContent != "" {
		cssPath := filepath.Join(staticDir, "style.css")
		if err := os.WriteFile(cssPath, []byte(cssContent), 0644); err != nil {
			return err
		}
	}

	// Write JavaScript file
	jsContent := GetScriptJS()
	if jsContent != "" {
		jsPath := filepath.Join(staticDir, "script.js")
		if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
			return err
		}
	}

	// Write clipboard JavaScript file
	clipboardContent := GetClipboardJS()
	if clipboardContent != "" {
		clipboardPath := filepath.Join(staticDir, "clipboard.js")
		if err := os.WriteFile(clipboardPath, []byte(clipboardContent), 0644); err != nil {
			return err
		}
	}

	// Write controls JavaScript file
	controlsContent := GetControlsJS()
	if controlsContent != "" {
		controlsPath := filepath.Join(staticDir, "controls.js")
		if err := os.WriteFile(controlsPath, []byte(controlsContent), 0644); err != nil {
			return err
		}
	}

	// Write navigation JavaScript file
	navigationContent := GetNavigationJS()
	if navigationContent != "" {
		navigationPath := filepath.Join(staticDir, "navigation.js")
		if err := os.WriteFile(navigationPath, []byte(navigationContent), 0644); err != nil {
			return err
		}
	}

	// Write agent-tooltip JavaScript file
	agentTooltipContent := GetAgentTooltipJS()
	if agentTooltipContent != "" {
		agentTooltipPath := filepath.Join(staticDir, "agent-tooltip.js")
		if err := os.WriteFile(agentTooltipPath, []byte(agentTooltipContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

// GetTemplatesFS returns the embedded filesystem containing templates.
// This allows custom template processing if needed.
func GetTemplatesFS() fs.FS {
	subFS, _ := fs.Sub(templatesFS, "templates")
	return subFS
}

// ListTemplateFiles returns a list of all embedded template files.
func ListTemplateFiles() ([]string, error) {
	var files []string

	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ReadTemplateFile reads a specific template file by name.
func ReadTemplateFile(name string) ([]byte, error) {
	return templatesFS.ReadFile("templates/" + name)
}
