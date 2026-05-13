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

// ValidateWritePath validates a path for write operations.
// Unlike ValidatePath, the file doesn't need to exist yet.
// It still checks: no absolute paths, no path traversal,
// blocked basenames (.env, etc.), blocked directories (.git, etc.),
// and symlink escape from workspace root.
func ValidateWritePath(requestPath string, workspaceRoot string) (string, error) {
	if requestPath == "" {
		return "", fmt.Errorf("path is empty")
	}

	// Reject absolute paths
	if filepath.IsAbs(requestPath) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", requestPath)
	}

	// Clean path
	cleanPath := filepath.Clean(requestPath)

	// Reject path traversal
	if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return "", fmt.Errorf("path traversal is not allowed: %s", requestPath)
	}

	// Resolve workspace root (handles symlinks like /var -> /private/var on macOS)
	resolvedRoot, err := filepath.EvalSymlinks(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("cannot resolve workspace root: %s", err.Error())
	}

	// Build full path
	fullPath := filepath.Join(resolvedRoot, cleanPath)

	// Check blocked basenames
	base := filepath.Base(fullPath)
	if blockedBasenames[base] {
		return "", fmt.Errorf("access denied: %s is a restricted file", base)
	}
	if strings.HasPrefix(base, ".env") {
		return "", fmt.Errorf("access denied: %s is a restricted file", base)
	}

	// Check blocked directories in path
	rel, err := filepath.Rel(resolvedRoot, fullPath)
	if err == nil {
		parts := strings.Split(rel, string(filepath.Separator))
		for _, part := range parts {
			if blockedDirNames[part] {
				return "", fmt.Errorf("access denied: %s is in a restricted directory", requestPath)
			}
		}
	}

	// Anti-symlink-escape: find the first existing ancestor and verify it's within workspace
	dir := fullPath
	found := false
	for i := 0; i < 100; i++ { // safety limit
		resolved, err := filepath.EvalSymlinks(dir)
		if err == nil {
			if !strings.HasPrefix(resolved, resolvedRoot) {
				return "", fmt.Errorf("path escapes workspace root: %s", requestPath)
			}
			found = true
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root without finding existing path
		}
		dir = parent
	}
	if !found {
		// Path is entirely new and within resolvedRoot (filepath.Join guarantees this
		// since we rejected ..), so it's safe
	}

	return fullPath, nil
}

// blockedBashPatterns contains shell command patterns that are never allowed.
var blockedBashPatterns = []string{
	"sudo",
	"su ",
	"passwd",
	"chsh",
	"useradd",
	"userdel",
	"groupadd",
	"kill ",
	"pkill",
	"killall",
	"shutdown",
	"reboot",
	"halt",
	"poweroff",
	"mkfs",
	"fdisk",
	"dd if=",
	"dd of=",
	"parted",
	"mount",
	"umount",
	"chmod 777 /",
	"chmod 777 /etc",
	"chown",
	"apt ",
	"apt-get",
	"yum ",
	"dnf ",
	"brew ",
	"> /dev/",
	"> /etc/",
	"> /proc/",
	"> /sys/",
	"| sh",
	"| bash",
	"curl .* | sh",
	"curl .* | bash",
	"wget .* | sh",
	"wget .* | bash",
}

// IsBashCommandAllowed checks if a command contains blocked patterns.
func IsBashCommandAllowed(command string) error {
	if command == "" {
		return fmt.Errorf("command is empty")
	}

	lower := strings.ToLower(strings.TrimSpace(command))

	// Check for extremely dangerous patterns
	for _, pattern := range blockedBashPatterns {
		if strings.Contains(lower, pattern) {
			return fmt.Errorf("command contains blocked pattern: %s", pattern)
		}
	}

	return nil
}

// IsGitCommandAllowed checks if a git command is safe to execute.
func IsGitCommandAllowed(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no git arguments provided")
	}

	subcommand := args[0]

	blockedGitCommands := map[string]bool{
		"push":        true,
		"fetch":       true,
		"pull":        true,
		"rebase":      true,
		"clean":       true,
		"cherry-pick": true,
		"merge":       true,
		"gc":          true,
		"prune":       true,
		"fsck":        true,
		"update-ref":  true,
	}

	if blockedGitCommands[subcommand] {
		return fmt.Errorf("destructive git command not allowed: git %s", subcommand)
	}

	// Check for dangerous flags on allowed commands
	if subcommand == "reset" {
		return fmt.Errorf("git reset is not allowed as it can destroy uncommitted changes")
	}

	// Check for dangerous operations on submodule
	if subcommand == "submodule" && len(args) > 1 {
		subAction := args[1]
		if subAction == "update" || subAction == "deinit" {
			return fmt.Errorf("git submodule %s is not allowed", subAction)
		}
	}

	// Check for dangerous operations on tag
	if subcommand == "tag" && len(args) > 1 {
		for _, a := range args[1:] {
			if a == "-d" || a == "--delete" {
				return fmt.Errorf("git tag --delete is not allowed")
			}
		}
	}

	// Check for dangerous operations on config
	if subcommand == "config" && len(args) > 1 {
		for _, a := range args[1:] {
			if strings.HasPrefix(a, "user.") || strings.HasPrefix(a, "credential.") || strings.HasPrefix(a, "http.") || strings.HasPrefix(a, "core.ssh") {
				return fmt.Errorf("git config changes to sensitive keys are not allowed")
			}
			// Also reject any config write operations
			if a == "--add" || a == "--unset" || a == "--unset-all" || a == "--replace-all" || a == "--set" {
				return fmt.Errorf("git config write operations are not allowed")
			}
		}
	}

	return nil
}
