package routes

import (
	"net/http"
	"photos/internal/handlers"
	"photos/internal/middlewares"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/gorilla/csrf"
)

func Service(cfg handlers.Config) http.Handler {
	r := chi.NewRouter()
	loadGlobalMiddlewares(r, cfg)

	r.Group(func(r chi.Router) {

		r.Get("/", cfg.ServeLandingPageHandler)

		// r.Post("/login_first_stage", cfg.HandlerLoginUserFirstStage)
		// r.Get("/refresh", cfg.HandlerRefreshTokens)

		// r.Group(func(r chi.Router) {
		// 	r.Use(AuthRestricted(cfg))
		// 	r.Get("/logout", cfg.HandlerLogoutUser)
		// 	r.Post("/change_password", cfg.HandlerUpdateUserPasswordWithOldPassword)
		//
		// })
	})
	return r
}

func loadGlobalMiddlewares(r *chi.Mux, cfg handlers.Config) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.AllowContentEncoding("gzip", "deflate", "gzip/deflate", "deflate/gzip"))
	r.Use(middleware.AllowContentType("application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.Compress(4, "application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.CleanPath)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.Server.RequestContextTimeout))
	r.Use(middlewares.MaxBodySize(cfg.Server.MaxBodySize))
	r.Use(csrf.Protect(
		cfg.Security.CsrfToken.Secret,
		csrf.MaxAge(int(cfg.Security.CsrfToken.CookieMaxAge.Seconds())),
		csrf.HttpOnly(cfg.Security.CsrfToken.CookieHTTPOnly),
		csrf.Secure(cfg.Security.CsrfToken.CookieSecure),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.RequestHeader(cfg.Security.CsrfToken.HeaderName),
		csrf.FieldName(cfg.Security.CsrfToken.FieldName),
		csrf.CookieName(cfg.Security.CsrfToken.CookieName),
	))
	// Rate limiter for all routes
	r.Use(httprate.Limit(
		30,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}),
	))
}
