package config

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"photos/pkg/db"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

// defaultConfig generates and returns a default configuration object.
//
// The function creates default values for various fields in the Config struct,
// including default ports, timeouts, and security settings like CSRF and session tokens.
//
// Returns:
//   - Config: The default configuration object.
//   - error: An error if secure token generation fails.
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

// Load loads the application configuration from a YAML file.
//
// If the configuration file does not exist, it prompts the user to create a default one.
// It also sets up logging, parses HTML templates, and initializes the database connection
// and HTTP client based on the configuration.
//
// Returns:
//   - Config: The application configuration object populated with values from the YAML file
//     or generated defaults.
func Load() Config {
	consoleWriter := zerolog.NewConsoleWriter()
	logFile, err := os.Create("logs")
	if err != nil {
		consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		consoleLogger.Fatal().Err(err).Msg("could not create log file")
	}
	logger := zerolog.New(zerolog.MultiLevelWriter(consoleWriter, logFile)).With().Timestamp().Logger()

	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.yml", "Path to the configuration file (default: config.yml)")
	flag.Parse()

	if len(flag.Args()) > 0 {
		flag.Usage()
		logger.Fatal().Msg("unexpected arguments were given")
	}

	// Check if the config file exists
	if _, err = os.Stat(cfgPath); os.IsNotExist(err) {
		logger.Info().Str("path", cfgPath).Msg("config file not found")
		logger.Info().Msg("would you like to create a default config file? (yes/no): ")

		var response string
		_, _ = fmt.Scanln(&response)
		response = strings.ToLower(response)

		if strings.HasPrefix(response, "y") {
			if err = createDefaultConfig(cfgPath); err != nil {
				logger.Fatal().Err(err).Msg("failed to create default config file")
			}
			logger.Info().Str("path", cfgPath).Msg("default config file created")
			logger.Info().Str("path", cfgPath).Msg("you now have to specify the database DSN in the created config file")
		} else {
			logger.Fatal().Msg("exiting program, no config file created")
		}
	}

	logger.Info().Str("path", cfgPath).Msg("using config file")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read config file")
	}

	cfg := Config{}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to unmarshal config file")
	}
	cfg.Templates, err = template.ParseGlob("assets/templates/*.html")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse html templates")
	}

	if cfg.DevMode.Enabled {
		cfg.DB.DB, err = db.New(cfg.DB.Dev.Username, cfg.DB.Dev.Password, cfg.DB.Dev.Host, cfg.DB.Dev.Port, cfg.DB.Dev.Name, cfg.DB.Dev.Cert, cfg.DB.Dev.MaxOpenConns, cfg.DB.Dev.MaxIdleConns, cfg.DB.Dev.ConnMaxLifetime, false)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to create database connection")
		}
	} else {
		cfg.DB.DB, err = db.New(cfg.DB.Prod.Username, cfg.DB.Prod.Password, cfg.DB.Prod.Host, cfg.DB.Prod.Port, cfg.DB.Prod.Name, cfg.DB.Prod.Cert, cfg.DB.Prod.MaxOpenConns, cfg.DB.Prod.MaxIdleConns, cfg.DB.Prod.ConnMaxLifetime, false)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to create database connection")
		}
	}
	cfg.HttpClient = newHTTPClient(6*time.Second, false, false, false, nil)
	cfg.Security.Session.SecureCookie = securecookie.New(cfg.Security.Session.Secret, nil)
	cfg.Logger = logger

	return cfg
}

// createDefaultConfig generates a default configuration file at the specified path.
//
// Parameters:
//   - path: The file path where the default configuration file should be created.
//
// Returns:
//   - error: An error if the configuration could not be created or written to the file.
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
