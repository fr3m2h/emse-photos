package handlers

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"photos/internal/db"
	"time"

	"github.com/gorilla/securecookie"
	"gopkg.in/yaml.v3"
)

type secretKey []byte

type Config struct {
	DevMode struct {
		Enabled        bool          `yaml:"enabled"`
		Username       string        `yaml:"username"`
		Password       string        `yaml:"password"`
		Secret         secretKey     `yaml:"secret"`
		CookieName     string        `yaml:"cookie_name"`
		CookieMaxAge   time.Duration `yaml:"cookie_max_age"`
		CookieSecure   bool          `yaml:"cookie_secure"`
		CookieHTTPOnly bool          `yaml:"cookie_http_only"`
		CookieSameSite http.SameSite `yaml:"cookie_same_site"`
	} `yaml:"dev_mode"`
	Server struct {
		Port                  int           `yaml:"port"`
		ReadTimeout           time.Duration `yaml:"read_timeout"`
		WriteTimeout          time.Duration `yaml:"write_timeout"`
		IdleTimeout           time.Duration `yaml:"idle_timeout"`
		RequestContextTimeout time.Duration `yaml:"request_context_timeout"`
		MaxHeaderBytes        int           `yaml:"max_header_bytes"`
		MaxBodySize           int64         `yaml:"max_body_size"`
	} `yaml:"server"`
	Security struct {
		CsrfToken struct {
			Secret         secretKey     `yaml:"secret"`
			FieldName      string        `yaml:"field_name"`
			HeaderName     string        `yaml:"header_name"`
			CookieName     string        `yaml:"cookie_name"`
			CookieMaxAge   time.Duration `yaml:"cookie_max_age"`
			CookieSecure   bool          `yaml:"cookie_secure"`
			CookieHTTPOnly bool          `yaml:"cookie_http_only"`
			CookieSameSite http.SameSite `yaml:"cookie_same_site"`
		} `yaml:"csrf_token"`
		Session struct {
			Secret         secretKey `yaml:"secret"`
			SecureCookie   *securecookie.SecureCookie
			CookieName     string        `yaml:"cookie_name"`
			CookieMaxAge   time.Duration `yaml:"cookie_max_age"`
			CookieSecure   bool          `yaml:"cookie_secure"`
			CookieHTTPOnly bool          `yaml:"cookie_http_only"`
			CookieSameSite http.SameSite `yaml:"cookie_same_site"`
		} `yaml:"session"`
	} `yaml:"security"`
	DB struct {
		*db.DB
		Prod struct {
			Name            string        `yaml:"name"`
			Username        string        `yaml:"username"`
			Password        string        `yaml:"password"`
			Port            string        `yaml:"port"`
			Host            string        `yaml:"host"`
			Cert            string        `yaml:"cert"`
			MaxIdleConns    int           `yaml:"max_idle_conns"`
			MaxOpenConns    int           `yaml:"max_open_conns"`
			ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
		} `yaml:"prod"`
		Dev struct {
			Name            string        `yaml:"name"`
			Username        string        `yaml:"username"`
			Password        string        `yaml:"password"`
			Port            string        `yaml:"port"`
			Host            string        `yaml:"host"`
			Cert            string        `yaml:"cert"`
			MaxIdleConns    int           `yaml:"max_idle_conns"`
			MaxOpenConns    int           `yaml:"max_open_conns"`
			ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
		} `yaml:"dev"`
	} `yaml:"db"`
	BaseUrls struct {
		Prod struct {
			Service string `yaml:"service"`
			Cas     string `yaml:"cas"`
		} `yaml:"prod"`
		Dev struct {
			Service string `yaml:"service"`
			Cas     string `yaml:"cas"`
		} `yaml:"dev"`
	} `yaml:"base_urls"`
	Routes struct {
		Landing     string `yaml:"landing"`
		Login       string `yaml:"login"`
		CasCallback string `yaml:"cas_callback"`

		Dashboard string `yaml:"dashboard"`
		Logout    string `yaml:"logout"`
	} `yaml:"routes"`

	HttpClient *http.Client
	Templates  *template.Template
}

func New() (cfg Config, err error) {
	var cfgPath string

	if n := len(os.Args); n == 1 {
		cfgPath = "config.yml"
	} else if n == 2 {
		cfgPath = os.Args[1]
	} else {
		err = fmt.Errorf("program only needs one argument")
		return
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		err = fmt.Errorf("error reading config file: %v", err)
		return
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		err = fmt.Errorf("error unmarshalling config file: %v", err)
		return
	}
	cfg.Templates, err = template.ParseGlob("assets/templates/*.html")
	if err != nil {
		err = fmt.Errorf("error parsing templates: %v", err)
		return
	}
	if cfg.DevMode.Enabled {
		cfg.DB.DB, err = db.New(cfg.DB.Dev.Username, cfg.DB.Dev.Password, cfg.DB.Dev.Host, cfg.DB.Dev.Port, cfg.DB.Dev.Name, cfg.DB.Dev.Cert, cfg.DB.Dev.MaxOpenConns, cfg.DB.Dev.MaxIdleConns, cfg.DB.Dev.ConnMaxLifetime, false)
		if err != nil {
			err = fmt.Errorf("error creating database connection: %v", err)
			return
		}
	} else {
		cfg.DB.DB, err = db.New(cfg.DB.Prod.Username, cfg.DB.Prod.Password, cfg.DB.Prod.Host, cfg.DB.Prod.Port, cfg.DB.Prod.Name, cfg.DB.Prod.Cert, cfg.DB.Prod.MaxOpenConns, cfg.DB.Prod.MaxIdleConns, cfg.DB.Prod.ConnMaxLifetime, false)
		if err != nil {
			err = fmt.Errorf("error creating database connection: %v", err)
			return
		}

	}
	cfg.HttpClient = newHTTPClient(6*time.Second, false, false, false, nil)
	cfg.Security.Session.SecureCookie = securecookie.New(cfg.Security.Session.Secret, nil)

	return
}

func (k secretKey) MarshalYAML() (interface{}, error) {
	return hex.EncodeToString(k), nil
}

func (k *secretKey) UnmarshalYAML(node *yaml.Node) error {
	value := node.Value
	ba, err := hex.DecodeString(value)
	if err != nil {
		return err
	}
	*k = ba
	return nil
}

func newHTTPClient(requestTimeout time.Duration, useCookie, disableKeepAlive, disableCompression bool, proxyFunc func(*http.Request) (*url.URL, error)) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 10 ^ 9,
		MaxConnsPerHost:     0,
		IdleConnTimeout:     10 * time.Second,
		DisableCompression:  disableCompression,
		DisableKeepAlives:   disableKeepAlive,
	}
	if proxyFunc == nil {
		transport.Proxy = http.ProxyFromEnvironment
	}
	client := http.Client{
		Timeout:   requestTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if useCookie {
		jar, _ := cookiejar.New(nil)
		client.Jar = jar
	}
	return &client
}
