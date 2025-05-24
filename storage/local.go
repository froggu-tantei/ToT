package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// LocalStorage implements FileStorage for local filesystem storage
type LocalStorage struct {
	UploadDir string
	BaseURL   string
}

// NewLocalStorage creates a new LocalStorage instance
func NewLocalStorage(uploadDir, baseURL string) *LocalStorage {
	return &LocalStorage{
		UploadDir: uploadDir,
		BaseURL:   baseURL,
	}
}

// Store saves a file to the local filesystem and returns its relative path
func (ls *LocalStorage) Store(file multipart.File, filename string) (string, error) {
	// Create upload directory if it doesn't exist
	if _, err := os.Stat(ls.UploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(ls.UploadDir, 0755); err != nil {
			return "", err
		}
	}

	// Create destination file
	filePath := filepath.Join(ls.UploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Return the file path relative to upload directory
	return "/" + filepath.Join(filepath.Base(ls.UploadDir), filename), nil
}

// Delete removes a file from the local filesystem
func (ls *LocalStorage) Delete(path string) error {
	// Handle paths that start with "/"
	if filepath.IsAbs(path) {
		path = path[1:] // Remove leading "/"
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File already doesn't exist, no need to delete
	}

	// Delete file
	return os.Remove(path)
}

// GetPublicURL returns the public URL for a stored file
func (ls *LocalStorage) GetPublicURL(path string) string {
	// If path already starts with baseURL, return it as is
	if ls.BaseURL == "" {
		return path
	}

	// Ensure path starts with "/"
	if path != "" && path[0] != '/' {
		path = "/" + path
	}

	return ls.BaseURL + path
}
