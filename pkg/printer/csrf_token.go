package printer

import (
	"errors"
	"regexp"
)

var errCSRFTokenNotFound = errors.New("printer: get: failed to find csrf token")

// parseBodyForCSRFToken returns the csrfToken contained in the html
// response input
func parseBodyForCSRFToken(bodyBytes []byte) (csrfToken string, err error) {
	// e.g. <id="CSRFToken1 name="CSRFToken" value="xyz"/>`
	regex := regexp.MustCompile(`id="CSRFToken[0-9]*"\s+name="CSRFToken"\s+value="([^"]*)"/>`)
	caps := regex.FindStringSubmatch(string(bodyBytes))

	// error if didn't find what was expected
	if len(caps) != 2 {
		return "", errCSRFTokenNotFound
	}

	return caps[1], nil
}
