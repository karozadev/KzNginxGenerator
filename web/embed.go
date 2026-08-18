// Package web embeds the static assets (HTML/JS/CSS) that power the
// KzNginxGenerator local Web UI, served by internal/server.
package web

import "embed"

//go:embed index.html static
var FS embed.FS
