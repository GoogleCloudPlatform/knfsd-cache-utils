package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type Config struct {
	Servers []*NetAppServer `hcl:"server,block"`
}

// Password allows passing the password as a simple string from the command
// line for testing locally. In production password block should be used
// instead.
type NetAppServer struct {
	Host           string          `hcl:"host,label"`
	URL            string          `hcl:"url"`
	User           string          `hcl:"user"`
	Password       string          `hcl:"password,optional"`
	SecurePassword *NetAppPassword `hcl:"password,block"`
	TLS            *TLSConfig      `hcl:"tls,block"`
}

type NetAppPassword struct {
	GCPSecret *GCPSecret `hcl:"google_cloud_secret,block"`
}

type validatable interface {
	validate() error
}

func parseConfigFile(file string) (*Config, error) {
	src, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(file)
	return parseConfig(baseDir, file, src)
}

func parseConfig(baseDir, file string, src []byte) (*Config, error) {
	config := &Config{}

	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"file": makeFileFunc(baseDir),
		},
	}

	err := hclsimple.Decode(file, src, ctx, config)
	if err != nil {
		return nil, err
	}

	err = config.validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	if len(c.Servers) == 0 {
		return errors.New("no servers defined")
	}

	for _, s := range c.Servers {
		err := s.validate()
		if err != nil {
			return fmt.Errorf("error parsing server '%s': %w", s.Host, err)
		}
	}

	return nil
}

func (s *NetAppServer) validate() error {
	if s.Host == "" {
		return errors.New("host not set")
	}

	if s.URL == "" {
		return errors.New("API URL not set")
	}

	if s.User == "" {
		return errors.New("username not set")
	}

	if s.Password == "" && s.SecurePassword == nil {
		return errors.New("no password provided")
	}

	if s.Password != "" && s.SecurePassword != nil {
		return errors.New("only one password source permitted")
	}

	if s.SecurePassword != nil {
		s.SecurePassword.validate()
	}

	// Ensure TLS always has a value, this simplifies the code later by
	// removing repeated nil checks.
	if s.TLS == nil {
		s.TLS = &TLSConfig{}
	}

	return nil
}

func (p *NetAppPassword) validate() error {
	sources := []validatable{p.GCPSecret}

	count := 0
	for _, s := range sources {
		if s != nil {
			count++
		}
	}

	if count == 0 {
		return errors.New("no password provided")
	}

	if count > 1 {
		return errors.New("only one password source permitted")
	}

	for _, s := range sources {
		if s != nil {
			return s.validate()
		}
	}

	return nil
}

func filePath(baseDir, path string) (string, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}

	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	} else {
		return filepath.Join(baseDir, path), nil
	}
}

func readFileBytes(baseDir, path string) ([]byte, error) {
	path, err := filePath(baseDir, path)
	if err != nil {
		return nil, err
	}

	src, err := os.ReadFile(path)
	return src, err
}

func makeFileFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			src, err := readFileBytes(baseDir, path)
			if err != nil {
				err = function.NewArgError(0, err)
				return cty.UnknownVal(cty.String), err
			}

			if !utf8.Valid(src) {
				return cty.UnknownVal(cty.String), fmt.Errorf("contents of %s is not valid UTF-8", path)
			}

			return cty.StringVal(string(src)), nil
		},
	})
}
