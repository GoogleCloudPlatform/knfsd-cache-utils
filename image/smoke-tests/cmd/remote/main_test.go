/*
 Copyright 2023 Google LLC

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
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
)

var (
	sourceHostFlag string
	proxyHostFlag  string

	sourceHost string
	proxyHost  string
	proxy      *client.KnfsdAgentClient
)

func getHost(attributeName, flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	host, err := QueryAttribute(attributeName)
	if err != nil {
		err = fmt.Errorf("metadata attribute %s: %w", attributeName, err)
	} else if host == "" {
		err = fmt.Errorf("metadata attribute %s not set", attributeName)
	}

	return host, err
}

func buildProxyURL(host string) (url.URL, error) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/",
	}
	return u, nil
}

func TestMain(m *testing.M) {
	var err error

	flag.StringVar(&sourceHostFlag, "source", "", "IP or DNS name of source NFS server")
	flag.StringVar(&proxyHostFlag, "proxy", "", "IP or DNS name of proxy NFS server")
	flag.Parse()

	sourceHost, err = getHost("source_host", sourceHostFlag)
	fatal(err)

	proxyHost, err = getHost("proxy_host", proxyHostFlag)
	fatal(err)

	proxyURL, err := buildProxyURL(proxyHost)
	fatal(err)

	proxy = client.NewKnfsdAgentClient(http.DefaultClient, proxyURL.String())

	code := m.Run()
	os.Exit(code)
}

func fatal(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "error starting tests: ")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
