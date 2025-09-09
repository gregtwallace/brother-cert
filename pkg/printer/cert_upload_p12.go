package printer

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"software.sslmate.com/src/go-pkcs12"
)

// helper funcs to create p12 from pem

var errUnsupportedKey = errors.New("printer: error: only rsa keys are supported")

// keyPemToKey returns the private key from pemBytes
func keyPemToKey(keyPem []byte) (key *rsa.PrivateKey, err error) {
	// decode private key
	keyPemBlock, _ := pem.Decode(keyPem)
	if keyPemBlock == nil {
		return nil, errors.New("printer: key pem block did not decode")
	}

	// parsing depends on block type
	switch keyPemBlock.Type {
	case "RSA PRIVATE KEY": // PKCS1
		var rsaKey *rsa.PrivateKey
		rsaKey, err = x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			return nil, err
		}

		// basic sanity check
		err = rsaKey.Validate()
		if err != nil {
			return nil, err
		}

		return rsaKey, nil

	case "PRIVATE KEY": // PKCS8
		pkcs8K, err := x509.ParsePKCS8PrivateKey(keyPemBlock.Bytes)
		if err != nil {
			return nil, err
		}

		switch pkcs8Key := pkcs8K.(type) {
		case *rsa.PrivateKey:
			// basic sanity check
			err = pkcs8Key.Validate()
			if err != nil {
				return nil, err
			}

			return pkcs8Key, nil

		default:
			// fallthrough
		}

	default:
		// fallthrough
	}

	return nil, errUnsupportedKey
}

// certPemToCerts returns the certificate from cert pem bytes. if the pem
// bytes contain more than one certificate, the first is returned as the
// certificate and the 2nd is returned as the only member of an array. The
// rest of the chain is discarded as more than 2 certs are too big to fit
// on the printer
func certPemToCerts(certPem []byte) (cert *x509.Certificate, certChain []*x509.Certificate, err error) {
	// decode 1st cert
	certPemBlock, rest := pem.Decode(certPem)
	if certPemBlock == nil {
		return nil, nil, errors.New("printer: cert leaf pem block did not decode")
	}

	// parse 1st cert
	cert, err = x509.ParseCertificate(certPemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// decode 2nd cert
	cert2PemBlock, _ := pem.Decode(rest)
	if cert2PemBlock == nil {
		// return early: no chain cert
		return cert, nil, nil
	}

	// parse 2nd cert
	cert2, err := x509.ParseCertificate(cert2PemBlock.Bytes)
	if err != nil {
		// there was a chain cert, but something is wrong with it
		return nil, nil, errors.New("printer: cert chain pem block did not decode")
	}

	return cert, []*x509.Certificate{cert2}, nil
}

// makeModernPfx returns the pkcs12 pfx data for the given key and cert pem
func makeModernPfx(keyPem, certPem []byte, password string) (pfxData []byte, err error) {
	// get private key
	key, err := keyPemToKey(keyPem)
	if err != nil {
		return nil, err
	}

	// get cert and chain (if there is a chain)
	cert, certChain, err := certPemToCerts(certPem)
	if err != nil {
		return nil, err
	}

	// encode using modern pkcs12 standard
	pfxData, err = pkcs12.Modern.Encode(key, cert, certChain, password)
	if err != nil {
		return nil, err
	}

	return pfxData, nil
}
