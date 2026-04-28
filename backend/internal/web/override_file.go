package web

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// tryServeOverrideFile serves a local override file when it exists.
// Files under overrideDir take precedence over embedded frontend assets.
func tryServeOverrideFile(c *gin.Context, overrideDir, cleanPath string) bool {
	if overrideDir == "" {
		return false
	}

	filePath := filepath.Join(overrideDir, filepath.Clean("/"+cleanPath))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}

	c.File(filePath)
	c.Abort()
	return true
}
