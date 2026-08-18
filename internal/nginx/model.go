// Package nginx provides an extensible object model and rendering engine
// for generating Nginx configuration files, from simple reverse proxies
// to complex load-balanced, cached, and secured virtual host setups.
package nginx

// LoadBalanceMethod is the algorithm used to distribute traffic across
// the servers of an Upstream block.
type LoadBalanceMethod string

const (
	LoadBalanceRoundRobin LoadBalanceMethod = "round_robin"
	LoadBalanceLeastConn  LoadBalanceMethod = "least_conn"
	LoadBalanceIPHash     LoadBalanceMethod = "ip_hash"
)

// UpstreamServer is a single backend server participating in an Upstream
// cluster.
type UpstreamServer struct {
	// Address is host:port (or a unix socket path prefixed with "unix:").
	Address string `json:"address"`
	// Weight influences how much traffic this server receives relative
	// to the others. Zero means "unset" (Nginx default of 1).
	Weight int `json:"weight,omitempty"`
	// MaxFails is the number of failed attempts before the server is
	// considered unavailable. Zero means "unset".
	MaxFails int `json:"maxFails,omitempty"`
	// FailTimeout is the duration (e.g. "10s") the server is marked
	// unavailable for after MaxFails is reached.
	FailTimeout string `json:"failTimeout,omitempty"`
	// Backup marks the server as a backup, only used when primary
	// servers are unavailable.
	Backup bool `json:"backup,omitempty"`
	// Down marks the server as permanently unavailable.
	Down bool `json:"down,omitempty"`
}

// Upstream is a named cluster of backend servers used for load balancing.
type Upstream struct {
	// Name is the identifier referenced by `proxy_pass http://<Name>;`.
	Name string `json:"name"`
	// Method is the load-balancing algorithm. Empty means Nginx's
	// default (round-robin).
	Method LoadBalanceMethod `json:"method,omitempty"`
	// Servers is the list of backend servers in the cluster.
	Servers []UpstreamServer `json:"servers"`
	// KeepAlive sets the number of idle keepalive connections per
	// worker kept open to upstream servers. Zero means "unset".
	KeepAlive int `json:"keepAlive,omitempty"`
	// CustomDirectives allows injecting raw Nginx directives inside the
	// upstream block (e.g. "zone my_zone 64k;").
	CustomDirectives []string `json:"customDirectives,omitempty"`
}

// SSL holds TLS/SSL configuration for a server block.
type SSL struct {
	Enabled bool `json:"enabled,omitempty"`
	// CertificatePath is the path to the fullchain certificate.
	CertificatePath string `json:"certificatePath,omitempty"`
	// CertificateKeyPath is the path to the private key.
	CertificateKeyPath string `json:"certificateKeyPath,omitempty"`
	// Protocols lists the allowed TLS protocol versions, e.g.
	// []string{"TLSv1.2", "TLSv1.3"}.
	Protocols []string `json:"protocols,omitempty"`
	// Ciphers is the OpenSSL cipher list string. Empty uses Nginx's
	// default.
	Ciphers string `json:"ciphers,omitempty"`
	// PreferServerCiphers enables `ssl_prefer_server_ciphers on;`.
	PreferServerCiphers bool `json:"preferServerCiphers,omitempty"`
	// HSTS enables the Strict-Transport-Security header.
	HSTS bool `json:"hsts,omitempty"`
	// HSTSMaxAge is the max-age value in seconds for HSTS. Zero uses
	// a sane default (31536000) when HSTS is enabled.
	HSTSMaxAge int `json:"hstsMaxAge,omitempty"`
	// HSTSIncludeSubDomains adds includeSubDomains to the HSTS header.
	HSTSIncludeSubDomains bool `json:"hstsIncludeSubDomains,omitempty"`
	// HSTSPreload adds preload to the HSTS header.
	HSTSPreload bool `json:"hstsPreload,omitempty"`
	// OCSPStapling enables `ssl_stapling on;` and related directives.
	OCSPStapling bool `json:"ocspStapling,omitempty"`
	// SessionCache configures `ssl_session_cache`, e.g. "shared:SSL:10m".
	SessionCache string `json:"sessionCache,omitempty"`
	// SessionTimeout configures `ssl_session_timeout`, e.g. "1d".
	SessionTimeout string `json:"sessionTimeout,omitempty"`
}

