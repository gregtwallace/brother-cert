package printer

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	urlCertList = "/net/security/certificate/certificate.html"
	urlCertView = "/net/security/certificate/view.html"
)

// getCertIDs loads the certificate page and parses it to obtain the
// IDs of the existing certificates
func (p *printer) getCertIDs() ([]string, error) {
	// get url & set path
	u, err := url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return nil, err
	}
	u.Path = urlCertList

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
		return nil, fmt.Errorf("printer: get of certificate list page failed (status code %d)", resp.StatusCode)
	}

	// parse IDs
	// e.g. `<td><a href="view.html?idx=58">View</a></td>`
	regex := regexp.MustCompile(`<a[^>]+href="view\.html\?idx=([^"]+)"[^>]*>`)
	caps := regex.FindAllSubmatch(bodyBytes, -1)

	// range through matches and get capture group (the actual ID)
	ids := []string{}
	for i := range caps {
		// if match is somehow the wrong length, skip it
		if len(caps[i]) != 2 {
			continue
		}

		ids = append(ids, string(caps[i][1]))
	}

	return ids, nil
}

// getCertgetCertIDSerialIDs loads the certificate view page and parses the
// cert's serial number hex string into hex data
func (p *printer) getCertIDSerial(id string) ([]byte, error) {
	// get url & set path
	u, err := url.ParseRequestURI(p.baseUrl)
	if err != nil {
		return nil, err
	}
	u.Path = urlCertView

	// make request
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// set cert id
	q := req.URL.Query()
	q.Add("idx", id)
	req.URL.RawQuery = q.Encode()

	// do request
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
		return nil, fmt.Errorf("printer: get certificate view page failed (status code %d)", resp.StatusCode)
	}

	// parse Serial Number string
	// e.g. `<dt>Serial&#32;Number</dt><dd>06:22:61:1a:32:3a:f8:ea:5b:be:3f:6c:53:a2:1e:d2:a4:c4</dd><dt>Issuer</dt>`
	regex := regexp.MustCompile(`<dt>Serial(?:\s|&#32;)Number</dt><dd>([A-Za-z0-9:]+)</dd>`)
	caps := regex.FindSubmatch(bodyBytes)

	if len(caps) < 2 {
		return nil, fmt.Errorf("printer: get cert serial for id '%s' from view page failed (unable to parse serial)", id)
	}

	// range over hex string and convert each value into a byte
	byteChars := ""
	serial := []byte{}

	for i := range len(caps[1]) {
		// ensure each byte is exactly 2 characters
		if caps[1][i] == '\x3A' {
			// allow flexibility for invalid `:` at start and end of string
			if (i != 0 && i != len(caps[1])-1) && len(byteChars) != 2 {
				return nil, fmt.Errorf("printer: get cert serial for id '%s' from view page failed (serial format incorrect '%s')", id, string(caps[1]))
			}

			// reset for next byte
			byteChars = ""
			continue
		}

		// append char to byteChars
		byteChars += string(caps[1][i])

		// not yet both chars
		if len(byteChars) == 1 {
			continue
		}

		// too many chars
		if len(byteChars) > 2 {
			return nil, fmt.Errorf("printer: get cert serial for id '%s' from view page failed (serial format incorrect '%s')", id, string(caps[1]))
		}

		// convert the letter/number values into a byte
		oneByte, err := hex.DecodeString(byteChars)
		if err != nil {
			return nil, fmt.Errorf("printer: get cert serial for id '%s' from view page failed (serial format incorrect '%s') (%s)", id, string(caps[1]), err)
		}

		serial = append(serial, oneByte...)
	}

	return serial, nil
}

// getCurrentCertIDFromHttpSettings is the preferred way to get the currently active HTTPS
// certificate ID as it definitively only requires one page load; however, this may not always
// work as at least some printers do not list certificates without a Common Name, even if said
// certificate is currently active
func (p *printer) getCurrentCertIDFromHttpSettings() (id string, name string, err error) {
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

// GetCurrentLeafCert() returns the current Certificate that is being used by the
// printer for SSL connections. This is achieved by performing a TLS handshake
// with the printer
func (p *printer) GetCurrentLeafCert() (*x509.Certificate, error) {
	// use tls handshake to get the serial of the active certificate
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", strings.TrimPrefix(p.baseUrl, "https://")+":443", conf)
	if err != nil {
		return nil, fmt.Errorf("printer: failed to perform tls handshake with printer (dial failed: %s)", err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) <= 0 {
		return nil, errors.New("printer: failed to get ssl cert from printer")
	}

	return certs[0], nil
}

// getCurrentCertIDFromCertList performs a tls handshake with the printer to retrieve the
// current SSL cert. Then it compares the cert used in the handshake against the cert list
// of the printer in order to determine which is active.
// NOTE: If there is more than one copy of the active cert on the printer (which is possible
// if you upload the same cert twice), it is not possible to distinguish which is which and
// only one will be deleted.
func (p *printer) getCurrentCertIDFromCertList() (id string, err error) {
	// get currently in use cert
	leafCert, err := p.GetCurrentLeafCert()
	if err != nil {
		return "", err
	}

	// get the list of all certs on the printer
	printerCertIDs, err := p.getCertIDs()
	if err != nil {
		return "", fmt.Errorf("printer: failed to get ssl cert list from printer (%s)", err)
	}

	// for each printer cert id, fetch its view page, parse the serial, and compare it against
	// the serial acquired during the tls handshake
	for _, certID := range printerCertIDs {
		certSerial, err := p.getCertIDSerial(certID)
		if err != nil {
			// failed? keep trying other options
			continue
		}

		// if serials match, return the id
		if bytes.EqualFold(certSerial, leafCert.SerialNumber.Bytes()) {
			return certID, nil
		}
	}

	return "", fmt.Errorf("printer: get current id from cert list failed (no serial match)")
}

// GetCurrentCertID returns the ID integer and name of the currently selected
// certificate
func (p *printer) GetCurrentCertID() (id string, name string, err error) {
	// try the "easy" method first
	id, name, err = p.getCurrentCertIDFromHttpSettings()
	// NOTE: Inverted error check!
	if err == nil {
		return id, name, nil
	}

	// easy way didn't work, try the longer way
	if !strings.HasPrefix(strings.ToLower(p.baseUrl), "https://") {
		return "", "", errors.New("printer: get current cert id failed (not in http settings list and https isn't available)")
	}

	id, err = p.getCurrentCertIDFromCertList()
	if err != nil {
		return "", "", err
	}

	return id, "[no name]", err
}
