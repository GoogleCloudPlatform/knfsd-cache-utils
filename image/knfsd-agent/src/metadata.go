package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	metadataServerURL = "http://metadata.google.internal"
)

// getMetadataValue fetches a metadata path from the GCE Metadata Server returning the output as a string
// if delimit is set to true then it will return the string after the last occurrence of /
func getMetadataValue(URI string, delimit bool) (string, error) {

	// Make a HTTP Client
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Build Request Object
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?alt=text", metadataServerURL, URI), nil)
	if err != nil {
		return "", err
	}

	// Set Headers
	req.Header.Set("Metadata-Flavor", "Google")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Validate Response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received invalid HTTP response code, got %d, wanted 200", resp.StatusCode)
	}

	// Return value
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Delimit if requested
	if delimit {
		return lastAfterDelimiter(string(body), "/")
	}

	// Else return full string
	return string(body), nil

}