// SecurityHeaders configures common HTTP security response headers.
type SecurityHeaders struct {
	// XFrameOptions sets the X-Frame-Options header, e.g. "SAMEORIGIN".
	// Empty disables the header.
	XFrameOptions string `json:"xFrameOptions,omitempty"`
	// ContentSecurityPolicy sets the Content-Security-Policy header.
	// Empty disables the header.
	ContentSecurityPolicy string `json:"contentSecurityPolicy,omitempty"`
	// XContentTypeOptionsNoSniff enables `X-Content-Type-Options: nosniff`.
	XContentTypeOptionsNoSniff bool `json:"xContentTypeOptionsNoSniff,omitempty"`
	// ReferrerPolicy sets the Referrer-Policy header. Empty disables it.
	ReferrerPolicy string `json:"referrerPolicy,omitempty"`
	// XSSProtection sets the legacy X-XSS-Protection header. Empty
	// disables it.
	XSSProtection string `json:"xssProtection,omitempty"`
	// PermissionsPolicy sets the Permissions-Policy header. Empty
	// disables it.
	PermissionsPolicy string `json:"permissionsPolicy,omitempty"`
}

// FastCGI configures PHP-FPM / FastCGI proxying for a location.
type FastCGI struct {
	Enabled bool `json:"enabled,omitempty"`
	// Pass is the FastCGI backend, e.g. "unix:/run/php/php8.3-fpm.sock"
	// or "127.0.0.1:9000".
	Pass string `json:"pass,omitempty"`
	// Index is the default script served for directory requests, e.g.
	// "index.php".
	Index string `json:"index,omitempty"`
	// ScriptFilenameRoot is the root directory used to build
	// SCRIPT_FILENAME. When empty, the location's Root is used.
	ScriptFilenameRoot string `json:"scriptFilenameRoot,omitempty"`
	// SplitPathInfo, when set, adds a fastcgi_split_path_info directive
	// with this regex (e.g. "^(.+\\.php)(/.+)$").
	SplitPathInfo string `json:"splitPathInfo,omitempty"`
	// Params holds additional fastcgi_param entries, key -> value.
	Params map[string]string `json:"params,omitempty"`
}

// FastCGICache configures a FastCGI cache zone applied to a location.
type FastCGICache struct {
	Enabled bool `json:"enabled,omitempty"`
	// ZoneName is the shared memory zone name declared at the http level
	// and referenced by `fastcgi_cache <ZoneName>;`.
	ZoneName string `json:"zoneName,omitempty"`
	// ZonePath is the on-disk cache path for the zone definition, e.g.
	// "/var/cache/nginx/fastcgi".
	ZonePath string `json:"zonePath,omitempty"`
	// ZoneSize is the shared memory size, e.g. "10m".
	ZoneSize string `json:"zoneSize,omitempty"`
	// MaxSize is the max on-disk cache size, e.g. "1g".
	MaxSize string `json:"maxSize,omitempty"`
	// InactiveTime is how long unused cache entries are kept, e.g. "60m".
	InactiveTime string `json:"inactiveTime,omitempty"`
	// ValidCodes maps HTTP status expressions to cache durations, e.g.
	// {"200 301 302": "10m", "any": "1m"}.
	ValidCodes map[string]string `json:"validCodes,omitempty"`
	// UseStale lists conditions under which stale cache is served, e.g.
	// []string{"error", "timeout", "updating"}.
	UseStale []string `json:"useStale,omitempty"`
	// Bypass lists variables that bypass the cache when non-empty/non-zero.
	Bypass []string `json:"bypass,omitempty"`
	// SkipIf lists variables that, when non-empty/non-zero, cause the
	// response not to be cached (fastcgi_no_cache).
	SkipIf []string `json:"skipIf,omitempty"`
}

// RateLimit configures a rate limiting zone applied to a location. The
// zone itself must be declared once at the http level via
// Config.RateLimitZones.
type RateLimit struct {
	Enabled bool `json:"enabled,omitempty"`
	// Zone is the name of the previously declared limit_req_zone.
	Zone string `json:"zone,omitempty"`
	// Burst is the number of requests allowed to burst above the rate.
	Burst int `json:"burst,omitempty"`
	// Nodelay enables the `nodelay` flag on the burst.
	Nodelay bool `json:"nodelay,omitempty"`
}

// RateLimitZone declares a shared memory zone for rate limiting at the
// http level, referenced by RateLimit.Zone.
type RateLimitZone struct {
	// Name is the zone identifier.
	Name string `json:"name"`
	// Key is the variable used to key the limit, e.g. "$binary_remote_addr".
	Key string `json:"key"`
	// ZoneSize is the shared memory size, e.g. "10m".
	ZoneSize string `json:"zoneSize"`
	// Rate is the allowed rate, e.g. "10r/s".
	Rate string `json:"rate"`
}

