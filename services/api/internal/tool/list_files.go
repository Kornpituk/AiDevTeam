package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListFilesInput represents the input for the list_files tool.
type ListFilesInput struct {
	Path  string `json:"path"`
	Depth int    `json:"depth,omitempty"` // 0 or 1 means shallow, 2+ means deeper
}

// FileEntry represents a single file or directory entry.
type FileEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size,omitempty"`
}

// ListFilesOutput represents the output from the list_files tool.
type ListFilesOutput struct {
	Path    string      `json:"path"`
	Entries []FileEntry `json:"entries"`
	Total   int         `json:"total"`
}

// blockedDirNames is also used during listing; we skip entries in blocked dirs.
// ListFiles lists files and directories at the given path with depth control.
func ListFiles(input ListFilesInput, workspaceRoot string) (*ListFilesOutput, error) {
	listPath := input.Path
	if listPath == "" {
		listPath = "."
	}

	safePath, err := ValidatePath(listPath, workspaceRoot)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(safePath)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %s", listPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", listPath)
	}

	depth := input.Depth
	if depth < 1 {
		depth = 1
	}

	var entries []FileEntry

	if depth == 1 {
		// Shallow list
		dirEntries, err := os.ReadDir(safePath)
		if err != nil {
			return nil, fmt.Errorf("cannot list directory: %s", listPath)
		}
		for _, entry := range dirEntries {
			if isBlockedDir(entry.Name()) {
				continue
			}
			if isBlockedFile(entry.Name()) {
				continue
			}
			fi, err := entry.Info()
			size := int64(0)
			if err == nil {
				size = fi.Size()
			}
			entries = append(entries, FileEntry{
				Name:  entry.Name(),
				IsDir: entry.IsDir(),
				Size:  size,
			})
		}
	} else {
		// Recursive walk with depth limit
		baseRel := strings.TrimPrefix(listPath, ".")
		baseRel = strings.TrimPrefix(baseRel, "/")

		err := filepath.WalkDir(safePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // skip errors
			}

			relPath, _ := filepath.Rel(safePath, path)
			if relPath == "." {
				return nil
			}

			// Check depth
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) > depth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip blocked entries
			for _, part := range parts {
				if isBlockedDir(part) {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
			if isBlockedFile(d.Name()) {
				return nil
			}

			fi, err := d.Info()
			size := int64(0)
			if err == nil {
				size = fi.Size()
			}
			entries = append(entries, FileEntry{
				Name:  relPath,
				IsDir: d.IsDir(),
				Size:  size,
			})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error listing directory: %s", listPath)
		}
	}

	// Sort: directories first, then files, alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})

	return &ListFilesOutput{
		Path:    listPath,
		Entries: entries,
		Total:   len(entries),
	}, nil
}

func isBlockedDir(name string) bool {
	// Skip hidden secret dirs
	return blockedDirNames[name]
}

func isBlockedFile(name string) bool {
	if blockedBasenames[name] {
		return true
	}
	if strings.HasPrefix(name, ".env") {
		return true
	}
	return false
}
