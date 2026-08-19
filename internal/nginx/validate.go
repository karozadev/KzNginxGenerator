package nginx

import (
	"errors"
	"fmt"
	"strings"
)

// unsafeChars lists characters that could break out of a quoted or
// single-line Nginx directive argument (a stray '"' closes a quoted
// string early, e.g. in `add_header X "..." always;`; a newline starts
// a new directive even without quotes). Checked wherever a value is
// rendered as a single directive argument.
//
// Fields meant to hold raw Nginx syntax (CustomDirectives, ProxyPass,
// rate-limit zone refs, ...) are excluded: they're a documented escape
// hatch for arbitrary directives, so ';'/'{'/'}' are expected there.
var unsafeChars = []string{"\"", "\n", "\r"}

// containsUnsafeChars reports whether s contains a character that could
// break out of a quoted or single-line Nginx directive argument.
func containsUnsafeChars(s string) bool {
	for _, c := range unsafeChars {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}

// checkSafe returns an error if s contains characters unsafe to
// interpolate into a single Nginx directive argument, wrapping the
// field name for context.
func checkSafe(field, s string) error {
	if containsUnsafeChars(s) {
		return fmt.Errorf("%s: value must not contain quotes or newlines", field)
	}
	return nil
}

// Validate checks the Config for structural and semantic errors that
// would prevent generating a valid Nginx configuration. It returns a
// joined error (see errors.Join) describing every problem found, or nil
// if the configuration is valid.
func (c Config) Validate() error {
	var errs []error

	if len(c.Servers) == 0 {
		errs = append(errs, errors.New("config: at least one server is required"))
	}

	seenUpstreams := map[string]bool{}
	for i, up := range c.Upstreams {
		if err := up.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("upstream[%d] %q: %w", i, up.Name, err))
		}
		if up.Name != "" {
			if seenUpstreams[up.Name] {
				errs = append(errs, fmt.Errorf("upstream[%d]: duplicate upstream name %q", i, up.Name))
			}
			seenUpstreams[up.Name] = true
		}
	}

	zones := map[string]bool{}
	for i, z := range c.RateLimitZones {
		if z.Name == "" {
			errs = append(errs, fmt.Errorf("rateLimitZone[%d]: name is required", i))
		}
		if z.Key == "" {
			errs = append(errs, fmt.Errorf("rateLimitZone[%d]: key is required", i))
		}
		if z.Rate == "" {
			errs = append(errs, fmt.Errorf("rateLimitZone[%d]: rate is required", i))
		}
		if z.Name != "" {
			if zones[z.Name] {
				errs = append(errs, fmt.Errorf("rateLimitZone[%d]: duplicate zone name %q", i, z.Name))
			}
			zones[z.Name] = true
		}
	}

	for i, srv := range c.Servers {
		if err := srv.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("server[%d]: %w", i, err))
		}
		for j, loc := range srv.Locations {
			if loc.RateLimit.Enabled && !zones[loc.RateLimit.Zone] {
				errs = append(errs, fmt.Errorf("server[%d].location[%d]: rate limit references undeclared zone %q", i, j, loc.RateLimit.Zone))
			}
		}
	}

	return errors.Join(errs...)
}

