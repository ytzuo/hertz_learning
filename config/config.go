package config

import "time"

type Config struct {
	Service   ServiceConfig
	HTTP      HTTPConfig
	Auth      AuthConfig
	RateLimit RateLimitConfig
}

type ServiceConfig struct {
	Name    string
	Version string
	Env     string
}

type HTTPConfig struct {
	Addr string
}

type AuthConfig struct {
	JWTSecret      string
	AccessTokenTTL time.Duration
	Issuer         string
}

type RateLimitConfig struct {
	PublicMaxRequests  int
	PrivateMaxRequests int
	OrderMaxRequests   int
	Window             time.Duration
}

func Load() Config {
	return Config{
		Service: ServiceConfig{
			Name:    "commerce-api",
			Version: "v1",
			Env:     "local",
		},
		HTTP: HTTPConfig{
			Addr: ":8888",
		},
		Auth: AuthConfig{
			JWTSecret:      "dev-secret-change-me",
			AccessTokenTTL: 2 * time.Hour,
			Issuer:         "commerce-api",
		},
		RateLimit: RateLimitConfig{
			PublicMaxRequests:  120,
			PrivateMaxRequests: 60,
			OrderMaxRequests:   10,
			Window:             time.Minute,
		},
	}
}
