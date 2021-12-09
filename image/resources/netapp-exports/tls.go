package main

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var oidExtensionSubjectAltName = []int{2, 5, 29, 17}

func verifyWithCommonName(config *tls.Config) func(tls.ConnectionState) error {
	return func(c tls.ConnectionState) error {
		t := config.Time
		if t == nil {
			t = time.Now
		}

		certs := c.PeerCertificates
		intermediates := x509.NewCertPool()
		for _, cert := range certs[1:] {
			intermediates.AddCert(cert)
		}

		opts := x509.VerifyOptions{
			Roots:         config.RootCAs,
			CurrentTime:   t(),
			DNSName:       "",
			Intermediates: intermediates,
		}

		cert := certs[0]
		if _, err := cert.Verify(opts); err != nil {
			return err
		}

		host := c.ServerName
		if hasSANExtension(cert) {
			// If the certificate contains a SAN ignore the common name and
			// only validate based on the SAN.
			if err := cert.VerifyHostname(host); err != nil {
				return err
			}
		} else {
			// If the certificate does not contain a SAN, fallback to the older
			// method of assuming the common name is a DNS name
			if err := verifyCommonName(cert, host); err != nil {
				return err
			}
		}

		return nil
	}
}

func verifyCommonName(c *x509.Certificate, host string) error {
	cn := c.Subject.CommonName
	if matchCommonName(cn, host) {
		return nil
	}
	return errors.New("x509: certificate is valid for " + cn + ", not " + host)
}

func matchCommonName(cn, host string) bool {
	// Based on https://github.com/golang/go/blob/go1.17.2/src/crypto/x509/verify.go#L1025
	host = toLowerCaseASCII(host)
	if validHostnameInput(host) && validHostnamePattern(cn) {
		return matchHostnames(cn, host)
	} else {
		return matchExactly(cn, host)
	}
}

// From https://github.com/golang/go/blob/go1.17.2/src/crypto/x509/x509.go#L755
func hasSANExtension(c *x509.Certificate) bool {
	return oidInExtensions(oidExtensionSubjectAltName, c.Extensions)
}

// From https://github.com/golang/go/blob/go1.17.2/src/crypto/x509/x509.go#L963
// oidNotInExtensions reports whether an extension with the given oid exists in
// extensions.
func oidInExtensions(oid asn1.ObjectIdentifier, extensions []pkix.Extension) bool {
	for _, e := range extensions {
		if e.Id.Equal(oid) {
			return true
		}
	}
	return false
}

func validHostnamePattern(host string) bool { return validHostname(host, true) }
func validHostnameInput(host string) bool   { return validHostname(host, false) }

// From https://github.com/golang/go/blob/go1.17.2/src/crypto/x509/verify.go#L889
// validHostname reports whether host is a valid hostname that can be matched or
// matched against according to RFC 6125 2.2, with some leniency to accommodate
// legacy values.
func validHostname(host string, isPattern bool) bool {
	if !isPattern {
		host = strings.TrimSuffix(host, ".")
	}
	if len(host) == 0 {
		return false
	}

	for i, part := range strings.Split(host, ".") {
		if part == "" {
			// Empty label.
			return false
		}
		if isPattern && i == 0 && part == "*" {
			// Only allow full left-most wildcards, as those are the only ones
			// we match, and matching literal '*' characters is probably never
			// the expected behavior.
			continue
		}
		for j, c := range part {
			if 'a' <= c && c <= 'z' {
				continue
			}
			if '0' <= c && c <= '9' {
				continue
			}
			if 'A' <= c && c <= 'Z' {
				continue
			}
			if c == '-' && j != 0 {
				continue
			}
			if c == '_' {
				// Not a valid character in hostnames, but commonly
				// found in deployments outside the WebPKI.
				continue
			}
			return false
		}
	}

	return true
}

// From https://github.com/golang/go/blob/go1.17.2/src/crypto/x509/verify.go#L933
func matchExactly(hostA, hostB string) bool {
	if hostA == "" || hostA == "." || hostB == "" || hostB == "." {
		return false
	}
	return toLowerCaseASCII(hostA) == toLowerCaseASCII(hostB)
}

// From https://github.com/golang/go/blob/go1.17.2/src/crypto/x509/verify.go#L940
func matchHostnames(pattern, host string) bool {
	pattern = toLowerCaseASCII(pattern)
	host = toLowerCaseASCII(strings.TrimSuffix(host, "."))

	if len(pattern) == 0 || len(host) == 0 {
		return false
	}

	patternParts := strings.Split(pattern, ".")
	hostParts := strings.Split(host, ".")

	if len(patternParts) != len(hostParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if i == 0 && patternPart == "*" {
			continue
		}
		if patternPart != hostParts[i] {
			return false
		}
	}

	return true
}

// From https://github.com/golang/go/blob/go1.17.2/src/crypto/x509/verify.go#L970
// toLowerCaseASCII returns a lower-case version of in. See RFC 6125 6.4.1. We use
// an explicitly ASCII function to avoid any sharp corners resulting from
// performing Unicode operations on DNS labels.
func toLowerCaseASCII(in string) string {
	// If the string is already lower-case then there's nothing to do.
	isAlreadyLowerCase := true
	for _, c := range in {
		if c == utf8.RuneError {
			// If we get a UTF-8 error then there might be
			// upper-case ASCII bytes in the invalid sequence.
			isAlreadyLowerCase = false
			break
		}
		if 'A' <= c && c <= 'Z' {
			isAlreadyLowerCase = false
			break
		}
	}

	if isAlreadyLowerCase {
		return in
	}

	out := []byte(in)
	for i, c := range out {
		if 'A' <= c && c <= 'Z' {
			out[i] += 'a' - 'A'
		}
	}
	return string(out)
}
