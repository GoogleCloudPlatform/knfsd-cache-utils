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
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-fsidd/log"

	"github.com/go-ini/ini"
	"go.uber.org/multierr"
)

const (
	defaultConfigFile = "/etc/knfsd-fsidd.conf"
	defaultSocketPath = "/run/fsidd.sock"
)

type Config struct {
	SocketPath string         `ini:"socket"`
	Database   DatabaseConfig `ini:"database"`
	Debug      bool           `ini:"debug"`
	Cache      bool           `ini:"cache"`
}

type DatabaseConfig struct {
	URL       string `ini:"url"`
	Instance  string `ini:"instance"`
	IAMAuth   bool   `ini:"iam-auth"`
	PrivateIP bool   `ini:"private-ip"`

	TableName   string `ini:"table-name"`
	CreateTable bool   `ini:"create-table"`
}

func (cfg *Config) Validate() error {
	var err error
	err = multierr.Append(err, required("socket-path", cfg.SocketPath))
	err = multierr.Append(err, cfg.Database.Validate())
	return err
}

func (cfg *DatabaseConfig) Validate() error {
	var err error
	err = multierr.Append(err, required("database-url", cfg.URL))
	err = multierr.Append(err, required("database-instance", cfg.Instance))
	err = multierr.Append(err, required("table-name", cfg.TableName))
	return err
}

func readDefaultConfig(cfg *Config) error {
	err := readConfig(cfg, defaultConfigFile)
	if errors.Is(err, os.ErrNotExist) {
		// if config file does not exist, use default values
		err = nil
	}
	return err
}

func readConfig(cfg *Config, name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return parseConfig(cfg, f)
}

func parseConfig(cfg *Config, r io.Reader) error {
	i, err := ini.Load(r)
	if err != nil {
		return err
	}
	return i.StrictMapTo(cfg)
}

func readEnv(cfg *Config) error {
	var err error
	envString(&cfg.SocketPath, "FSID_SOCKET")
	envString(&cfg.Database.URL, "FSID_DATABASE_URL")
	envString(&cfg.Database.Instance, "FSID_DATABASE_INSTANCE")
	envString(&cfg.Database.TableName, "FSID_TABLE_NAME")
	err = multierr.Append(err, envBool(&cfg.Database.IAMAuth, "FSID_IAM_AUTH"))
	err = multierr.Append(err, envBool(&cfg.Database.PrivateIP, "FSID_PRIVATE_IP"))
	err = multierr.Append(err, envBool(&cfg.Debug, "FSID_DEBUG"))
	err = multierr.Append(err, envBool(&cfg.Debug, "FSID_CACHE"))
	return err
}

func envString(value *string, key string) {
	if s, _ := os.LookupEnv(key); s != "" {
		*value = s
	}
}

func envBool(value *bool, key string) error {
	if s, _ := os.LookupEnv(key); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("invalid argument %q for %q: %w", s, key, err)
		}
		*value = b
	}
	return nil
}

func required(name, value string) error {
	if value == "" {
		return fmt.Errorf("required: %q", name)
	} else {
		return nil
	}
}

func printConfigError(err error) {
	msg := &strings.Builder{}
	fmt.Fprintln(msg, "invalid configuration:")
	for _, e := range multierr.Errors(err) {
		fmt.Fprintf(msg, "  - %v\n", e)
	}
	log.Error.Print(msg.String())
}
