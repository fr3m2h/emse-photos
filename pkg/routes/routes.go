package routes

import (
	"net/http"
	"photos/pkg/handlers"
	"photos/pkg/middlewares"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/gorilla/csrf"
	"github.com/rs/zerolog/hlog"
)

func Service(cfg handlers.Config) http.Handler {
	r := chi.NewRouter()
	loadGlobalMiddlewares(r, cfg)

	r.NotFound(cfg.ServeNotFoundHandler)
	r.Get(cfg.Routes.Favicon, handlers.ServeFaviconHandler)
	r.Get(cfg.Routes.Landing, cfg.ServeLandingHandler)

	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			10,
			time.Minute,
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
			}),
		))
		r.Get(cfg.Routes.Login, cfg.LoginHandler)
		r.Get(cfg.Routes.CasCallback, cfg.CasCallbackHandler)
	})
	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthRestricted(cfg))
		r.Get(cfg.Routes.Dashboard, cfg.ServeDashboardHandler)
		r.Get(cfg.Routes.Logout, cfg.LogoutHandler)
	})
	return r
}

func loadGlobalMiddlewares(r *chi.Mux, cfg handlers.Config) {
	r.Use(hlog.RemoteAddrHandler("ip"), hlog.UserAgentHandler("ua"), hlog.RefererHandler("referer"), hlog.RequestIDHandler("req-id", "X-Request-Id"))
	r.Use(hlog.NewHandler(cfg.Logger))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.AllowContentEncoding("gzip", "deflate", "gzip/deflate", "deflate/gzip"))
	r.Use(middleware.AllowContentType("application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.CleanPath, middleware.RedirectSlashes)
	r.Use(middleware.Compress(4, "application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.Timeout(cfg.Server.RequestContextTimeout))
	r.Use(middlewares.MaxBodySize(cfg.Server.MaxBodySize))
	r.Use(csrf.Protect(
		cfg.Security.Csrf.Secret,
		csrf.MaxAge(int(cfg.Security.Csrf.CookieMaxAge.Seconds())),
		csrf.HttpOnly(cfg.Security.Csrf.CookieHTTPOnly),
		csrf.Secure(cfg.Security.Csrf.CookieSecure),
		csrf.SameSite(csrf.SameSiteMode(cfg.Security.Csrf.CookieSameSite)),
		csrf.RequestHeader(cfg.Security.Csrf.HeaderName),
		csrf.FieldName(cfg.Security.Csrf.FieldName),
		csrf.CookieName(cfg.Security.Csrf.CookieName),
	))
	r.Use(httprate.Limit(
		60,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}),
	))
}
