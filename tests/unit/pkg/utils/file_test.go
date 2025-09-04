package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SekiroKenjii/go-blog-engine/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// Test constants
const (
	fileScheme   = "file://"
	testFileName = "test.txt"
	testFilePath = "/tmp/test.txt"
	testFileURL  = "file:///tmp/test.txt"
)

func TestEnsureFileURL(t *testing.T) {
	// Get current directory for test setup
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected func() string
		hasError bool
	}{
		{
			name:  "absolute path",
			input: testFilePath,
			expected: func() string {
				return testFileURL
			},
			hasError: false,
		},
		{
			name:  "relative path",
			input: testFileName,
			expected: func() string {
				return fileScheme + filepath.Join(currentDir, testFileName)
			},
			hasError: false,
		},
		{
			name:  "already file URL with absolute path",
			input: testFileURL,
			expected: func() string {
				return testFileURL
			},
			hasError: false,
		},
		{
			name:  "file URL with relative path",
			input: fileScheme + testFileName,
			expected: func() string {
				return fileScheme + filepath.Join(currentDir, testFileName)
			},
			hasError: false,
		},
		{
			name:  "empty path",
			input: "",
			expected: func() string {
				return fileScheme + currentDir
			},
			hasError: false,
		},
		{
			name:  "nested relative path",
			input: "dir/subdir/" + testFileName,
			expected: func() string {
				return fileScheme + filepath.Join(currentDir, "dir", "subdir", testFileName)
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.EnsureFileURL(tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected(), result)
				assert.True(t, strings.HasPrefix(result, fileScheme))
			}
		})
	}
}

func TestFetchContentFromURL(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test content"))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name          string
		url           string
		expectError   bool
		expectContent string
	}{
		{
			name:          "successful fetch",
			url:           server.URL + "/test",
			expectError:   false,
			expectContent: "test content",
		},
		{
			name:          "server error",
			url:           server.URL + "/error",
			expectError:   false, // HTTP errors don't cause Go errors, just different status
			expectContent: "",    // Empty content for error responses
		},
		{
			name:          "not found",
			url:           server.URL + "/notfound",
			expectError:   false,
			expectContent: "", // Server might return empty for 404
		},
		{
			name:        "invalid URL",
			url:         "not-a-valid-url",
			expectError: true,
		},
		{
			name:        "unreachable URL",
			url:         "http://localhost:99999/test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := utils.FetchContentFromURL(tt.url)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, content)
			} else {
				assert.NoError(t, err)
				if tt.expectContent != "" {
					assert.Equal(t, tt.expectContent, content)
				}
			}
		})
	}
}

func TestReadFileFromURL(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	testContent := "test file content"
	_, err = tempFile.WriteString(testContent)
	assert.NoError(t, err)
	tempFile.Close()

	tests := []struct {
		name          string
		url           string
		expectError   bool
		expectContent string
	}{
		{
			name:          "valid file URL",
			url:           fileScheme + tempFile.Name(),
			expectError:   false,
			expectContent: testContent,
		},
		{
			name:        "non-file URL scheme",
			url:         "http://example.com/test",
			expectError: true,
		},
		{
			name:        "invalid URL",
			url:         "not-a-valid-url",
			expectError: true,
		},
		{
			name:        "file URL to non-existent file",
			url:         "file:///non/existent/file.txt",
			expectError: true,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
		{
			name:        "malformed file URL",
			url:         fileScheme,
			expectError: true, // Empty file path should cause an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := utils.ReadFileFromURL(tt.url)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, content)
			} else {
				assert.NoError(t, err)
				if tt.expectContent != "" {
					assert.Equal(t, []byte(tt.expectContent), content)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkEnsureFileURL(b *testing.B) {
	testPath := testFilePath
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.EnsureFileURL(testPath)
	}
}

func BenchmarkEnsureFileURLRelative(b *testing.B) {
	testPath := testFileName
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.EnsureFileURL(testPath)
	}
}
