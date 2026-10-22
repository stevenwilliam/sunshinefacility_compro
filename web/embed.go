// Package web embeds the built single-page app (web/dist) into the binary.
// Build the frontend first: `make web` (CLAUDE.md, Makefile).
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist is the built site with dist/ stripped from the paths.
func Dist() (fs.FS, error) { return fs.Sub(dist, "dist") }
