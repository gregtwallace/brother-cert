package printer

import (
	"errors"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const urlHttpCertServerSettings = "net/net/certificate/http.html"

var (
	errGetFailed             = errors.New("printer: get: failed")
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

	req.Header.Set("User-Agent", p.userAgent)

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
		return nil, errGetFailed
	}

	return bodyBytes, nil
}

// GetCurrentCertID returns the ID integer and name of the currently selected
// certificate
func (p *printer) GetCurrentCertID() (id int, name string, err error) {
	// GET http settings
	bodyBytes, err := p.getHttpSettings()
	if err != nil {
		return -1, "", err
	}

	// find the selected cert in the returned html
	// e.g. `<option value="3" selected="selected">xxx</option>`
	regex := regexp.MustCompile(`<option\s+?value="(\S*)"\s+?selected="selected">(\S*)<\/option>`)
	caps := regex.FindStringSubmatch(string(bodyBytes))

	// error if didn't find what was expected
	if len(caps) != 3 {
		return -1, "", errCurrentCertIdNotFound
	}

	// id must be a number
	id, err = strconv.Atoi(caps[1])
	if err != nil {
		return -1, "", errCurrentCertIdNotFound
	}

	// name will be in html char codes, so unescape it
	return id, html.UnescapeString(caps[2]), nil
}

// SetActiveCert sets the printers active certificate the specified ID and
// then restarts the printer (to make the new cert active)
func (p *printer) SetActiveCert(id int) error {
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
	data.Set("B903", strconv.Itoa(id))
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
	req.Header.Set("User-Agent", p.userAgent)

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
	req.Header.Set("User-Agent", p.userAgent)

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
