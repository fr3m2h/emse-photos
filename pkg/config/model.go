package config

import (
	"html/template"
	"net/http"
	"photos/pkg/db"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog"
)

// Config represents the main configuration structure for the application.
// It includes settings for development mode, server, security, database, base URLs, and routes.
type Config struct {
	DevMode  DevMode  `yaml:"dev_mode"`  // Development mode settings.
	Server   Server   `yaml:"server"`    // Server-related configuration.
	Security Security `yaml:"security"`  // Security settings such as CSRF and session tokens.
	DB       DB       `yaml:"db"`        // Database connection details for development and production.
	BaseURLs BaseURLs `yaml:"base_urls"` // URLs for different environments (Dev and Prod).
	Routes   Routes   `yaml:"routes"`    // Application route paths.

	HttpClient *http.Client       `yaml:"-"` // HTTP client instance (excluded from YAML).
	Templates  *template.Template `yaml:"-"` // Parsed HTML templates (excluded from YAML).
	Logger     zerolog.Logger     `yaml:"-"` // Logger instance (excluded from YAML).
}

// DevMode contains the configuration for development mode.
type DevMode struct {
	Enabled bool `yaml:"enabled"` // Indicates if development mode is enabled.
}

// Server holds the configuration for the HTTP server.
type Server struct {
	Port                  int           `yaml:"port"`                    // Server port.
	ReadTimeout           time.Duration `yaml:"read_timeout"`            // Maximum duration for reading requests.
	WriteTimeout          time.Duration `yaml:"write_timeout"`           // Maximum duration for writing responses.
	IdleTimeout           time.Duration `yaml:"idle_timeout"`            // Maximum duration for keeping idle connections.
	RequestContextTimeout time.Duration `yaml:"request_context_timeout"` // Context timeout for requests.
	MaxHeaderBytes        int           `yaml:"max_header_bytes"`        // Maximum size of request headers.
	MaxBodySize           int64         `yaml:"max_body_size"`           // Maximum size of request bodies.
}

// Token represents a base token configuration for CSRF and session tokens.
type Token struct {
	Secret         secretKey     `yaml:"secret"`           // The secret key used for token generation.
	CookieName     string        `yaml:"cookie_name"`      // Name of the token's cookie.
	CookieMaxAge   time.Duration `yaml:"cookie_max_age"`   // Maximum age of the cookie.
	CookieSecure   bool          `yaml:"cookie_secure"`    // Whether the cookie requires a secure connection.
	CookieHTTPOnly bool          `yaml:"cookie_http_only"` // Whether the cookie is HTTP-only.
	CookieSameSite http.SameSite `yaml:"cookie_same_site"` // SameSite policy for the cookie.
}

// CsrfToken represents the configuration for CSRF tokens.
type CsrfToken struct {
	Token
	FieldName  string `yaml:"field_name"`  // Name of the hidden form field for CSRF tokens.
	HeaderName string `yaml:"header_name"` // Name of the HTTP header for CSRF tokens.
}

// SessionToken represents the configuration for session tokens.
type SessionToken struct {
	Token
	SecureCookie *securecookie.SecureCookie `yaml:"-"` // SecureCookie instance for session handling (excluded from YAML).
}

// Security holds the security-related configurations such as CSRF and session tokens.
type Security struct {
	Csrf    CsrfToken    `yaml:"csrf"`    // CSRF token configuration.
	Session SessionToken `yaml:"session"` // Session token configuration.
}

// DSN represents the Data Source Name (DSN) configuration for database connections.
type DSN struct {
	Name            string        `yaml:"name"`              // Database name.
	Username        string        `yaml:"username"`          // Database username.
	Password        string        `yaml:"password"`          // Database password.
	Port            string        `yaml:"port"`              // Port of the database server.
	Host            string        `yaml:"host"`              // Hostname or IP of the database server.
	Cert            string        `yaml:"cert"`              // TLS certificate path.
	MaxIdleConns    int           `yaml:"max_idle_conns"`    // Maximum number of idle connections.
	MaxOpenConns    int           `yaml:"max_open_conns"`    // Maximum number of open connections.
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"` // Maximum lifetime of a single connection.
}

// DB represents the database configuration for development and production environments.
type DB struct {
	*db.DB `yaml:"-"` // Embedded DB instance for database operations (excluded from YAML).
	Dev    DSN        `yaml:"dev"`  // Development database configuration.
	Prod   DSN        `yaml:"prod"` // Production database configuration.
}

// Routes contains the paths for various application routes.
type Routes struct {
	Favicon     string `yaml:"favicon"`      // Path to the favicon.
	Landing     string `yaml:"landing"`      // Path to the landing page.
	Login       string `yaml:"login"`        // Path to the login page.
	CasCallback string `yaml:"cas_callback"` // Path to the CAS callback.
	Dashboard   string `yaml:"dashboard"`    // Path to the user dashboard.
	Logout      string `yaml:"logout"`       // Path to the logout page.
}

// BaseURL represents the configuration for a set of URLs.
type BaseURL struct {
	Service string `yaml:"service"` // Base URL for the service.
	Cas     string `yaml:"cas"`     // Base URL for the CAS server.
}

// BaseURLs contains the service and CAS URLs for development and production environments.
type BaseURLs struct {
	Dev  BaseURL `yaml:"dev"`  // Development environment URLs.
	Prod BaseURL `yaml:"prod"` // Production environment URLs.
}
