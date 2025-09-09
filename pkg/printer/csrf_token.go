package printer

import (
	"errors"
	"regexp"
)

var errCSRFTokenNotFound = errors.New("printer: get: failed to find csrf token")

// parseBodyForCSRFToken returns the csrfToken contained in the html
// response input
func parseBodyForCSRFToken(bodyBytes []byte) (csrfToken string, err error) {
	// e.g. `<input type="hidden" id="CSRFToken" name="CSRFToken" value="JRL[...snip...]bQ=="/>`
	regex := regexp.MustCompile(`<input[^>]+(?:id="CSRFToken"[^>]+value="([^"]+)"[^>]*|value="([^"]+)"[^>]+id="CSRFToken"[^>]*)>`)
	caps := regex.FindSubmatch(bodyBytes)

	// error if wrong length
	if len(caps) != 3 {
		return "", errCSRFTokenNotFound
	}

	// return the non-empty capture group (either caps[1] or caps[2])
	if len(caps[1]) > 0 {
		return string(caps[1]), nil
	}
	return string(caps[2]), nil
}
