// internal/storage/store.go
package storage

import "errors"

// ErrNotFound is returned when a file is not found in the store.
var ErrNotFound = errors.New("file not found")

// FileMetadata contains metadata about a stored file.
type FileMetadata struct {
	Filename string
	MimeType string
	Size     int64
}

// Store defines the interface for file storage operations.
type Store interface {
	// Store saves data with the given filename and MIME type, returning a unique file ID.
	Store(data []byte, filename string, mimeType string) (fileID string, err error)

	// Get retrieves file data and metadata by file ID.
	Get(fileID string) (data []byte, metadata FileMetadata, err error)

	// GetPath returns the virtual file path for the given file ID.
	GetPath(fileID string) (filePath string, err error)

	// Delete removes a file from the store.
	Delete(fileID string) error

	// Clear removes all files from the store.
	Clear() error
}
