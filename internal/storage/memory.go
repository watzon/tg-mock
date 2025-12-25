// internal/storage/memory.go
package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

// memoryFile represents a file stored in memory.
type memoryFile struct {
	data     []byte
	metadata FileMetadata
	path     string
}

// MemoryStore is an in-memory implementation of the Store interface.
type MemoryStore struct {
	mu    sync.RWMutex
	files map[string]*memoryFile
}

// Ensure MemoryStore implements Store interface at compile time.
var _ Store = (*MemoryStore)(nil)

// NewMemoryStore creates a new in-memory file store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		files: make(map[string]*memoryFile),
	}
}

// Store saves data with the given filename and MIME type, returning a unique file ID.
func (s *MemoryStore) Store(data []byte, filename string, mimeType string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileID := s.generateFileID()
	path := fmt.Sprintf("documents/%s", filename)

	// Make a copy of the data to prevent external modifications
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	s.files[fileID] = &memoryFile{
		data: dataCopy,
		metadata: FileMetadata{
			Filename: filename,
			MimeType: mimeType,
			Size:     int64(len(data)),
		},
		path: path,
	}

	return fileID, nil
}

// Get retrieves file data and metadata by file ID.
func (s *MemoryStore) Get(fileID string) ([]byte, FileMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, ok := s.files[fileID]
	if !ok {
		return nil, FileMetadata{}, ErrNotFound
	}

	return file.data, file.metadata, nil
}

// GetPath returns the virtual file path for the given file ID.
func (s *MemoryStore) GetPath(fileID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, ok := s.files[fileID]
	if !ok {
		return "", ErrNotFound
	}

	return file.path, nil
}

// Delete removes a file from the store.
func (s *MemoryStore) Delete(fileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.files, fileID)
	return nil
}

// Clear removes all files from the store.
func (s *MemoryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.files = make(map[string]*memoryFile)
	return nil
}

// generateFileID creates a unique file ID using crypto/rand.
func (s *MemoryStore) generateFileID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
