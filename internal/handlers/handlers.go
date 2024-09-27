package handlers

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func respondWithMessage(w http.ResponseWriter, error string, status int) {
	if status >= 500 {
		log.Printf("5xx error: %s", error)
		http.Error(w, "Internal server error", status)
		return
	}
	http.Error(w, error, status)
}

func renderTemplate(w http.ResponseWriter, t *template.Template, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		respondWithMessage(w, fmt.Sprintf("error executing template: %v", err), http.StatusInternalServerError)
		return
	}
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
	cfg.Templates, err = template.ParseGlob("internal/templates/*.html")
	if err != nil {
		err = fmt.Errorf("error parsing templates: %v", err)
		return
	}
	return
}

type Config struct {
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
			CookieSameSite string        `yaml:"cookie_same_site"`
		} `yaml:"csrf_token"`
	} `yaml:"security"`
	Templates *template.Template
}

type secretKey []byte

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
