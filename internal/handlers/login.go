package handlers

import (
	"fmt"
	"net/http"
)

func (cfg Config) CasCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ticket from the request
	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		respondWithMessage(w, "Ticket is missing", http.StatusBadRequest)
		return
	}

	// Validate the ticket with the CAS server
	userID, err := validateCASTicket(ticket)
	if err != nil {
		respondWithMessage(w, "Ticket validation failed", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "User %s is authenticated", userID)
}

func validateCASTicket(ticket string) (string, error) {
	// validateURL := fmt.Sprintf("%s/serviceValidate?service=%s&ticket=%s", casBaseURL, url.QueryEscape(serviceURL), url.QueryEscape(ticket))
	//
	// // Send the validation request to the CAS server
	// resp, err := http.Get(validateURL)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to validate ticket: %w", err)
	// }
	// defer resp.Body.Close()
	//
	// if resp.StatusCode != http.StatusOK {
	// 	return "", fmt.Errorf("ticket validation failed with status: %d", resp.StatusCode)
	// }
	//
	// // Parse the CAS response to extract the user ID
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read response body: %w", err)
	// }
	//
	// // In a real application, parse the XML response to get the userID
	// // Here we'll just simulate extracting the user ID
	// userID := parseUserIDFromCASResponse(body)
	// return userID, nil
	return "", nil
}
