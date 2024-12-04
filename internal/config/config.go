package config

import (
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"photos/internal/db"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"gopkg.in/yaml.v3"
)

func defaultConfig() (Config, error) {
	s1, err := generateSecureHex(16)
	if err != nil {
		return Config{}, err
	}
	s2, err := generateSecureHex(16)
	if err != nil {
		return Config{}, err
	}

	defaultCfg := Config{
		DevMode: DevMode{
			Enabled: true,
		},
		Server: Server{
			Port:                  8080,
			ReadTimeout:           6 * time.Second,
			WriteTimeout:          12 * time.Second,
			RequestContextTimeout: 12 * time.Second,
			IdleTimeout:           30 * time.Second,
			MaxHeaderBytes:        1024 * 4,
			MaxBodySize:           1024,
		},
		Security: Security{
			Csrf: CsrfToken{
				Token: Token{
					Secret:         s1,
					CookieName:     "csrf_token",
					CookieMaxAge:   10 * time.Minute,
					CookieSecure:   true,
					CookieHTTPOnly: true,
					CookieSameSite: http.SameSiteStrictMode,
				},
				FieldName:  "csrf_token",
				HeaderName: "X-CSRF-TOKEN",
			},
			Session: SessionToken{
				Token: Token{
					Secret:         s2,
					CookieName:     "session_token",
					CookieMaxAge:   time.Hour,
					CookieSecure:   true,
					CookieHTTPOnly: true,
					CookieSameSite: http.SameSiteStrictMode,
				},
				SecureCookie: securecookie.New(s2, nil),
			},
		},
		BaseURLs: BaseURLs{
			Dev: BaseURL{
				Service: "http://127.0.0.1:8888",
				Cas:     "http://127.0.0.1:3000/cas",
			},
			Prod: BaseURL{
				Service: "https://portail-etu.emse.fr/photos",
				Cas:     "https://cas.emse.fr",
			},
		},
		Routes: Routes{
			Favicon:     "/favicon.ico",
			Landing:     "/",
			Login:       "/login",
			CasCallback: "/cas",
			Dashboard:   "/dashboard",
			Logout:      "/logout",
		},
	}

	return defaultCfg, nil
}

func Load(logger *slog.Logger) (cfg Config, err error) {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.yml", "Path to the configuration file (default: config.yml)")
	flag.Parse()

	if len(flag.Args()) > 0 {
		flag.Usage()
		err = fmt.Errorf("unexpected arguments were provided")
		return
	}

	// Check if the config file exists
	if _, err = os.Stat(cfgPath); os.IsNotExist(err) {
		fmt.Printf(">Config file not found at %s\n", cfgPath)
		fmt.Print(">Would you like to create a default config file? (yes/no): ")
		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(response)

		if response == "yes" || response == "y" {
			if err = createDefaultConfig(cfgPath); err != nil {
				err = fmt.Errorf("failed to create default config file: %v\n", err)
				return
			}
			logger.Info("Default config file created", "path", cfgPath)
			logger.Info("You now have to specify the database DSN in the created config file", "path", cfgPath)
		} else {
			err = fmt.Errorf("exiting program, no config file created")
			return
		}
	}

	logger.Info("using config file", "path", cfgPath)
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

	return cfg, nil
}

func createDefaultConfig(path string) error {
	cfg, err := defaultConfig()
	if err != nil {
		return err
	}

	// Marshal the default config into YAML
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Write the YAML data to the specified file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write default config to file: %w", err)
	}
	return nil
}
