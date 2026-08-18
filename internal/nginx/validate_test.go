package nginx

import "testing"

func TestConfigValidate_NoServers(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for config with no servers, got nil")
	}
}

func TestConfigValidate_DuplicateUpstreamNames(t *testing.T) {
	cfg := Config{
		Upstreams: []Upstream{
			{Name: "backend", Servers: []UpstreamServer{{Address: "127.0.0.1:8000"}}},
			{Name: "backend", Servers: []UpstreamServer{{Address: "127.0.0.1:8001"}}},
		},
		Servers: []Server{
			{ServerNames: []string{"example.com"}, Locations: []Location{{Path: "/"}}},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for duplicate upstream names, got nil")
	}
}

func TestConfigValidate_RateLimitZoneMustBeDeclared(t *testing.T) {
	cfg := Config{
		Servers: []Server{
			{
				ServerNames: []string{"example.com"},
				Locations: []Location{
					{Path: "/", RateLimit: RateLimit{Enabled: true, Zone: "undeclared"}},
				},
			},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for undeclared rate limit zone, got nil")
	}
}

func TestConfigValidate_RateLimitZoneDeclaredOK(t *testing.T) {
	cfg := Config{
		RateLimitZones: []RateLimitZone{
			{Name: "api", Key: "$binary_remote_addr", ZoneSize: "10m", Rate: "10r/s"},
		},
		Servers: []Server{
			{
				ServerNames: []string{"example.com"},
				Locations: []Location{
					{Path: "/", RateLimit: RateLimit{Enabled: true, Zone: "api"}},
				},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestUpstreamValidate(t *testing.T) {
	tests := []struct {
		name    string
		up      Upstream
		wantErr bool
	}{
		{"valid", Upstream{Name: "backend", Servers: []UpstreamServer{{Address: "127.0.0.1:8000"}}}, false},
		{"missing name", Upstream{Servers: []UpstreamServer{{Address: "127.0.0.1:8000"}}}, true},
		{"no servers", Upstream{Name: "backend"}, true},
		{"server missing address", Upstream{Name: "backend", Servers: []UpstreamServer{{}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.up.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServerValidate(t *testing.T) {
	tests := []struct {
		name    string
		srv     Server
		wantErr bool
	}{
		{"valid", Server{ServerNames: []string{"example.com"}}, false},
		{"missing server name", Server{}, true},
		{
			"ssl enabled without cert",
			Server{ServerNames: []string{"example.com"}, SSL: SSL{Enabled: true}},
			true,
		},
		{
			"ssl enabled with cert",
			Server{
				ServerNames: []string{"example.com"},
				SSL: SSL{
					Enabled:            true,
					CertificatePath:    "/etc/ssl/cert.pem",
					CertificateKeyPath: "/etc/ssl/key.pem",
				},
			},
			false,
		},
		{
			"http3 without ssl",
			Server{ServerNames: []string{"example.com"}, HTTP3: true},
			true,
		},
		{
			"invalid nested location",
			Server{ServerNames: []string{"example.com"}, Locations: []Location{{}}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srv.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocationValidate(t *testing.T) {
	tests := []struct {
		name    string
		loc     Location
		wantErr bool
	}{
		{"valid root", Location{Path: "/"}, false},
		{"missing path", Location{}, true},
		{"fastcgi without pass", Location{Path: "/", FastCGI: FastCGI{Enabled: true}}, true},
		{"fastcgi with pass", Location{Path: "/", FastCGI: FastCGI{Enabled: true, Pass: "127.0.0.1:9000"}}, false},
		{
			"fastcgi cache without fastcgi",
			Location{Path: "/", FastCGICache: FastCGICache{Enabled: true, ZoneName: "cache"}},
			true,
		},
		{
			"fastcgi cache without zone name",
			Location{
				Path:         "/",
				FastCGI:      FastCGI{Enabled: true, Pass: "127.0.0.1:9000"},
				FastCGICache: FastCGICache{Enabled: true},
			},
			true,
		},
		{"rate limit without zone", Location{Path: "/", RateLimit: RateLimit{Enabled: true}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loc.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func TestLocationValidate_FastCGICacheRequiresZoneDefinition(t *testing.T) {
	loc := Location{
		Path: "/",
		FastCGI: FastCGI{Enabled: true, Pass: "127.0.0.1:9000"},
		FastCGICache: FastCGICache{Enabled: true, ZoneName: "cache"},
	}
	if err := loc.Validate(); err == nil {
		t.Fatal("expected error for cache without zone path and size, got nil")
	}
}
