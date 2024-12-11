package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"photos/pkg/db/query"
	"time"
)

type casResponse struct {
	XMLName               xml.Name               `xml:"http://www.yale.edu/tp/cas serviceResponse"`
	AuthenticationSuccess *authenticationSuccess `xml:"authenticationSuccess"`
	AuthenticationFailure *authenticationFailure `xml:"authenticationFailure"`
}

type authenticationSuccess struct {
	User       string     `xml:"user"`
	Attributes attributes `xml:"attributes"`
}

type attributes struct {
	CN               string `xml:"cn"`
	Email            string `xml:"email"`
	DepartmentNumber string `xml:"departmentNumber"`
	BusinessCategory string `xml:"businessCategory"`
}

type authenticationFailure struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

func (cfg Config) LoginHandler(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	casLoginUrlWithCallback := ""
	if cfg.DevMode.Enabled {
		params.Add("service", fmt.Sprintf("%s%s", cfg.BaseURLs.Dev.Service, cfg.Routes.CasCallback))
		casLoginUrlWithCallback = fmt.Sprintf("%s/login?%s", cfg.BaseURLs.Dev.Cas, params.Encode())
	} else {
		params.Add("service", fmt.Sprintf("%s%s", cfg.BaseURLs.Prod.Service, cfg.Routes.CasCallback))
		casLoginUrlWithCallback = fmt.Sprintf("%s/login?%s", cfg.BaseURLs.Prod.Cas, params.Encode())
	}

	cookie, err := r.Cookie(cfg.Security.Session.CookieName)
	if err != nil {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	var data map[string]string
	err = cfg.Security.Session.SecureCookie.Decode(cfg.Security.Session.CookieName, cookie.Value, &data)
	if err != nil {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	sessionToken, ok := data[cfg.Security.Session.CookieName]
	if !ok || sessionToken == "" {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	session, err := cfg.DB.GetSessionWithToken(r.Context(), sessionToken)
	if err != nil {
		log.Printf("DB Failure: %v", err)
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	if session.CreationDate.Add(cfg.Security.Session.CookieMaxAge).Before(time.Now()) {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	http.Redirect(w, r, cfg.Routes.Dashboard, http.StatusFound)
}

func (cfg Config) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cookie := &http.Cookie{
		Name:   cfg.Security.Session.CookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	err := cfg.DB.DeleteSessionWithToken(r.Context(), ctx.Value(cfg.Security.Session.CookieName).(string))
	if err != nil {
		log.Printf("DB Failure: %v", err)
	}
	http.Redirect(w, r, cfg.Routes.Landing, http.StatusFound)
}

func (cfg Config) CasCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		RespondWithMessage(w, "Ticket is missing", http.StatusBadRequest)
		return
	}

	//Now we have to validate the ticket with the CAS server
	validationURL := ""
	if cfg.DevMode.Enabled {
		validationURL = fmt.Sprintf("%s/serviceValidate?service=%s&ticket=%s", cfg.BaseURLs.Dev.Cas, url.QueryEscape(cfg.BaseURLs.Dev.Service), url.QueryEscape(ticket))
	} else {
		validationURL = fmt.Sprintf("%s/serviceValidate?service=%s&ticket=%s", cfg.BaseURLs.Prod.Cas, url.QueryEscape(cfg.BaseURLs.Prod.Service), url.QueryEscape(ticket))
	}

	resp, err := cfg.HttpClient.Get(validationURL)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Error while validating CAS ticket: %v", err), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		RespondWithMessage(w, fmt.Sprintf("Error while validating CAS ticket, got a non 200 status code: %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Error while validating CAS ticket: %v", err), http.StatusInternalServerError)
		return
	}
	err = resp.Body.Close()
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Error while validating CAS ticket: %v", err), http.StatusInternalServerError)
		return
	}
	var casResponse casResponse
	err = xml.Unmarshal(body, &casResponse)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Error while unmarshaling CAS response: %v", err), http.StatusInternalServerError)
		return
	}
	if casResponse.AuthenticationFailure != nil {
		RespondWithMessage(w, fmt.Sprintf("Authentification Failure: %s", casResponse.AuthenticationFailure.Message), http.StatusBadRequest)
		return
	}

	//Prepare transaction
	tx, err := cfg.DB.BeginTx(ctx, nil)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}
	qtx := cfg.DB.WithTx(tx)

	attributes := casResponse.AuthenticationSuccess.Attributes
	var businessCategory query.UsersBusinessCategory
	if attributes.BusinessCategory == "ELEVE" {
		businessCategory = query.UsersBusinessCategorySTUDENT
	} else {
		businessCategory = query.UsersBusinessCategoryTEACHER
	}
	err = qtx.AttemptCreatingUser(r.Context(), query.AttemptCreatingUserParams{
		Email:            attributes.Email,
		DepartmentNumber: attributes.DepartmentNumber,
		BusinessCategory: businessCategory,
		FullName:         attributes.CN,
	})
	if err != nil {
		_ = tx.Rollback()
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}
	userInfo, err := qtx.GetUserLastInsertID(ctx)
	if err != nil {
		_ = tx.Rollback()
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}

	//Create session for user
	sessionToken, err := generateSessionID(32)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Failed to to generate session token: %v", err), http.StatusInternalServerError)
		return
	}
	data := map[string]string{
		cfg.Security.Session.CookieName: sessionToken,
	}
	encoded, err := cfg.Security.Session.SecureCookie.Encode(cfg.Security.Session.CookieName, data)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Failed to set session: %v", err), http.StatusInternalServerError)
		return
	}
	cookie := &http.Cookie{
		Name:     cfg.Security.Session.CookieName,
		MaxAge:   int(cfg.Security.Session.CookieMaxAge.Seconds()),
		Secure:   cfg.Security.Session.CookieSecure,
		HttpOnly: cfg.Security.Session.CookieHTTPOnly,
		SameSite: cfg.Security.Session.CookieSameSite,
		Value:    encoded,
		Path:     "/",
	}
	err = cfg.DB.CreateSession(r.Context(), query.CreateSessionParams{UserID: userInfo.UserID, SessionToken: sessionToken})
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %v", err), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, cfg.Routes.Dashboard, http.StatusFound)
}

func generateSessionID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
