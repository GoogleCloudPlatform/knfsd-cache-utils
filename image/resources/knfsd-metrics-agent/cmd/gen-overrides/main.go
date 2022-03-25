package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
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
