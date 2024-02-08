package printer

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const urlCertDelete = "/net/security/certificate/delete.html"

var errCertDeleteInvalidID = errors.New("printer: cant delete cert (invalid id)")

// DeleteCert deletes the certificate with the specified ID from the
// printer
func (p *printer) DeleteCert(id int) error {
	// verify ID actually exists and isn't 0 ('Preset') which isn't valid
	if id <= 0 {
		return errCertDeleteInvalidID
	}
	existingIDs, err := p.getCertIDs()
	if err != nil {
		return err
	}

	validID := false
	for _, existingID := range existingIDs {
		if existingID == id {
			validID = true
			break
		}
	}
	if !validID {
		return errCertDeleteInvalidID
	}

	// first get the delete page to get CSRFToken
	// get url & set path
	u, err := url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return err
	}
	u.Path = urlCertDelete

	// make and do request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	// req query
	query := req.URL.Query()
	query.Set("idx", strconv.Itoa(id))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read body of response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// OK status?
	if resp.StatusCode != http.StatusOK {
		return errGetFailed
	}

	// find CSRFToken
	csrfToken, err := parseBodyForCSRFToken(bodyBytes)
	if err != nil {
		return err
	}

	// first delete form
	// form values
	data := url.Values{}
	data.Set("pageid", "383")
	data.Set("CSRFToken", csrfToken)
	data.Set("B8ea", "")
	data.Set("B8fc", "")
	data.Set("hidden_certificate_process_control", "1")
	data.Set("hidden_certificate_idx", strconv.Itoa(id))

	// get url & set path
	u, err = url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return err
	}
	u.Path = urlCertDelete

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
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// OK status?
	if resp.StatusCode != http.StatusOK {
		return errGetFailed
	}

	// find CSRFToken
	csrfToken, err = parseBodyForCSRFToken(bodyBytes)
	if err != nil {
		return err
	}

	// second delete (confirmation) form
	// form values
	data = url.Values{}
	data.Set("pageid", "383")
	data.Set("CSRFToken", csrfToken)
	data.Set("B8ea", "")
	data.Set("B8eb", "")
	data.Set("hidden_certificate_process_control", "2")
	data.Set("hidden_certificate_idx", strconv.Itoa(id))

	// get url & set path
	u, err = url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return err
	}
	u.Path = urlCertDelete

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

	// read and discard entire body
	_, _ = io.Copy(io.Discard, resp.Body)

	// normally the webUI would show a waiting screen for ~7 seconds. insert
	// a delay here to account for any processing the device might do
	// before next steps
	time.Sleep(10 * time.Second)

	// check id list and ensure its gone
	existingIDs, err = p.getCertIDs()
	if err != nil {
		return err
	}

	idFound := false
	for _, existingID := range existingIDs {
		if existingID == id {
			idFound = true
			break
		}
	}
	if idFound {
		return errors.New("printer: failed to delete cert (still exists)")
	}

	return nil
}
