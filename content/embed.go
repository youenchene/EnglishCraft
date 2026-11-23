// Package content holds the versioned, offline teaching library.
package content

import "embed"

//go:embed curriculum/*.json missions/*.json
var Files embed.FS
