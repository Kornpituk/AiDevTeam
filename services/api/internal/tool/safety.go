package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// blockedPaths contains filenames or directory names that should never be read.
var blockedBasenames = map[string]bool{
	".env":       true,
	".env.local": true,
	".env.*":     true, // handled via prefix match
}

// blockedDirPrefixes contains directory name prefixes that are always blocked.
var blockedDirNames = map[string]bool{
	".git":        true,
	"node_modules": true,
	".next":       true,
	"dist":        true,
	"build":       true,
	"vendor":      true,
	"__pycache__": true,
}

// binaryExtensions contains file extensions for binary files that should not be read.
var binaryExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".bmp":  true,
	".ico":  true,
	".svg":  true,
	".webp": true,
	".woff": true,
	".woff2": true,
	".eot":  true,
	".ttf":  true,
	".otf":  true,
	".pdf":  true,
	".zip":  true,
	".tar":  true,
	".gz":   true,
	".bz2":  true,
	".7z":   true,
	".rar":  true,
	".exe":  true,
	".dll":  true,
	".so":   true,
	".dylib": true,
	".bin":  true,
	".o":    true,
	".a":    true,
	".class": true,
	".pyc":  true,
	".db":   true,
	".sqlite": true,
	".DS_Store": true,
}

// ValidatePath checks that the given path is safe to access.
// Returns the safe absolute path or an error.
func ValidatePath(requestPath string, workspaceRoot string) (string, error) {
	if requestPath == "" {
		return "", fmt.Errorf("path is empty")
	}

	// Reject absolute paths
	if filepath.IsAbs(requestPath) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", requestPath)
	}

	// Clean the path to resolve any ./ or simple constructs (NOT ../)
	cleanPath := filepath.Clean(requestPath)

	// Reject path traversal that goes above workspace root
	if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return "", fmt.Errorf("path traversal is not allowed: %s", requestPath)
	}

	// Resolve workspace root to handle symlinks (e.g., /var -> /private/var on macOS)
	resolvedRoot, err := filepath.EvalSymlinks(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("cannot resolve workspace root: %s", err.Error())
	}

	// Build the full path
	fullPath := filepath.Join(resolvedRoot, cleanPath)

	// Verify the resolved path is within workspace root (anti-symlink protection)
	resolvedPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", requestPath)
		}
		// If we can't resolve symlinks, reject for safety
		return "", fmt.Errorf("cannot resolve path: %s", requestPath)
	}

	if !strings.HasPrefix(resolvedPath, resolvedRoot) {
		return "", fmt.Errorf("path escapes workspace root: %s", requestPath)
	}

	// Check blocked basenames
	base := filepath.Base(resolvedPath)
	if blockedBasenames[base] {
		return "", fmt.Errorf("access denied: %s is a restricted file", base)
	}
	// Handle .env.* pattern
	if strings.HasPrefix(base, ".env") {
		return "", fmt.Errorf("access denied: %s is a restricted file", base)
	}

	// Check if any component in the path is a blocked directory
	rel, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err == nil {
		parts := strings.Split(rel, string(filepath.Separator))
		for _, part := range parts {
			if blockedDirNames[part] {
				return "", fmt.Errorf("access denied: %s is in a restricted directory", requestPath)
			}
		}
	}

	return resolvedPath, nil
}

// IsBinaryFile checks if a file appears to be binary based on extension.
func IsBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return binaryExtensions[ext]
}

// HasNullBytes checks if content contains null bytes (binary content detection).
func HasNullBytes(content []byte, checkBytes int) bool {
	limit := checkBytes
	if len(content) < limit {
		limit = len(content)
	}
	for i := 0; i < limit; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}
