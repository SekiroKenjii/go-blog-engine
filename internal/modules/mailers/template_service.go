package mailers

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
)

// TemplateService handles email template rendering
type TemplateService struct {
	templateDir string
}

// NewTemplateService creates a new template service
func NewTemplateService(templateDir string) *TemplateService {
	return &TemplateService{
		templateDir: templateDir,
	}
}

// RenderTemplate renders an email template with the given data
func (ts *TemplateService) RenderTemplate(templateName string, data any) (string, error) {
	templatePath := filepath.Join(ts.templateDir, templateName)

	// First, try to parse the template file
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Execute the template with data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// RenderTemplateWithFallback renders a template with a fallback option
func (ts *TemplateService) RenderTemplateWithFallback(templateName string, data any, fallbackTemplate string) (string, error) {
	// Try to render the primary template
	body, err := ts.RenderTemplate(templateName, data)
	if err != nil {
		logger.Warn(fmt.Sprintf("Primary template %s failed, using fallback: %v", templateName, err))

		// Use fallback template
		tmpl, parseErr := template.New("fallback").Parse(fallbackTemplate)
		if parseErr != nil {
			return "", fmt.Errorf("both primary template and fallback failed: primary=%w, fallback=%v", err, parseErr)
		}

		var buf bytes.Buffer
		if execErr := tmpl.Execute(&buf, data); execErr != nil {
			return "", fmt.Errorf("both primary template and fallback execution failed: primary=%w, fallback=%v", err, execErr)
		}

		return buf.String(), nil
	}

	return body, nil
}
