package printer

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const urlLogin = "/general/status.html"

var errLoginNoAuth = errors.New("printer: login: no auth cookie received (wrong password?)")

// login performs the login command against the remote printer. it is
// used internally as part of the printer creation process to ensure
// credentials are valid
func (p *printer) login(password string) error {
	// login form values
	data := url.Values{}
	data.Set("B8d9", password)
	data.Set("loginurl", urlLogin)

	// get url & set path
	u, err := url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return err
	}
	u.Path = urlLogin

	// make and do request
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.httpClient.Do(req)
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
