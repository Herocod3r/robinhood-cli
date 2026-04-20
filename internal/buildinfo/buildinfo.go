// Package buildinfo exposes version metadata set at build time via -ldflags.
package buildinfo

// Version is the semver version of the CLI, set at build time.
var Version = "dev"

// Commit is the git short SHA, set at build time.
var Commit = "none"

// SchemaVersion is the stable JSON envelope version. Never changes within v1.
const SchemaVersion = "robinhood-cli/v1"
