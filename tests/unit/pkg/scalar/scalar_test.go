package scalar

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/pkg/scalar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	htmlDoctype       = "<!DOCTYPE html>"
	defaultPageTitle  = "Scalar API Reference"
	basicThemeComment = "/* basic theme */"
)

func TestApiReferenceHTML(t *testing.T) {
	t.Run("basic spec content string", func(t *testing.T) {
		options := &scalar.Options{
			SpecContent: `{"openapi": "3.0.0", "info": {"title": "Test API", "version": "1.0.0"}}`,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, htmlDoctype)
		assert.Contains(t, html, defaultPageTitle)
		assert.Contains(t, html, `"openapi": "3.0.0"`)
	})

	t.Run("spec content as map", func(t *testing.T) {
		specMap := map[string]any{
			"openapi": "3.0.0",
			"info": map[string]any{
				"title":   "Test API",
				"version": "1.0.0",
			},
		}

		options := &scalar.Options{
			SpecContent: specMap,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, htmlDoctype)
		assert.Contains(t, html, `"openapi":"3.0.0"`)
		assert.Contains(t, html, `"title":"Test API"`)
	})

	t.Run("spec content as function", func(t *testing.T) {
		specFunc := func() map[string]any {
			return map[string]any{
				"openapi": "3.0.0",
				"info": map[string]any{
					"title":   "Dynamic API",
					"version": "2.0.0",
				},
			}
		}

		options := &scalar.Options{
			SpecContent: specFunc,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, htmlDoctype)
		assert.Contains(t, html, `"openapi":"3.0.0"`)
		assert.Contains(t, html, `"title":"Dynamic API"`)
	})

	t.Run("custom page title", func(t *testing.T) {
		options := &scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "My Custom API Docs",
			},
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, "<title>My Custom API Docs</title>")
	})

	t.Run("with custom theme", func(t *testing.T) {
		options := &scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
			Theme:       "dark",
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.NotContains(t, html, scalar.CustomThemeCSS)
	})

	t.Run("without custom theme uses default CSS", func(t *testing.T) {
		options := &scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, basicThemeComment)
	})

	t.Run("uses default CDN", func(t *testing.T) {
		options := &scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, scalar.DefaultCDN)
	})

	t.Run("uses custom CDN", func(t *testing.T) {
		customCDN := "https://custom-cdn.com/scalar"
		options := &scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
			CDN:         customCDN,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, customCDN)
		assert.NotContains(t, html, scalar.DefaultCDN)
	})
}

func TestApiReferenceHTMLWithSpecURL(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"openapi": "3.0.0", "info": {"title": "Remote API", "version": "1.0.0"}}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("fetch from HTTP URL", func(t *testing.T) {
		options := &scalar.Options{
			SpecURL: server.URL + "/openapi.json",
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, `"title": "Remote API"`)
	})

	t.Run("fetch from invalid HTTP URL", func(t *testing.T) {
		options := &scalar.Options{
			SpecURL: "http://localhost:99999/nonexistent.json", // Use an invalid port instead
		}

		_, err := scalar.ApiReferenceHTML(options)
		assert.Error(t, err)
	})

	t.Run("fetch from file URL", func(t *testing.T) {
		// Create a temporary file
		tempDir := t.TempDir()
		specFile := filepath.Join(tempDir, "openapi.json")
		specContent := `{"openapi": "3.0.0", "info": {"title": "File API", "version": "1.0.0"}}`
		require.NoError(t, os.WriteFile(specFile, []byte(specContent), 0o644))

		options := &scalar.Options{
			SpecURL: specFile,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, `"title": "File API"`)
	})

	t.Run("fetch from non-existent file", func(t *testing.T) {
		options := &scalar.Options{
			SpecURL: "/non/existent/file.json",
		}

		_, err := scalar.ApiReferenceHTML(options)
		assert.Error(t, err)
	})
}

