package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"unicode/utf8"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"google.golang.org/api/transport"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type GCPSecret struct {
	ServiceAccountKey string `hcl:"service_account_key,optional"`
	Project           string `hcl:"project,optional"`
	Name              string `hcl:"name"`
	Version           string `hcl:"version,optional"`
}

var cachedProjectID = ""

func (s *GCPSecret) validate() error {
	if s.Name == "" {
		return errors.New("name required for Google Cloud Secret")
	}
	return nil
}

func (s *GCPSecret) fullName(ctx context.Context) (string, error) {
	name := s.Name
	if name == "" {
		return "", errors.New("secret name not set")
	}

	project := s.Project
	if project == "" {
		var err error
		project, err = projectIDFromEnvironment(ctx)
		if err != nil {
			return "", err
		}
		log.Printf("Resolved project ID as %s\n", project)
	}

	version := s.Version
	if version == "" {
		version = "latest"
	}

	fullName := fmt.Sprintf(
		"projects/%s/secrets/%s/versions/%s",
		project, name, version)

	return fullName, nil
}

func (s *GCPSecret) get(ctx context.Context) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	name, err := s.fullName(ctx)
	if err != nil {
		return "", err
	}

	request := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	response, err := client.AccessSecretVersion(ctx, request)
	if err != nil {
		return "", err
	}

	data := response.GetPayload().GetData()
	if data == nil {
		return "", errors.New("secret did not contain any data")
	}

	if !utf8.Valid(data) {
		return "", errors.New("secret is not valid UTF8")
	}

	return string(data), nil
}

func projectIDFromEnvironment(ctx context.Context) (string, error) {
	if cachedProjectID != "" {
		return cachedProjectID, nil
	}

	creds, err := transport.Creds(ctx)
	if err != nil {
		return "", fmt.Errorf("could not get project ID from environment: %w", err)
	}

	project := creds.ProjectID
	if project == "" {
		return "", fmt.Errorf("could not get project ID from environment")
	}

	cachedProjectID = project
	return project, nil
}
