// Command kznginx is the KzNginxGenerator CLI: it generates robust,
// optimized Nginx configurations, from simple reverse proxies to complex
// load-balanced/SSL/FastCGI architectures, via a local web UI or the
// command line.
package main

import (
	"fmt"
	"os"

	"github.com/karoza/kz-nginx-generator/cmd"
)

// Version is the kznginx release version, e.g. "v1.0.0" or
// "v1.0.0-beta.1". Overridden at build time via
// `-ldflags "-X main.Version=..."`.
var Version = "dev"

// GitCommit is the git revision the binary was built from. Overridden
// at build time via `-ldflags "-X main.GitCommit=..."`.
var GitCommit = "none"

func main() {
	cmd.Version = Version
	cmd.GitCommit = GitCommit

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
