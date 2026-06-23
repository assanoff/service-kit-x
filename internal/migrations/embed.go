// Package migrations embeds the SQL migration files so they ship inside the
// binary and can be applied with the `migrate` command.
package migrations

import "embed"

// FS holds the embedded goose migration files.
//
//go:embed *.sql
var FS embed.FS
