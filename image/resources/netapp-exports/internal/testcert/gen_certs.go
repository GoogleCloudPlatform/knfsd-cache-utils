//go:build ignore

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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

var out = os.Stdout

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Fatalf("could not generate private key: %v", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(1000000 * time.Hour)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},

		NotBefore: notBefore,
		NotAfter:  notAfter,

		IsCA:     true,
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,

		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	cert, err := encodeCert(&template, key)
	if err != nil {
		log.Fatal(err)
	}

	template.DNSNames = nil
	template.IPAddresses = nil
	template.Subject.CommonName = "localhost"
	commonNameCert, err := encodeCert(&template, key)
	if err != nil {
		log.Fatal(err)
	}

	keyOut, err := encodeKey(key)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(out, "package testcert")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "// generated")
	fmt.Fprintln(out, "// go run gen_certs.go > testcert.go")
	fmt.Fprintln(out)
	write("PEM encoded TLS certificate with SAN for localhost and 127.0.0.1", "LocalhostCert", cert)
	fmt.Fprintln(out)
	write("PEM encoded TLS certificate with CN for localhost (no SAN)", "CommonNameCert", commonNameCert)
	fmt.Fprintln(out)
	write("PEM encoded private key for above TLS certificates", "PrivateKey", keyOut)
}

func write(comment, name, block string) {
	fmt.Fprintf(out, "// %s\nvar %s = []byte(`%s`)\n", comment, name, block)
}

func encodeCert(template *x509.Certificate, key *rsa.PrivateKey) (string, error) {
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return "", err
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	return strings.TrimSpace(string(pemBytes)), nil
}

func encodeKey(key *rsa.PrivateKey) (string, error) {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	return strings.TrimSpace(string(pemBytes)), nil
}
