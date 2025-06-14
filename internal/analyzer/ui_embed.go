package analyzer

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed ui
var uiFS embed.FS

// getUIFileSystem returns a http.FileSystem for the embedded UI files
func getUIFileSystem() http.FileSystem {
	// Get the subdirectory "ui" from the embedded filesystem
	subFS, err := fs.Sub(uiFS, "ui")
	if err != nil {
		panic(err)
	}
	return http.FS(subFS)
}
