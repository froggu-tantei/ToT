package storage

import (
	"mime/multipart"
)

// FileStorage defines the interface for file operations
type FileStorage interface {
	// Store saves a file and returns its public path
	Store(file multipart.File, filename string) (string, error)

	// Delete removes a file by its path
	Delete(path string) error

	// GetPublicURL returns the public URL for a stored file
	GetPublicURL(path string) string
}
