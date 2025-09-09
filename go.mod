module github.com/gregtwallace/brother-cert

go 1.22.1

require (
	github.com/peterbourgon/ff/v4 v4.0.0-alpha.4
	software.sslmate.com/src/go-pkcs12 v0.6.0
)

require golang.org/x/crypto v0.11.0 // indirect

replace github.com/gregtwallace/brother-cert/cmd/brother-cert => /pkg/cmd/brother-cert

replace github.com/gregtwallace/brother-cert/pkg/app => /pkg/app

replace github.com/gregtwallace/brother-cert/pkg/printer => /pkg/printer
