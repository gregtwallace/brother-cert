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
}

// PrinterConfig contains the information necessary to create a printer
// type which interfaces with a remote Brother printer
type Config struct {
	Hostname  string
	Password  string
	UserAgent string
	UseHttp   bool
}

// custom transport to add User-Agent
type printerTransport struct {
	userAgent string
}

func (trans *printerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// always set user-agent
	req.Header.Set("User-Agent", trans.userAgent)

	return http.DefaultTransport.RoundTrip(req)
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
			Timeout: 30 * time.Second,
			Transport: &printerTransport{
				userAgent: cfg.UserAgent,
			},
		},
		baseUrl: baseUrl,
	}

	// login & get cookie
	err = p.login(cfg.Password)
	if err != nil {
		return nil, err
	}

	return p, nil
}
