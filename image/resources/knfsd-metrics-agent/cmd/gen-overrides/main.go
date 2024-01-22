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
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type metadata struct {
	Name    string            `yaml:"name"`
	Metrics map[string]metric `yaml:"metrics"`
}

type metric struct {
	Enabled bool `yaml:"enabled"`
}

func main() {
	log.SetFlags(0)

	files, err := findMetadata()
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		var meta metadata
		err = readYamlFile(f, &meta)
		if err != nil {
			log.Fatal(err)
		}
		if len(meta.Metrics) > 1 {
			formatOverrides(meta)
		}
	}
}

func findMetadata() ([]string, error) {
	var files []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if d.Name() == "metadata.yaml" && d.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func formatOverrides(meta metadata) {
	fmt.Println()
	fmt.Println(meta.Name)
	fmt.Printf("    # metrics:\n")
	for k, v := range meta.Metrics {
		fmt.Printf("    #   %s:\n", k)
		fmt.Printf("    #     enabled: %v\n", !v.Enabled)
	}
}

func readYamlFile(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	err = d.Decode(v)
	if err != nil {
		return err
	}

	return nil
}
