package writers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/templates"
)

// FileWriter handles writing generated files
type FileWriter struct {
	outputDir string
	data      *config.TemplateData
}

// NewFileWriter creates a new file writer
func NewFileWriter(outputDir string, data *config.TemplateData) *FileWriter {
	return &FileWriter{
		outputDir: outputDir,
		data:      data,
	}
}

// WriteTemplate processes a template and writes it to a file
func (w *FileWriter) WriteTemplate(templatePath, outputPath string) error {
	// Read template from embedded FS
	content, err := templates.FS.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Create template functions
	funcMap := template.FuncMap{
		"title": config.Title,
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, w.data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	// Write to file
	return w.WriteFile(outputPath, buf.Bytes())
}

// WriteFile writes content to a file
func (w *FileWriter) WriteFile(relativePath string, content []byte) error {
	fullPath := filepath.Join(w.outputDir, relativePath)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}

// CopyStatic copies a static file from the embedded FS
func (w *FileWriter) CopyStatic(sourcePath, destPath string) error {
	content, err := templates.FS.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read static file %s: %w", sourcePath, err)
	}

	return w.WriteFile(destPath, content)
}

// EnsureDir creates a directory if it doesn't exist
func (w *FileWriter) EnsureDir(relativePath string) error {
	fullPath := filepath.Join(w.outputDir, relativePath)
	return os.MkdirAll(fullPath, 0755)
}

// OutputDir returns the output directory
func (w *FileWriter) OutputDir() string {
	return w.outputDir
}
