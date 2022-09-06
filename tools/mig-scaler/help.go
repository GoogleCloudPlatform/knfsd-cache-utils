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
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed help
var helpFiles embed.FS

type Help struct {
	Topic string
}

func NewHelp(cfg *Config) Help {
	return Help{
		Topic: getHelpTopic(cfg),
	}
}

func (h Help) Execute(context.Context) error {
	f, err := helpFiles.Open(helpPath(h.Topic))
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unknown topic \"%s\"", h.Topic)
	} else if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(os.Stdout, f)
	if err != nil {
		return err
	}

	return nil
}

func showUsage(cfg *Config) {
	// Standard pflags.PrintDefaults isn't very helpful as it doesn't
	// understand commands. This will be called on an error, such as an
	// unknown flag, so keep the instructions short. Direct the user to the
	// help pages, try and include the topic if possible.
	topic := getHelpTopic(cfg)
	if topic == "main" || topic == "help" {
		topic = ""
	}
	if topic != "" && !topicExists(topic) {
		topic = ""
	}

	if topic == "" {
		fmt.Fprint(os.Stderr,
			"For a list of supported commands run:\n"+
				"    mig-scaler help\n",
		)
	} else {
		fmt.Fprintf(os.Stderr,
			"For a list of supported commands run:\n"+
				"    mig-scaler help %s\n",
			topic,
		)
	}
}

func getHelpTopic(cfg *Config) string {
	// check command first to support "mig-scaler list --help"
	topic := cfg.Command

	if topic == "help" {
		// support "mig-scaler help list"
		topic = cfg.Arg(0)
	}

	if topic == "" {
		topic = "main"
	}

	return topic
}

func topicExists(topic string) bool {
	_, err := fs.Stat(helpFiles, helpPath(topic))
	return err == nil
}

func helpPath(topic string) string {
	return filepath.Join("help", topic+".txt")
}
