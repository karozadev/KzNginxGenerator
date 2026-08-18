package nginx

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var funcMap = template.FuncMap{
	"join":        strings.Join,
	"lbDirective": lbDirective,
	"hstsMaxAge":  hstsMaxAge,
	"listenLines": listenLines,
}

var tmpl = template.Must(template.New("nginx").Funcs(funcMap).ParseFS(templateFS, "templates/*.tmpl"))

// lbDirective maps a LoadBalanceMethod to its Nginx directive name.
// Round-robin has no directive since it is Nginx's implicit default.
func lbDirective(m LoadBalanceMethod) string {
	switch m {
	case LoadBalanceLeastConn:
		return "least_conn"
	case LoadBalanceIPHash:
		return "ip_hash"
	default:
		return ""
	}
}

// hstsMaxAge returns the configured HSTS max-age, defaulting to one year
// (31536000 seconds) when unset.
func hstsMaxAge(s SSL) int {
	if s.HSTSMaxAge > 0 {
		return s.HSTSMaxAge
	}
	return 31536000
}

// listenLines derives the `listen` directive values for a Server. When
// Listen is explicitly set it is used as-is; otherwise sensible defaults
// are derived from whether SSL is enabled.
func listenLines(s Server) []string {
	if len(s.Listen) > 0 {
		return s.Listen
	}
	if s.SSL.Enabled {
		return []string{"443 ssl"}
	}
	return []string{"80"}
}

// collectFastCGICacheZones gathers the unique fastcgi_cache_path zone
// declarations referenced across all servers/locations of the config, so
// they can be emitted once at the top of the file (as required by
// Nginx, which declares cache zones at the http level).
func collectFastCGICacheZones(cfg Config) []FastCGICache {
	seen := map[string]bool{}
	var zones []FastCGICache
	for _, srv := range cfg.Servers {
		for _, loc := range srv.Locations {
			c := loc.FastCGICache
			if !c.Enabled || c.ZonePath == "" || c.ZoneName == "" {
				continue
			}
			if seen[c.ZoneName] {
				continue
			}
			seen[c.ZoneName] = true
			zones = append(zones, c)
		}
	}
	return zones
}

// Render renders a full Nginx configuration from cfg and returns it as a
// string. The config is validated first; a validation error is returned
// unmodified (wrapped) rather than producing partial output.
func Render(cfg Config) (string, error) {
	var sb strings.Builder
	if err := RenderTo(&sb, cfg); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// RenderTo validates cfg and writes the rendered Nginx configuration to w.
func RenderTo(w io.Writer, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	for _, zone := range collectFastCGICacheZones(cfg) {
		if err := tmpl.ExecuteTemplate(w, "fastcgicachezone.tmpl", zone); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	for _, zone := range cfg.RateLimitZones {
		if err := tmpl.ExecuteTemplate(w, "ratelimitzone.tmpl", zone); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	for _, up := range cfg.Upstreams {
		if err := tmpl.ExecuteTemplate(w, "upstream.tmpl", up); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	for _, srv := range cfg.Servers {
		if srv.SSL.Enabled && srv.RedirectToHTTPS {
			if err := tmpl.ExecuteTemplate(w, "redirect_server.tmpl", srv); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		if err := tmpl.ExecuteTemplate(w, "server.tmpl", srv); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	return nil
}
