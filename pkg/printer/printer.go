package printer

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

// printer is a struct to interact with a remote Brother printer
type printer struct {
	httpClient *http.Client
	baseUrl    string
	userAgent  string
}

// PrinterConfig contains the information necessary to create a printer
// type which interfaces with a remote Brother printer
type Config struct {
	Hostname  string
	Password  string
	UserAgent string
	UseHttp   bool
}

// NewPrinter creates a new printer from a PrinterConfig
func NewPrinter(cfg Config) (*printer, error) {
	baseUrl := "https://" + cfg.Hostname
	// http instead?
	if cfg.UseHttp {
		baseUrl = "http://" + cfg.Hostname
	}

	// make cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	p := &printer{
		httpClient: &http.Client{
			// disable redirect (POSTs return 301 and if client follows it loses the post response)
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: jar,

			// set client timeout
			Timeout:   30 * time.Second,
			Transport: http.DefaultTransport,
		},
		baseUrl:   baseUrl,
		userAgent: cfg.UserAgent,
	}

	// login & get cookie
	err = p.login(cfg.Password)
	if err != nil {
		return nil, err
	}

	return p, nil
}
