package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	defaultPort   int
	defaultTicket string
)

func main() {
	flag.IntVar(&defaultPort, "port", 3000, "Port to run the server on")
	flag.StringVar(&defaultTicket, "ticket", "ST-12345", "Default mock ticket")
	flag.Parse()

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	// Routes
	r.Get("/cas/login", casLoginHandler)
	r.Get("/cas/serviceValidate", casServiceValidateHandler)

	addr := fmt.Sprintf("127.0.0.1:%d", defaultPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	fmt.Printf("Mock CAS server running on: %s\n", addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}

// Mock CAS login endpoint
func casLoginHandler(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	if service != "" {
		http.Redirect(w, r, fmt.Sprintf("%s?ticket=%s", service, defaultTicket), http.StatusFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Mock CAS login page. Use /cas/login?service=<service-url> to log in."))
}

// Mock CAS serviceValidate endpoint
func casServiceValidateHandler(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("ticket")
	service := r.URL.Query().Get("service")

	if ticket == defaultTicket && service != "" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationSuccess>
    <cas:user>jdoe</cas:user>
    <cas:attributes>
      <cas:cn>John Doe</cas:cn>
      <cas:email>jdoe@example.com</cas:email>
      <cas:departmentNumber>ICM 2A</cas:departmentNumber>
      <cas:businessCategory>ELEVE</cas:businessCategory>
    </cas:attributes>
  </cas:authenticationSuccess>
</cas:serviceResponse>
		`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(`
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationFailure code="INVALID_TICKET">
    Ticket %s is not recognized
  </cas:authenticationFailure>
</cas:serviceResponse>
	`, ticket)))
}
