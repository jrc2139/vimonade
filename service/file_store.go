package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

// FileStore is an interface to store laptop files
type FileStore interface {
	// Save saves a new laptop file to the store
	Save(name string, fileType string, fileData bytes.Buffer) (string, error)
}

// DiskFileStore stores file on disk, and its info on memory
type DiskFileStore struct {
	mutex      sync.RWMutex
	fileFolder string
	files      map[string]*FileInfo
}

// FileInfo contains information of the laptop file
type FileInfo struct {
	Name string
	Type string
	Path string
}

// NewDiskFileStore returns a new DiskFileStore
func NewDiskFileStore(fileFolder string) *DiskFileStore {
	return &DiskFileStore{
		fileFolder: fileFolder,
		files:      make(map[string]*FileInfo),
	}
}

// Save adds a new file to a laptop
func (store *DiskFileStore) Save(
	name string,
	fileType string,
	fileData bytes.Buffer,
) (string, error) {
	// filePath := fmt.Sprintf("%s/%s%s", store.fileFolder, name, fileType)
	filePath := fmt.Sprintf("%s/%s", store.fileFolder, name)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot create file file: %s", err)
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write file to file: %s", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.files[name] = &FileInfo{
		Name: name,
		Type: fileType,
		Path: filePath,
	}

	return name, nil
}
