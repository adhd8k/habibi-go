package services

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

type FileInfo struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Type         string `json:"type"` // "file" or "directory"
	Size         int64  `json:"size,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	RelativePath string `json:"relative_path"`
}

// SearchFiles searches for files matching the query
func (s *FileService) SearchFiles(rootPath string, query string) ([]FileInfo, error) {
	var files []FileInfo
	query = strings.ToLower(query)

	// Limit results to prevent overwhelming the UI
	maxResults := 50
	count := 0

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories and common ignore patterns
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "__pycache__" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
		}

		// Check if name matches query
		if strings.Contains(strings.ToLower(d.Name()), query) {
			relPath, _ := filepath.Rel(rootPath, path)
			
			fileType := "file"
			if d.IsDir() {
				fileType = "directory"
			}

			info, err := d.Info()
			if err != nil {
				return nil // Skip if can't get info
			}

			files = append(files, FileInfo{
				Path:         path,
				Name:         d.Name(),
				Type:         fileType,
				Size:         info.Size(),
				LastModified: info.ModTime().Format("2006-01-02 15:04:05"),
				RelativePath: relPath,
			})

			count++
			if count >= maxResults {
				return filepath.SkipAll
			}
		}

		return nil
	})

	return files, err
}

// ListFiles lists files in a directory
func (s *FileService) ListFiles(dirPath string) ([]FileInfo, error) {
	var files []FileInfo

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	rootPath := dirPath
	for _, entry := range entries {
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		path := filepath.Join(dirPath, entry.Name())
		relPath, _ := filepath.Rel(rootPath, path)

		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip if can't get info
		}

		files = append(files, FileInfo{
			Path:         path,
			Name:         entry.Name(),
			Type:         fileType,
			Size:         info.Size(),
			LastModified: info.ModTime().Format("2006-01-02 15:04:05"),
			RelativePath: relPath,
		})
	}

	return files, nil
}