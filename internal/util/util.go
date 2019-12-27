package util

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
)

// IsHTMLFile checks if the given path indicates a HTML file type
func IsHTMLFile(path string) bool {
	exts := strings.Split(path, ".")
	ext := strings.ToLower(exts[len(exts)-1])
	htmlExts := map[string]bool{"html": true, "htm": true}
	return htmlExts[ext]
}

// StringInSlice checks if the given str exists within the given list slice.
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// IsPortFree checks if the given port number is available or not
func IsPortFree(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

// FormatRootPath standardises the root path used by file servers
func FormatRootPath(path string) string {
	basePath := filepath.Base(path)

	if strings.Contains(basePath, ".") {
		return filepath.Dir(path)
	}

	return path
}
