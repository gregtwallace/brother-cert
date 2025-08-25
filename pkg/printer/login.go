package printer

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const urlLogin = "/general/status.html"

var (
	errLoginNoAuth           = errors.New("printer: login: no auth cookie received (wrong password?)")
	errPasswordFieldNotFound = errors.New("printer: login: password field not found in login form")
)

// parsePasswordFieldName returns the name attribute of the password input field
// from the HTML login form
func parsePasswordFieldName(bodyBytes []byte) (fieldName string, err error) {
	// Look for input elements with type="password"
	// This regex handles both orders: type first or name first
	// e.g. <input type="password" name="Baf9" ... /> or <input name="Baf9" type="password" ... />
	regex := regexp.MustCompile(`<input[^>]*(?:type="password"[^>]*name="([^"]*)"[^>]*|name="([^"]*)"[^>]*type="password"[^>]*)>`)
	caps := regex.FindStringSubmatch(string(bodyBytes))

	// error if didn't find what was expected
	if len(caps) < 2 {
		return "", errPasswordFieldNotFound
	}

	// return the non-empty capture group (either caps[1] or caps[2])
	if caps[1] != "" {
		return caps[1], nil
	}
	return caps[2], nil
}

// login performs the login command against the remote printer. it is
// used internally as part of the printer creation process to ensure
// credentials are valid
func (p *printer) login(password string) error {
	// get url & set path
	u, err := url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return err
	}
	u.Path = urlLogin

	// first, fetch the login page to discover the password field name
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read the login page HTML
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// parse the password field name from the HTML
	passwordFieldName, err := parsePasswordFieldName(bodyBytes)
	if err != nil {
		return err
	}

	// login form values using the discovered field name
	data := url.Values{}
	data.Set(passwordFieldName, password)
	data.Set("loginurl", urlLogin)

	// make and do login request
	req, err = http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", p.userAgent)

	resp, err = p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read and discard entire body (should be empty)
	_, _ = io.Copy(io.Discard, resp.Body)

	// confirm got cookie
	foundAuthCookie := false
	for _, c := range resp.Cookies() {
		if c.Name == "AuthCookie" {
			foundAuthCookie = true
			break
		}
	}
	if !foundAuthCookie {
		return errLoginNoAuth
	}

	// set cookies in jar
	p.httpClient.Jar.SetCookies(u, resp.Cookies())

	return nil
}
