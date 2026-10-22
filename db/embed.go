// Package db embeds the numbered SQL migrations. They are the source of truth
// for the schema; the Go structs map onto them, never the other way round.
package db

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS
