package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const FileURLPrefix = "file://"

// EnsureFileURL ensures that the provided file path is a valid file URL.
// If the path is relative, it resolves it against the current working directory.
// If the path is already a file URL, it checks if the path is absolute.
// If the path is not absolute, it resolves it against the current working directory.
// It returns the file URL as a string or an error if any issues occur during the process.
// The returned file URL will always start with "file://".
func EnsureFileURL(filePath string) (string, error) {
	if after, ok := strings.CutPrefix(filePath, FileURLPrefix); ok {
		if path := after; !filepath.IsAbs(path) {
			currentDir, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("error getting current directory: %w", err)
			}

			resolvedPath := filepath.Join(currentDir, path)

			return FileURLPrefix + resolvedPath, nil
		}

		return filePath, nil
	}

	if filepath.IsAbs(filePath) {
		return FileURLPrefix + filePath, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	resolvedPath := filepath.Join(currentDir, filePath)

	return FileURLPrefix + resolvedPath, nil
}

// FetchContentFromURL fetches the content from a given file URL.
// It performs an HTTP GET request to the URL and returns the content as a string.
// If the URL scheme is not "file", it returns an error.
func FetchContentFromURL(fileURL string) (string, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("error getting file content: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading file content: %w", err)
	}

	return string(content), nil
}

// ReadFileFromURL reads the content of a file from a given file URL.
// It parses the URL, checks if the scheme is "file", and reads the file content.
// If the URL scheme is not "file", it returns an error.
// It returns the file content as a byte slice or an error if any issues occur.
func ReadFileFromURL(fileURL string) ([]byte, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	if parsedURL.Scheme != "file" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	return os.ReadFile(parsedURL.Path)
}
