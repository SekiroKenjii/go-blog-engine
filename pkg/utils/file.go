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

func EnsureFileURL(filePath string) (string, error) {
	if strings.HasPrefix(filePath, "file://") {
		if path := strings.TrimPrefix(filePath, "file://"); !filepath.IsAbs(path) {
			currentDir, err := os.Getwd()
			if err != nil {
				return "", fmt.Errorf("error getting current directory: %w", err)
			}

			resolvedPath := filepath.Join(currentDir, path)

			return "file://" + resolvedPath, nil
		}

		return filePath, nil
	}

	if filepath.IsAbs(filePath) {
		return "file://" + filePath, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	resolvedPath := filepath.Join(currentDir, filePath)

	return "file://" + resolvedPath, nil
}

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
