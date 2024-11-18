package config

import (
	"html/template"
	"net/http"
	"photos/internal/db"
	"time"

	"github.com/gorilla/securecookie"
)

type Config struct {
	DevMode  DevMode  `yaml:"dev_mode"`
	Server   Server   `yaml:"server"`
	Security Security `yaml:"security"`
	DB       DB       `yaml:"db"`
	BaseURLs BaseURLs `yaml:"base_urls"`
	Routes   Routes   `yaml:"routes"`

	HttpClient *http.Client       `yaml:"-"`
	Templates  *template.Template `yaml:"-"`
}

type DevMode struct {
	Enabled        bool          `yaml:"enabled"`
	Username       string        `yaml:"username"`
	Password       string        `yaml:"password"`
	Secret         secretKey     `yaml:"secret"`
	CookieName     string        `yaml:"cookie_name"`
	CookieMaxAge   time.Duration `yaml:"cookie_max_age"`
	CookieSecure   bool          `yaml:"cookie_secure"`
	CookieHTTPOnly bool          `yaml:"cookie_http_only"`
	CookieSameSite http.SameSite `yaml:"cookie_same_site"`
}
type Server struct {
	Port                  int           `yaml:"port"`
	ReadTimeout           time.Duration `yaml:"read_timeout"`
	WriteTimeout          time.Duration `yaml:"write_timeout"`
	IdleTimeout           time.Duration `yaml:"idle_timeout"`
	RequestContextTimeout time.Duration `yaml:"request_context_timeout"`
	MaxHeaderBytes        int           `yaml:"max_header_bytes"`
	MaxBodySize           int64         `yaml:"max_body_size"`
}

type Token struct {
	Secret         secretKey     `yaml:"secret"`
	CookieName     string        `yaml:"cookie_name"`
	CookieMaxAge   time.Duration `yaml:"cookie_max_age"`
	CookieSecure   bool          `yaml:"cookie_secure"`
	CookieHTTPOnly bool          `yaml:"cookie_http_only"`
	CookieSameSite http.SameSite `yaml:"cookie_same_site"`
}

type CsrfToken struct {
	Token
	FieldName  string `yaml:"field_name"`
	HeaderName string `yaml:"header_name"`
}

type SessionToken struct {
	Token
	SecureCookie *securecookie.SecureCookie `yaml:"-"`
}

type Security struct {
	Csrf    CsrfToken    `yaml:"csrf"`
	Session SessionToken `yaml:"session"`
}

type DSN struct {
	Name            string        `yaml:"name"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Port            string        `yaml:"port"`
	Host            string        `yaml:"host"`
	Cert            string        `yaml:"cert"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type DB struct {
	*db.DB `yaml:"-"`
	Dev    DSN `yaml:"dev"`
	Prod   DSN `yaml:"prod"`
}

type Routes struct {
	Favicon     string `yaml:"favicon"`
	Landing     string `yaml:"landing"`
	Login       string `yaml:"login"`
	CasCallback string `yaml:"cas_callback"`

	Dashboard string `yaml:"dashboard"`
	Logout    string `yaml:"logout"`
}

type BaseURL struct {
	Service string `yaml:"service"`
	Cas     string `yaml:"cas"`
}

type BaseURLs struct {
	Dev  BaseURL `yaml:"dev"`
	Prod BaseURL `yaml:"prod"`
}