// WebSocket enables and configures WebSocket proxying headers for a
// location.
type WebSocket struct {
	Enabled bool `json:"enabled,omitempty"`
}

// Location represents a single `location` block within a server.
type Location struct {
	// Path is the location match, e.g. "/", "/api/", "~ \\.php$".
	Path string `json:"path"`
	// Root sets the document root for this location. Empty inherits
	// from the server.
	Root string `json:"root,omitempty"`
	// Alias sets an `alias` instead of `root` for this location.
	Alias string `json:"alias,omitempty"`
	// Index lists index files, e.g. []string{"index.html", "index.htm"}.
	Index []string `json:"index,omitempty"`
	// TryFiles sets the `try_files` arguments, e.g.
	// []string{"$uri", "$uri/", "=404"}.
	TryFiles []string `json:"tryFiles,omitempty"`

	// ProxyPass, when set, enables reverse-proxying to this URL/upstream,
	// e.g. "http://backend" or "http://127.0.0.1:3000".
	ProxyPass string `json:"proxyPass,omitempty"`
	// ProxySetHeaders holds additional proxy_set_header entries.
	ProxySetHeaders map[string]string `json:"proxySetHeaders,omitempty"`

	WebSocket WebSocket `json:"webSocket,omitempty"`

	FastCGI      FastCGI      `json:"fastCGI,omitempty"`
	FastCGICache FastCGICache `json:"fastCGICache,omitempty"`

	RateLimit RateLimit `json:"rateLimit,omitempty"`

	// ReturnCode and ReturnValue, when ReturnCode != 0, emit a `return`
	// directive, e.g. `return 301 https://example.com$request_uri;`.
	ReturnCode  int    `json:"returnCode,omitempty"`
	ReturnValue string `json:"returnValue,omitempty"`

	// CustomDirectives allows injecting raw Nginx directives inside the
	// location block for anything not covered by the structured model.
	CustomDirectives []string `json:"customDirectives,omitempty"`
}

// Server represents a single `server` block (virtual host).
type Server struct {
	// ServerNames lists the domain names served, e.g.
	// []string{"example.com", "www.example.com"}.
	ServerNames []string `json:"serverNames"`
	// Listen lists listen directives, e.g. []string{"80", "443 ssl"}.
	// When empty, sensible defaults are derived from SSL/HTTP2/HTTP3.
	Listen []string `json:"listen,omitempty"`
	// HTTP2 enables HTTP/2 on the SSL listener.
	HTTP2 bool `json:"http2,omitempty"`
	// HTTP3 enables HTTP/3 (QUIC) on a UDP listener alongside TLS.
	HTTP3 bool `json:"http3,omitempty"`

	// Root is the default document root for the server.
	Root string `json:"root,omitempty"`
	// Index lists default index files for the server.
	Index []string `json:"index,omitempty"`

	SSL             SSL             `json:"ssl,omitempty"`
	SecurityHeaders SecurityHeaders `json:"securityHeaders,omitempty"`

	// RedirectToHTTPS, when true and SSL is enabled, emits a plain HTTP
	// server block that 301-redirects to HTTPS.
	RedirectToHTTPS bool `json:"redirectToHTTPS,omitempty"`

	// AccessLog and ErrorLog set log paths. Empty leaves Nginx defaults.
	AccessLog string `json:"accessLog,omitempty"`
	ErrorLog  string `json:"errorLog,omitempty"`

	// Locations holds the ordered list of location blocks.
	Locations []Location `json:"locations,omitempty"`

	// CustomDirectives allows injecting raw Nginx directives inside the
	// server block.
	CustomDirectives []string `json:"customDirectives,omitempty"`
}

// Config is the root object describing a full Nginx configuration to
// render, potentially spanning multiple upstreams and server blocks.
type Config struct {
	// Upstreams declares load-balanced backend clusters, referenced by
	// name from Location.ProxyPass (e.g. "http://my_upstream").
	Upstreams []Upstream `json:"upstreams,omitempty"`
	// RateLimitZones declares http-level limit_req_zone entries.
	RateLimitZones []RateLimitZone `json:"rateLimitZones,omitempty"`
	// Servers holds the virtual hosts to render.
	Servers []Server `json:"servers"`
}
