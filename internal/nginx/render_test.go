package nginx

import (
	"strings"
	"testing"
)

func mustRender(t *testing.T, cfg Config) string {
	t.Helper()
	out, err := Render(cfg)
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	return out
}

func requireContains(t *testing.T, out, substr string) {
	t.Helper()
	if !strings.Contains(out, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, out)
	}
}

func requireNotContains(t *testing.T, out, substr string) {
	t.Helper()
	if strings.Contains(out, substr) {
		t.Errorf("expected output to NOT contain %q, got:\n%s", substr, out)
	}
}

func TestRender_InvalidConfigReturnsError(t *testing.T) {
	if _, err := Render(Config{}); err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

func TestRender_SimpleReverseProxy(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames: []string{"example.com"},
				Locations: []Location{
					{Path: "/", ProxyPass: "http://127.0.0.1:3000"},
				},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "server {")
	requireContains(t, out, "listen 80;")
	requireContains(t, out, "server_name example.com;")
	requireContains(t, out, "location / {")
	requireContains(t, out, "proxy_pass http://127.0.0.1:3000;")
	requireContains(t, out, "proxy_set_header Host $host;")
}

func TestRender_UpstreamRoundRobin(t *testing.T) {
	cfg := Config{
		Upstreams: []Upstream{
			{
				Name: "backend",
				Servers: []UpstreamServer{
					{Address: "10.0.0.1:8000", Weight: 3},
					{Address: "10.0.0.2:8000", Backup: true},
				},
			},
		},
		Servers: []Server{
			{ServerNames: []string{"example.com"}, Locations: []Location{{Path: "/", ProxyPass: "http://backend"}}},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "upstream backend {")
	requireContains(t, out, "server 10.0.0.1:8000 weight=3;")
	requireContains(t, out, "server 10.0.0.2:8000 backup;")
	// Round robin has no explicit directive.
	requireNotContains(t, out, "round_robin")
}

func TestRender_UpstreamLeastConnAndIPHash(t *testing.T) {
	for _, tt := range []struct {
		method    LoadBalanceMethod
		directive string
	}{
		{LoadBalanceLeastConn, "least_conn;"},
		{LoadBalanceIPHash, "ip_hash;"},
	} {
		cfg := Config{
			Upstreams: []Upstream{
				{Name: "backend", Method: tt.method, Servers: []UpstreamServer{{Address: "10.0.0.1:8000"}}, KeepAlive: 32},
			},
			Servers: []Server{
				{ServerNames: []string{"example.com"}, Locations: []Location{{Path: "/", ProxyPass: "http://backend"}}},
			},
		}
		out := mustRender(t, cfg)
		requireContains(t, out, tt.directive)
		requireContains(t, out, "keepalive 32;")
	}
}

func TestRender_SSLAndSecurityHeaders(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames:     []string{"example.com", "www.example.com"},
				HTTP2:           true,
				RedirectToHTTPS: true,
				SSL: SSL{
					Enabled:               true,
					CertificatePath:       "/etc/ssl/example.com.crt",
					CertificateKeyPath:    "/etc/ssl/example.com.key",
					Protocols:             []string{"TLSv1.2", "TLSv1.3"},
					PreferServerCiphers:   true,
					OCSPStapling:          true,
					HSTS:                  true,
					HSTSIncludeSubDomains: true,
					HSTSPreload:           true,
				},
				SecurityHeaders: SecurityHeaders{
					XFrameOptions:              "SAMEORIGIN",
					ContentSecurityPolicy:      "default-src 'self'",
					XContentTypeOptionsNoSniff: true,
					ReferrerPolicy:             "no-referrer",
				},
				Locations: []Location{{Path: "/", Root: "/var/www/html"}},
			},
		},
	}
	out := mustRender(t, cfg)

	// HTTP -> HTTPS redirect block.
	requireContains(t, out, "return 301 https://$host$request_uri;")

	// HTTPS server block.
	requireContains(t, out, "listen 443 ssl;")
	requireContains(t, out, "http2 on;")
	requireContains(t, out, "server_name example.com www.example.com;")
	requireContains(t, out, "ssl_certificate /etc/ssl/example.com.crt;")
	requireContains(t, out, "ssl_certificate_key /etc/ssl/example.com.key;")
	requireContains(t, out, "ssl_protocols TLSv1.2 TLSv1.3;")
	requireContains(t, out, "ssl_prefer_server_ciphers on;")
	requireContains(t, out, "ssl_stapling on;")
	requireContains(t, out, `add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;`)
	requireContains(t, out, `add_header X-Frame-Options "SAMEORIGIN" always;`)
	requireContains(t, out, `add_header Content-Security-Policy "default-src 'self'" always;`)
	requireContains(t, out, `add_header X-Content-Type-Options "nosniff" always;`)
	requireContains(t, out, `add_header Referrer-Policy "no-referrer" always;`)
}

