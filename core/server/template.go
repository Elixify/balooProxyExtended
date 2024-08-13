package server

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
)

// templateCache is a global map that stores parsed templates
var templateCache = make(map[string]*template.Template)
var mu sync.Mutex // mutex to handle concurrent access to the cache

// RenderTemplate renders the specified HTML template with the provided data and returns it as a string
func RenderTemplate(templateFile string, data map[string]interface{}) (string, error) {
	mu.Lock()
	tmpl, found := templateCache[templateFile]
	mu.Unlock()

	if !found {
		// Parse the template file and store it in the cache
		var err error
		tmpl, err = template.ParseFiles(templateFile)
		if err != nil {
			return "", fmt.Errorf("failed to parse template: %w", err)
		}

		mu.Lock()
		templateCache[templateFile] = tmpl
		mu.Unlock()
	}

	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buffer.String(), nil
}
