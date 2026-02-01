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
