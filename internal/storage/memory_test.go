// internal/storage/memory_test.go
package storage

import "testing"

func TestMemoryStore(t *testing.T) {
	s := NewMemoryStore()

	// Store file
	data := []byte("hello world")
	fileID, err := s.Store(data, "test.txt", "text/plain")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if fileID == "" {
		t.Error("fileID should not be empty")
	}

	// Get file
	retrieved, meta, err := s.Get(fileID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(data) {
		t.Errorf("got %q, want %q", retrieved, data)
	}

	if meta.Filename != "test.txt" {
		t.Errorf("filename = %q, want test.txt", meta.Filename)
	}

	if meta.MimeType != "text/plain" {
		t.Errorf("mimeType = %q, want text/plain", meta.MimeType)
	}

	if meta.Size != int64(len(data)) {
		t.Errorf("size = %d, want %d", meta.Size, len(data))
	}

	// Get path
	path, err := s.GetPath(fileID)
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}
	if path == "" {
		t.Error("path should not be empty")
	}

	// Delete
	if err := s.Delete(fileID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should be gone
	_, _, err = s.Get(fileID)
	if err == nil {
		t.Error("expected error after delete")
	}

	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_Clear(t *testing.T) {
	s := NewMemoryStore()

	// Store a couple of files
	_, err := s.Store([]byte("file1"), "file1.txt", "text/plain")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	fileID2, err := s.Store([]byte("file2"), "file2.txt", "text/plain")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Clear all files
	if err := s.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// All files should be gone
	_, _, err = s.Get(fileID2)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after Clear, got %v", err)
	}
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	s := NewMemoryStore()

	// Try to get a non-existent file
	_, _, err := s.Get("nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// GetPath for non-existent file
	_, err = s.GetPath("nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_UniqueFileIDs(t *testing.T) {
	s := NewMemoryStore()

	// Store multiple files and verify unique IDs
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		fileID, err := s.Store([]byte("data"), "test.txt", "text/plain")
		if err != nil {
			t.Fatalf("Store failed: %v", err)
		}
		if ids[fileID] {
			t.Errorf("duplicate fileID generated: %s", fileID)
		}
		ids[fileID] = true
	}
}
