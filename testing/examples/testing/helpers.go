/*
 Copyright 2024 Google LLC

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

package testing

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/logger"
	scope "github.com/gruntwork-io/terratest/modules/test-structure"
)

func GetTestID(t TestingT, testFolder string) string {
	path := scope.FormatTestDataPath(testFolder, "test-id")

	if bytes, err := os.ReadFile(path); err == nil {
		id := strings.TrimSpace(string(bytes))
		if id != "" {
			logger.Default.Logf(t, "Using existing TestID \"%s\"", id)
			return id
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("Failed to read %s: %v", path, err)
	}

	id := gcp.RandomValidGcpName()
	bytes := []byte(id)
	logger.Default.Logf(t, "Created new TestID \"%s\"", id)

	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		t.Fatalf("Failed to create folder %s: %v", parentDir, err)
	}

	if err := os.WriteFile(path, bytes, 0644); err != nil {
		t.Fatalf("Failed to save value %s: %v", path, err)
	}

	return id
}

func FindVarFiles(name string) ([]string, error) {
	var err error
	names := []string{
		"terraform.tfvars",
		name + ".tfvars",
	}
	paths := make([]string, 0, len(names))

	for _, n := range names {
		paths, err = appendVarFile(paths, n)
		if err != nil {
			return nil, err
		}
	}

	return paths, nil
}

func appendVarFile(paths []string, name string) ([]string, error) {
	name, err := resolveVarFile(name)
	if err != nil {
		if os.IsNotExist(err) {
			return paths, nil
		} else {
			return paths, err
		}
	} else {
		paths = append(paths, name)
		return paths, nil
	}
}

func resolveVarFile(name string) (string, error) {
	_, err := os.Stat(name)
	if err != nil {
		return "", err
	}
	return filepath.Abs(name)
}
