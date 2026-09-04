package frontend

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distEmbedFS embed.FS

// DistFS returns the filesystem pointing to the contents of the frontend build (dist directory).
func DistFS() (fs.FS, error) {
	return fs.Sub(distEmbedFS, "dist")
}