func TestRender_HTTP3(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames: []string{"example.com"},
				HTTP2:       true,
				HTTP3:       true,
				SSL: SSL{
					Enabled:            true,
					CertificatePath:    "/etc/ssl/example.com.crt",
					CertificateKeyPath: "/etc/ssl/example.com.key",
				},
				Locations: []Location{{Path: "/", Root: "/var/www/html"}},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "listen 443 quic reuseport;")
	requireContains(t, out, "http3 on;")
	requireContains(t, out, `add_header Alt-Svc 'h3=":443"; ma=86400' always;`)
}

func TestRender_WebSocketLocation(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames: []string{"ws.example.com"},
				Locations: []Location{
					{Path: "/socket", ProxyPass: "http://127.0.0.1:9001", WebSocket: WebSocket{Enabled: true}},
				},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "location /socket {")
	requireContains(t, out, "proxy_http_version 1.1;")
	requireContains(t, out, "proxy_set_header Upgrade $http_upgrade;")
	requireContains(t, out, `proxy_set_header Connection "upgrade";`)
}

func TestRender_FastCGIAndCache(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames: []string{"php.example.com"},
				Root:        "/var/www/php",
				Locations: []Location{
					{
						Path: "~ \\.php$",
						Root: "/var/www/php",
						FastCGI: FastCGI{
							Enabled:       true,
							Pass:          "unix:/run/php/php8.3-fpm.sock",
							Index:         "index.php",
							SplitPathInfo: `^(.+\.php)(/.+)$`,
							Params:        map[string]string{"PATH_INFO": "$fastcgi_path_info"},
						},
						FastCGICache: FastCGICache{
							Enabled:      true,
							ZoneName:     "PHPCACHE",
							ZonePath:     "/var/cache/nginx/fastcgi",
							ZoneSize:     "10m",
							MaxSize:      "1g",
							InactiveTime: "60m",
							ValidCodes:   map[string]string{"200 301 302": "10m"},
							UseStale:     []string{"error", "timeout", "updating"},
							Bypass:       []string{"$cookie_nocache"},
							SkipIf:       []string{"$cookie_nocache"},
						},
					},
				},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "fastcgi_cache_path /var/cache/nginx/fastcgi levels=1:2 keys_zone=PHPCACHE:10m max_size=1g inactive=60m;")
	requireContains(t, out, "fastcgi_pass unix:/run/php/php8.3-fpm.sock;")
	requireContains(t, out, "fastcgi_index index.php;")
	requireContains(t, out, `fastcgi_split_path_info ^(.+\.php)(/.+)$;`)
	requireContains(t, out, "fastcgi_param SCRIPT_FILENAME /var/www/php$fastcgi_script_name;")
	requireContains(t, out, "fastcgi_param PATH_INFO $fastcgi_path_info;")
	requireContains(t, out, "fastcgi_cache PHPCACHE;")
	requireContains(t, out, "fastcgi_cache_valid 200 301 302 10m;")
	requireContains(t, out, "fastcgi_cache_use_stale error timeout updating;")
	requireContains(t, out, "fastcgi_cache_bypass $cookie_nocache;")
	requireContains(t, out, "fastcgi_no_cache $cookie_nocache;")
}

func TestRender_RateLimit(t *testing.T) {
	cfg := Config{
		RateLimitZones: []RateLimitZone{
			{Name: "api", Key: "$binary_remote_addr", ZoneSize: "10m", Rate: "10r/s"},
		},
		Servers: []Server{
			{
				ServerNames: []string{"api.example.com"},
				Locations: []Location{
					{
						Path:      "/api/",
						ProxyPass: "http://127.0.0.1:4000",
						RateLimit: RateLimit{Enabled: true, Zone: "api", Burst: 20, Nodelay: true},
					},
				},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;")
	requireContains(t, out, "limit_req zone=api burst=20 nodelay;")
}

func TestRender_CustomDirectives(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames:      []string{"example.com"},
				CustomDirectives: []string{"client_max_body_size 50m;"},
				Locations: []Location{
					{Path: "/", ProxyPass: "http://127.0.0.1:3000", CustomDirectives: []string{"proxy_read_timeout 90s;"}},
				},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "client_max_body_size 50m;")
	requireContains(t, out, "proxy_read_timeout 90s;")
}

func TestRender_ExplicitListenOverridesDefault(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames: []string{"example.com"},
				Listen:      []string{"8080", "[::]:8080"},
				Locations:   []Location{{Path: "/", Root: "/var/www/html"}},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "listen 8080;")
	requireContains(t, out, "listen [::]:8080;")
	requireNotContains(t, out, "listen 80;")
}

func TestRender_ReturnDirective(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames: []string{"old.example.com"},
				Locations: []Location{
					{Path: "/", ReturnCode: 301, ReturnValue: "https://new.example.com$request_uri"},
				},
			},
		},
	}
	out := mustRender(t, cfg)
	requireContains(t, out, "return 301 https://new.example.com$request_uri;")
}