// Validate checks a single Upstream for structural errors.
func (u Upstream) Validate() error {
	var errs []error
	if u.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if len(u.Servers) == 0 {
		errs = append(errs, errors.New("at least one server is required"))
	}
	for i, s := range u.Servers {
		if s.Address == "" {
			errs = append(errs, fmt.Errorf("server[%d]: address is required", i))
		}
		if err := checkSafe(fmt.Sprintf("server[%d].address", i), s.Address); err != nil {
			errs = append(errs, err)
		}
		if err := checkSafe(fmt.Sprintf("server[%d].failTimeout", i), s.FailTimeout); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Validate checks a single Server for structural errors.
func (s Server) Validate() error {
	var errs []error
	if len(s.ServerNames) == 0 {
		errs = append(errs, errors.New("at least one server name is required"))
	}
	for i, name := range s.ServerNames {
		if err := checkSafe(fmt.Sprintf("serverNames[%d]", i), name); err != nil {
			errs = append(errs, err)
		}
	}
	if err := checkSafe("root", s.Root); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("accessLog", s.AccessLog); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("errorLog", s.ErrorLog); err != nil {
		errs = append(errs, err)
	}

	if s.SSL.Enabled {
		if s.SSL.CertificatePath == "" {
			errs = append(errs, errors.New("ssl: certificate path is required when ssl is enabled"))
		}
		if s.SSL.CertificateKeyPath == "" {
			errs = append(errs, errors.New("ssl: certificate key path is required when ssl is enabled"))
		}
		if err := checkSafe("ssl.certificatePath", s.SSL.CertificatePath); err != nil {
			errs = append(errs, err)
		}
		if err := checkSafe("ssl.certificateKeyPath", s.SSL.CertificateKeyPath); err != nil {
			errs = append(errs, err)
		}
		if err := checkSafe("ssl.ciphers", s.SSL.Ciphers); err != nil {
			errs = append(errs, err)
		}
		if err := checkSafe("ssl.sessionCache", s.SSL.SessionCache); err != nil {
			errs = append(errs, err)
		}
		if err := checkSafe("ssl.sessionTimeout", s.SSL.SessionTimeout); err != nil {
			errs = append(errs, err)
		}
	}
	if s.HTTP3 && !s.SSL.Enabled {
		errs = append(errs, errors.New("http3 requires ssl to be enabled"))
	}

	if err := checkSafe("securityHeaders.xFrameOptions", s.SecurityHeaders.XFrameOptions); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("securityHeaders.contentSecurityPolicy", s.SecurityHeaders.ContentSecurityPolicy); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("securityHeaders.referrerPolicy", s.SecurityHeaders.ReferrerPolicy); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("securityHeaders.xssProtection", s.SecurityHeaders.XSSProtection); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("securityHeaders.permissionsPolicy", s.SecurityHeaders.PermissionsPolicy); err != nil {
		errs = append(errs, err)
	}

	for i, loc := range s.Locations {
		if err := loc.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("location[%d] %q: %w", i, loc.Path, err))
		}
	}
	return errors.Join(errs...)
}

// Validate checks a single Location for structural errors.
func (l Location) Validate() error {
	var errs []error
	if l.Path == "" {
		errs = append(errs, errors.New("path is required"))
	}
	if err := checkSafe("root", l.Root); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("alias", l.Alias); err != nil {
		errs = append(errs, err)
	}
	for key, value := range l.ProxySetHeaders {
		if err := checkSafe(fmt.Sprintf("proxySetHeaders[%q] key", key), key); err != nil {
			errs = append(errs, err)
		}
		if err := checkSafe(fmt.Sprintf("proxySetHeaders[%q] value", key), value); err != nil {
			errs = append(errs, err)
		}
	}

	if l.FastCGI.Enabled && l.FastCGI.Pass == "" {
		errs = append(errs, errors.New("fastcgi: pass is required when fastcgi is enabled"))
	}
	if err := checkSafe("fastCGI.pass", l.FastCGI.Pass); err != nil {
		errs = append(errs, err)
	}
	if err := checkSafe("fastCGI.index", l.FastCGI.Index); err != nil {
		errs = append(errs, err)
	}
	for key, value := range l.FastCGI.Params {
		if err := checkSafe(fmt.Sprintf("fastCGI.params[%q] key", key), key); err != nil {
			errs = append(errs, err)
		}
		if err := checkSafe(fmt.Sprintf("fastCGI.params[%q] value", key), value); err != nil {
			errs = append(errs, err)
		}
	}

	if l.FastCGICache.Enabled {
		if !l.FastCGI.Enabled {
			errs = append(errs, errors.New("fastcgiCache: fastcgi must be enabled to use fastcgi caching"))
		}
		if l.FastCGICache.ZoneName == "" {
			errs = append(errs, errors.New("fastcgiCache: zone name is required when cache is enabled"))
		}
	}
	if l.RateLimit.Enabled && l.RateLimit.Zone == "" {
		errs = append(errs, errors.New("rateLimit: zone is required when rate limiting is enabled"))
	}
	return errors.Join(errs...)
}