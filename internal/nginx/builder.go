package nginx

// SimpleReverseProxyOptions configures NewSimpleReverseProxyConfig.
type SimpleReverseProxyOptions struct {
	// Domain is the primary server_name for the vhost (required).
	Domain string
	// ProxyPass is the backend URL to reverse-proxy to, e.g.
	// "http://localhost:8000" (required).
	ProxyPass string
	// WebSocket enables WebSocket upgrade headers on the root location.
	WebSocket bool
	// SSLCertificatePath and SSLCertificateKeyPath, when both set,
	// enable HTTPS with an HTTP->HTTPS redirect and HSTS.
	SSLCertificatePath    string
	SSLCertificateKeyPath string
}

// NewSimpleReverseProxyConfig builds a minimal single-server Config that
// reverse-proxies all traffic for a domain to a backend URL. It is the
// building block behind `kznginx generate --domain ... --proxy ...`.
func NewSimpleReverseProxyConfig(opts SimpleReverseProxyOptions) Config {
	srv := Server{
		ServerNames: []string{opts.Domain},
		Locations: []Location{
			{
				Path:      "/",
				ProxyPass: opts.ProxyPass,
				WebSocket: WebSocket{Enabled: opts.WebSocket},
			},
		},
	}

	if opts.SSLCertificatePath != "" && opts.SSLCertificateKeyPath != "" {
		srv.SSL = SSL{
			Enabled:               true,
			CertificatePath:       opts.SSLCertificatePath,
			CertificateKeyPath:    opts.SSLCertificateKeyPath,
			Protocols:             []string{"TLSv1.2", "TLSv1.3"},
			PreferServerCiphers:   true,
			HSTS:                  true,
			HSTSIncludeSubDomains: true,
		}
		srv.HTTP2 = true
		srv.RedirectToHTTPS = true
	}

	return Config{Servers: []Server{srv}}
}
