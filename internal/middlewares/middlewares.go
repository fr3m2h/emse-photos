package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/http"
	"photos/internal/handlers"
)

func MaxBodySize(size int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r2 := *r
			r2.Body = http.MaxBytesReader(w, r.Body, size)
			next.ServeHTTP(w, &r2)
		})
	}
}

// func AuthRestricted(cfg handlers.Config) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			sessionToken, err := r.Cookie(cfg.Security.Session.CookieName)
// 			if err != nil {
// 				redirectToLogin(w, r, cfg)
// 				return
// 			}
// 			splitted := strings.SplitN(sessionToken.String(), ":", 2)
// 			if len(splitted) != 2 {
// 				redirectToLogin(w, r, cfg)
// 				return
// 			}
// 			if !validateMAC([]byte(splitted[0]), []byte(splitted[1]), cfg.Security.Session.Secret) {
// 				redirectToLogin(w, r, cfg)
// 				return
// 			}
// 			if exists, err := cfg.DB.DoesSessionExist(r.Context(), splitted[0]); err == nil || !exists {
// 				redirectToLogin(w, r, cfg)
// 				return
// 			}
// 			// TO DO: ADD THE USER VALUE IN REQUEST CONTEXT
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

//	func DevRestricted(cfg handlers.Config) func(http.Handler) http.Handler {
//		return func(next http.Handler) http.Handler {
//			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//				devToken, err := r.Cookie(cfg.Security.Session.CookieName)
//				if err != nil {
//					redirectToLogin(w, r, cfg)
//					return
//				}
//				splitted := strings.SplitN(devToken.String(), ":", 2)
//				if len(splitted) != 2 {
//					redirectToLogin(w, r, cfg)
//					return
//				}
//				if !validateMAC([]byte(splitted[0]), []byte(splitted[1]), cfg.Security.Session.Secret) {
//					redirectToLogin(w, r, cfg)
//					return
//				}
//				if exists, err := cfg.DB.DoesSessionExist(r.Context(), splitted[0]); err == nil || !exists {
//					redirectToLogin(w, r, cfg)
//					return
//				}
//				next.ServeHTTP(w, r)
//			})
//		}
//	}
func redirectToLogin(w http.ResponseWriter, r *http.Request, cfg handlers.Config) {
	http.SetCookie(w, &http.Cookie{
		Name:   cfg.Security.Session.CookieName,
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// generateHMAC returns the HMAC of message using secretKey.
func generateMAC(message, secretKey []byte) []byte {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write(message)
	return mac.Sum(nil)
}

// validateHMAC reports whether messageMAC is a valid HMAC tag for message.
// Copied from https://pkg.go.dev/crypto/hmac
func validateMAC(message, messageMAC, secretKey []byte) bool {
	return hmac.Equal(messageMAC, generateMAC(message, secretKey))
}
