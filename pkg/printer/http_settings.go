package printer

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const urlHttpCertServerSettings = "net/net/certificate/http.html"

var (
	errCurrentCertIdNotFound = errors.New("printer: get: failed to find current cert id")
)

// getHttpSettings fetches the HTTP Server Settings page
func (p *printer) getHttpSettings() ([]byte, error) {
	// get url & set path
	u, err := url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return nil, err
	}
	u.Path = urlHttpCertServerSettings

	// make and do request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read body of response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// OK status?
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("printer: get of http settings page failed (status code %d)", resp.StatusCode)
	}

	return bodyBytes, nil
}

// GetCurrentCertID returns the ID integer and name of the currently selected
// certificate
func (p *printer) GetCurrentCertID() (id string, name string, err error) {
	// GET http settings
	bodyBytes, err := p.getHttpSettings()
	if err != nil {
		return "", "", err
	}

	// find the selected cert in the returned html
	// e.g. `<option value="3" selected="selected">xxx</option>`
	regex := regexp.MustCompile(`<option[^>]+(?:value="([^"]+)"[^>]+selected="selected"[^>]*|selected="selected"[^>]+value="([^"]+)"[^>]*)>(\S*)<\/option>`)
	caps := regex.FindSubmatch(bodyBytes)

	// len must be 4 ([0] is the entire match)
	if len(caps) != 4 {
		return "", "", errCurrentCertIdNotFound
	}

	// first capture opportunity for id
	id = ""
	if len(caps[1]) != 0 {
		id = string(caps[1])
	}

	// second capture opportunity for id
	if id == "" && len(caps[2]) != 0 {
		id = string(caps[2])
	}

	// verify valid id obtained
	if id == "" {
		return "", "", errCurrentCertIdNotFound
	}

	// name will be in html char codes, so unescape it
	return id, html.UnescapeString(string(caps[3])), nil
}

// SetActiveCert sets the printers active certificate the specified ID and
// then restarts the printer (to make the new cert active)
func (p *printer) SetActiveCert(id string) error {
	// GET http settings
	bodyBytes, err := p.getHttpSettings()
	if err != nil {
		return err
	}

	// find CSRFToken
	csrfToken, err := parseBodyForCSRFToken(bodyBytes)
	if err != nil {
		return err
	}

	// submit initial form to change the cert
	data := url.Values{}
	data.Set("pageid", "326")
	data.Set("CSRFToken", csrfToken)
	data.Set("B903", id)
	// B91d always seems to be 1, but wasn't needed here
	// Enable HTTPS for WebUI and IPP
	data.Set("B86c", "1")
	data.Set("B87e", "1")
	// there are some other values here but don't set them (which should
	// leave them as-is in most cases)

	// get url & set path
	u, err := url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return err
	}
	u.Path = urlHttpCertServerSettings

	// make and do request
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read body of response
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// OK status?
	if resp.StatusCode != http.StatusOK {
		return errors.New("printer: failed to post set active cert form")
	}

	// find next CSRFToken
	csrfToken, err = parseBodyForCSRFToken(bodyBytes)
	if err != nil {
		return err
	}

	// submit confirmation (& reboot now)
	data = url.Values{}
	data.Set("pageid", "326")
	data.Set("CSRFToken", csrfToken)
	// 4 == do NOT activate other secure protos
	// 5 == DO activate other secure protos
	data.Set("http_page_mode", "5")

	// get url & set path
	u, err = url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return err
	}
	u.Path = urlHttpCertServerSettings

	// make and do request
	req, err = http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read body of response
	_, _ = io.Copy(io.Discard, resp.Body)

	// OK status?
	if resp.StatusCode != http.StatusOK {
		return errors.New("printer: failed to post set active cert form")
	}

	return nil
}