func TestApiReferenceHTMLErrorCases(t *testing.T) {
	t.Run("neither specURL nor specContent provided", func(t *testing.T) {
		options := &scalar.Options{}

		_, err := scalar.ApiReferenceHTML(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "specURL or specContent must be provided")
	})

	t.Run("invalid HTTP URL", func(t *testing.T) {
		options := &scalar.Options{
			SpecURL: "http://invalid-domain-that-does-not-exist.com/spec.json",
		}

		_, err := scalar.ApiReferenceHTML(options)
		assert.Error(t, err)
	})
}

func TestSafeJSONConfiguration(t *testing.T) {
	// Since safeJSONConfiguration is not exported, we test it indirectly
	t.Run("JSON escaping in configuration", func(t *testing.T) {
		options := &scalar.Options{
			SpecContent: `{"description": "API with \"quotes\""}`,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Test \"API\" Documentation",
			},
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)

		// Check that quotes are properly escaped in the data-configuration attribute
		assert.Contains(t, html, `data-configuration=`)
		// The quotes in the configuration should be escaped as &quot;
		assert.Contains(t, html, `&quot;`)
	})
}

func TestSpecContentHandler(t *testing.T) {
	// Since specContentHandler is not exported, we test it indirectly through ApiReferenceHTML

	t.Run("handles different spec content types", func(t *testing.T) {
		testCases := []struct {
			name        string
			specContent any
			expectError bool
			expectHTML  bool
		}{
			{
				name:        "string content",
				specContent: `{"openapi": "3.0.0"}`,
				expectError: false,
				expectHTML:  true,
			},
			{
				name: "map content",
				specContent: map[string]any{
					"openapi": "3.0.0",
					"info":    map[string]any{"title": "Map API"},
				},
				expectError: false,
				expectHTML:  true,
			},
			{
				name: "function content",
				specContent: func() map[string]any {
					return map[string]any{
						"openapi": "3.0.0",
						"info":    map[string]any{"title": "Function API"},
					}
				},
				expectError: false,
				expectHTML:  true,
			},
			{
				name:        "unsupported content type",
				specContent: 123,
				expectError: false,
				expectHTML:  true, // Should still generate HTML but with empty spec content
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := &scalar.Options{
					SpecContent: tc.specContent,
				}

				html, err := scalar.ApiReferenceHTML(options)

				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					if tc.expectHTML {
						assert.Contains(t, html, htmlDoctype)
					}
				}
			})
		}
	})
}

func TestDefaultOptions(t *testing.T) {
	t.Run("applies default values", func(t *testing.T) {
		input := scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
			// Leave other fields empty to test defaults
		}

		html, err := scalar.ApiReferenceHTML(&input)
		assert.NoError(t, err)

		// Should use default CDN
		assert.Contains(t, html, scalar.DefaultCDN)

		// Should use default page title
		assert.Contains(t, html, defaultPageTitle)

		// Should include custom theme CSS
		assert.Contains(t, html, basicThemeComment)
	})

	t.Run("preserves provided values", func(t *testing.T) {
		customCDN := "https://custom.cdn.com/scalar"
		input := scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
			CDN:         customCDN,
			Theme:       "dark",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Custom Title",
			},
		}

		html, err := scalar.ApiReferenceHTML(&input)
		assert.NoError(t, err)

		// Should use custom CDN
		assert.Contains(t, html, customCDN)
		assert.NotContains(t, html, scalar.DefaultCDN)

		// Should use custom page title
		assert.Contains(t, html, "Custom Title")
		assert.NotContains(t, html, defaultPageTitle)

		// Should not include custom theme CSS when theme is set
		assert.NotContains(t, html, basicThemeComment)
	})
}

