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
	// Parse command-line arguments
	flag.IntVar(&defaultPort, "port", 3000, "Port to run the server on")
	flag.StringVar(&defaultTicket, "ticket", "ST-12345", "Default mock ticket")
	flag.Parse()

	// Create a new Chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)                    // Log each request
	r.Use(middleware.Recoverer)                 // Recover from panics to prevent server crashes
	r.Use(middleware.Timeout(10 * time.Second)) // Set request timeout

	// Routes
	r.Get("/cas/login", casLoginHandler)
	r.Get("/cas/serviceValidate", casServiceValidateHandler)

	// Start the server
	addr := fmt.Sprintf(":%d", defaultPort)
	fmt.Printf("Mock CAS server running on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
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
	w.Write([]byte("Mock CAS login page. Use /cas/login?service=<service-url> to log in."))
}

// Mock CAS serviceValidate endpoint
func casServiceValidateHandler(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("ticket")
	service := r.URL.Query().Get("service")

	if ticket == defaultTicket && service != "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
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
	w.Write([]byte(fmt.Sprintf(`
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationFailure code="INVALID_TICKET">
    Ticket %s is not recognized
  </cas:authenticationFailure>
</cas:serviceResponse>
	`, ticket)))
}
