// Package cmd implements the kznginx command-line interface (Cobra):
// `kznginx ui`, `kznginx generate`, and `kznginx version`.
package cmd

import "github.com/spf13/cobra"

// Version and GitCommit are populated by main() from the package-level
// variables it declares, which are themselves injected at build time via
// linker flags (-ldflags "-X main.Version=... -X main.GitCommit=...").
// They default to "dev" / "none" for local `go run`/`go build` without
// ldflags.
var (
	Version   = "dev"
	GitCommit = "none"
)

var rootCmd = &cobra.Command{
	Use:   "kznginx",
	Short: "KzNginxGenerator (by Karoza) — générateur de configurations Nginx",
	Long: `KzNginxGenerator (by Karoza) génère des configurations Nginx robustes et
optimisées, du simple reverse proxy aux architectures complexes
(load balancing, SSL/TLS, FastCGI, cache, rate limiting), via une
interface web locale ou en ligne de commande.`,
	SilenceUsage: true,
}

// Execute runs the root kznginx command.
func Execute() error {
	return rootCmd.Execute()
}
