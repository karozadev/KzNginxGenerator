package nginx

import (
	"errors"
	"fmt"
)

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
	}
	return errors.Join(errs...)
}

// Validate checks a single Server for structural errors.
func (s Server) Validate() error {
	var errs []error
	if len(s.ServerNames) == 0 {
		errs = append(errs, errors.New("at least one server name is required"))
	}
	if s.SSL.Enabled {
		if s.SSL.CertificatePath == "" {
			errs = append(errs, errors.New("ssl: certificate path is required when ssl is enabled"))
		}
		if s.SSL.CertificateKeyPath == "" {
			errs = append(errs, errors.New("ssl: certificate key path is required when ssl is enabled"))
		}
	}
	if s.HTTP3 && !s.SSL.Enabled {
		errs = append(errs, errors.New("http3 requires ssl to be enabled"))
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
	if l.FastCGI.Enabled && l.FastCGI.Pass == "" {
		errs = append(errs, errors.New("fastcgi: pass is required when fastcgi is enabled"))
	}
	if l.FastCGICache.Enabled {
		if !l.FastCGI.Enabled {
			errs = append(errs, errors.New("fastcgiCache: fastcgi must be enabled to use fastcgi caching"))
		}
		if l.FastCGICache.ZonePath == "" {
		errs = append(errs, errors.New("fastcgiCache: zone path is required when cache is enabled"))
	}
	if l.FastCGICache.ZoneSize == "" {
		errs = append(errs, errors.New("fastcgiCache: zone size is required when cache is enabled"))
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
