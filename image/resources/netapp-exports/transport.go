/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"crypto/x509"
	"errors"
	"net/http"
)

type TLSConfig struct {
	CACertificate   string `hcl:"ca_certificate,optional"`
	AllowCommonName bool   `hcl:"allow_common_name,optional"`

	// This attribute comes from a command line flag instead of HCL.
	// When true this allows unencrypted HTTP connections and
	// does not verify the server certificate for HTTPS connections.
	insecure bool
}

func (config *TLSConfig) transport() (http.RoundTripper, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	tls := transport.TLSClientConfig

	if config.CACertificate != "" {
		ca := x509.NewCertPool()
		ok := ca.AppendCertsFromPEM([]byte(config.CACertificate))
		if !ok {
			return nil, errors.New("ca_certificate did not contain any PEM encoded certificates")
		}
		tls.RootCAs = ca
	}

	if config.insecure {
		tls.InsecureSkipVerify = true
	} else {
		transport.RegisterProtocol("http", denyHTTPTransport{})

		if config.AllowCommonName {
			// Replace the standard validation with our custom validation
			tls.InsecureSkipVerify = true
			tls.VerifyConnection = verifyWithCommonName(tls)
		}
	}

	return transport, nil
}

type denyHTTPTransport struct{}

func (rt denyHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("HTTP not permitted unless insecure specified")
}
