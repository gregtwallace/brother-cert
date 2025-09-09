package printer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// SetActiveCert sets the printers active certificate the specified ID and
// then restarts the printer (to make the new cert active)
// Note: This function even works of the `id` is not in the dropdown box of the printer's
// cert picker (which happens when the cert does not have a Common Name)
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