func TestComplexSpecContent(t *testing.T) {
	t.Run("complex OpenAPI specification", func(t *testing.T) {
		complexSpec := map[string]any{
			"openapi": "3.0.0",
			"info": map[string]any{
				"title":       "Complex API",
				"version":     "1.0.0",
				"description": "A complex API with multiple endpoints",
			},
			"servers": []map[string]any{
				{"url": "https://api.example.com/v1"},
			},
			"paths": map[string]any{
				"/users": map[string]any{
					"get": map[string]any{
						"summary":     "Get users",
						"description": "Retrieve a list of users",
						"responses": map[string]any{
							"200": map[string]any{
								"description": "Success",
							},
						},
					},
				},
			},
		}

		options := &scalar.Options{
			SpecContent: complexSpec,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)
		assert.Contains(t, html, htmlDoctype)

		// Verify that the complex spec content is properly embedded
		assert.Contains(t, html, `"title":"Complex API"`)
		assert.Contains(t, html, `"paths"`)
		assert.Contains(t, html, `"/users"`)
	})
}

func TestHTMLStructure(t *testing.T) {
	t.Run("generates valid HTML structure", func(t *testing.T) {
		options := &scalar.Options{
			SpecContent: `{"openapi": "3.0.0"}`,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)

		// Check HTML structure
		assert.Contains(t, html, htmlDoctype)
		assert.Contains(t, html, "<html>")
		assert.Contains(t, html, "<head>")
		assert.Contains(t, html, "<title>")
		assert.Contains(t, html, "<meta charset=\"utf-8\"")
		assert.Contains(t, html, "<meta name=\"viewport\"")
		assert.Contains(t, html, "<style>")
		assert.Contains(t, html, "</head>")
		assert.Contains(t, html, "<body>")
		assert.Contains(t, html, `<script id="api-reference"`)
		assert.Contains(t, html, `type="application/json"`)
		assert.Contains(t, html, `data-configuration=`)
		assert.Contains(t, html, `<script src=`)
		assert.Contains(t, html, "</body>")
		assert.Contains(t, html, "</html>")
	})
}

func TestJSONSerialization(t *testing.T) {
	t.Run("properly serializes various data types", func(t *testing.T) {
		specContent := map[string]any{
			"string_field":  "test",
			"number_field":  42,
			"boolean_field": true,
			"array_field":   []string{"item1", "item2"},
			"object_field": map[string]any{
				"nested_key": "nested_value",
			},
			"null_field": nil,
		}

		options := &scalar.Options{
			SpecContent: specContent,
		}

		html, err := scalar.ApiReferenceHTML(options)
		assert.NoError(t, err)

		// Verify JSON serialization
		assert.Contains(t, html, `"string_field":"test"`)
		assert.Contains(t, html, `"number_field":42`)
		assert.Contains(t, html, `"boolean_field":true`)
		assert.Contains(t, html, `"array_field":["item1","item2"]`)
		assert.Contains(t, html, `"nested_key":"nested_value"`)
		assert.Contains(t, html, `"null_field":null`)
	})
}

// Benchmark tests
func BenchmarkApiReferenceHTMLString(b *testing.B) {
	options := &scalar.Options{
		SpecContent: `{"openapi": "3.0.0", "info": {"title": "Benchmark API", "version": "1.0.0"}}`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scalar.ApiReferenceHTML(options)
	}
}

func BenchmarkApiReferenceHTMLMap(b *testing.B) {
	specMap := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":   "Benchmark API",
			"version": "1.0.0",
		},
	}

	options := &scalar.Options{
		SpecContent: specMap,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scalar.ApiReferenceHTML(options)
	}
}

func BenchmarkApiReferenceHTMLFunction(b *testing.B) {
	specFunc := func() map[string]any {
		return map[string]any{
			"openapi": "3.0.0",
			"info": map[string]any{
				"title":   "Benchmark API",
				"version": "1.0.0",
			},
		}
	}

	options := &scalar.Options{
		SpecContent: specFunc,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scalar.ApiReferenceHTML(options)
	}
}
