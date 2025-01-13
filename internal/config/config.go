package config

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"photos/internal/db"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

// The defaultConfig function generates a default configuration object for the application.
// It initializes fields for development mode, server settings, security tokens, base URLs, and routes.
// If there is an error while generating secure tokens, the function returns the error.
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
		PhotosDir: "./photos_dir",
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
			Event:       "/event",
			Photos:      "/photos",
		},
	}
	return defaultCfg, nil
}

// The Load function reads the application configuration from a YAML file or generates a default configuration.
// It sets up logging, initializes database connections, parses HTML templates, and creates an HTTP client.
// If the configuration file does not exist, the function prompts the user to create a default one.
func Load() Config {
	consoleWriter := zerolog.NewConsoleWriter()
	logFile, err := os.Create("logs")
	if err != nil {
		consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		consoleLogger.Fatal().Err(err).Msg("Could not create log file.")
	}
	logger := zerolog.New(zerolog.MultiLevelWriter(consoleWriter, logFile)).With().Timestamp().Logger()

	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.yml", "Path to the configuration file (default: config.yml)")
	flag.Parse()

	if len(flag.Args()) > 0 {
		flag.Usage()
		logger.Fatal().Msg("Unexpected arguments were provided.")
	}

	if _, err = os.Stat(cfgPath); os.IsNotExist(err) {
		logger.Info().Str("path", cfgPath).Msg("Config file not found.")
		logger.Info().Msg("Would you like to create a default config file? (yes/no): ")

		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(response)

		if strings.HasPrefix(response, "y") {
			if err = createDefaultConfig(cfgPath); err != nil {
				logger.Fatal().Err(err).Msg("Failed to create default config file.")
			}
			logger.Info().Str("path", cfgPath).Msg("Default config file created.")
			logger.Info().Str("path", cfgPath).Msg("You need to specify the database DSN in the created config file.")
		} else {
			logger.Fatal().Msg("Exiting program as no config file was created.")
		}
	}

	logger.Info().Str("path", cfgPath).Msg("Using config file.")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read the config file.")
	}

	cfg := Config{}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse the config file.")
	}
	cfg.Templates, err = template.ParseGlob("assets/templates/*.html")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse HTML templates.")
	}

	if cfg.DevMode.Enabled {
		cfg.DB.DB, err = db.New(cfg.DB.Dev.Username, cfg.DB.Dev.Password, cfg.DB.Dev.Host, cfg.DB.Dev.Port, cfg.DB.Dev.Name, cfg.DB.Dev.Cert, cfg.DB.Dev.MaxOpenConns, cfg.DB.Dev.MaxIdleConns, cfg.DB.Dev.ConnMaxLifetime, false)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to establish a database connection.")
		}
	} else {
		cfg.DB.DB, err = db.New(cfg.DB.Prod.Username, cfg.DB.Prod.Password, cfg.DB.Prod.Host, cfg.DB.Prod.Port, cfg.DB.Prod.Name, cfg.DB.Prod.Cert, cfg.DB.Prod.MaxOpenConns, cfg.DB.Prod.MaxIdleConns, cfg.DB.Prod.ConnMaxLifetime, false)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to establish a database connection.")
		}
	}
	cfg.HttpClient = newHTTPClient(6*time.Second, false, false, false, nil)
	cfg.Security.Session.SecureCookie = securecookie.New(cfg.Security.Session.Secret, nil)
	cfg.Logger = logger

	return cfg
}

// The createDefaultConfig function creates a default configuration file at the given path.
// It serializes the default configuration settings into YAML format and writes them to the specified file.
// If the file cannot be created or written to, the function returns an error.
func createDefaultConfig(path string) error {
	cfg, err := defaultConfig()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("Failed to marshal default configuration: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("Failed to write default configuration to file: %w", err)
	}
	return nil
}
