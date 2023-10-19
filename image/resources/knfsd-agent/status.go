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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
)

type ServiceHealth client.ServiceHealth

func (sh *ServiceHealth) ReadLog(unit string) {
	log, err := readServiceLog(unit)
	sh.Log = log
	if err != nil {
		sh.Warn("read systemd log", err)
	}
}

func (sh *ServiceHealth) Pass(name string) {
	sh.Add(name, client.CHECK_PASS, nil)
}

func (sh *ServiceHealth) Warn(name string, err error) {
	sh.Add(name, client.CHECK_WARN, err)
}

func (sh *ServiceHealth) Fail(name string, err error) {
	sh.Add(name, client.CHECK_FAIL, err)
}

func (sh *ServiceHealth) Ok(name string, ok bool) {
	health := client.CHECK_FAIL
	if ok {
		health = client.CHECK_PASS
	}
	sh.Add(name, health, nil)
}

func (sh *ServiceHealth) Check(name string, err error) {
	health := client.CHECK_PASS
	if err != nil {
		health = client.CHECK_FAIL
	}
	sh.Add(name, health, err)
}

func (sh *ServiceHealth) Add(name string, health client.Check, err error) {
	if health < sh.Health {
		sh.Health = health
	}

	msg := ""
	if err != nil {
		msg = err.Error()
	}

	sh.Checks = append(sh.Checks, client.ServiceCheck{
		Name:   name,
		Result: health,
		Error:  msg,
	})
}

func handleStatus(*http.Request) (*client.StatusResponse, error) {
	return &client.StatusResponse{
		Services: []client.ServiceHealth{
			cachefilesdStatus(),
		},
	}, nil
}

func cachefilesdStatus() client.ServiceHealth {
	health := ServiceHealth{Name: "cachefilesd", Health: client.CHECK_PASS}
	health.ReadLog("cachefilesd.service")
	health.Check("enabled", checkCachefilesdEnabled())
	health.Check("running", checkCachefilesdRunning())
	health.Check("fscache mounted", checkFSCacheMount())
	return client.ServiceHealth(health)
}

func checkCachefilesdEnabled() error {
	f, err := os.Open("/etc/default/cachefilesd")
	if err != nil {
		return err
	}
	defer f.Close()

	// Look for the last line starting with "RUN="
	var run string
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "RUN=") {
			run = line
		}
	}

	err = s.Err()
	if err != nil {
		return err
	}

	switch run {
	case "":
		return errors.New("'RUN=yes' not found")
	case "RUN=yes":
		return nil
	default:
		return fmt.Errorf("found '%s'; expected 'RUN=yes'", run)
	}
}

func checkCachefilesdRunning() error {
	s, err := readSystemdState("cachefilesd.service")
	if err != nil {
		return err
	}
	if s != "active (running)" {
		return fmt.Errorf("incorrect state, expected active (running) but was %s", s)
	}
	return nil
}

func checkFSCacheMount() error {
	cmd := exec.Command("mountpoint", "--quiet", "/var/cache/fscache")
	err := cmd.Run()
	return err
}

func readPidFile(name string) (int, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(bytes.TrimSpace(data)))
}

func readServiceLog(unit string) (string, error) {
	cmd := exec.Command("journalctl", "--boot", "--lines", "50", "--unit", unit)
	out, err := cmd.Output()
	return string(out), err
}

func readSystemdState(unit string) (string, error) {
	p, err := readSystemdProperties(unit, "ActiveState", "SubState")
	if err != nil {
		return "", fmt.Errorf("systemctl: could not get systemd state of %s: %w", unit, err)
	}
	return fmt.Sprintf("%s (%s)", p["ActiveState"], p["SubState"]), nil
}

func readSystemdProperties(unit string, properties ...string) (map[string]string, error) {
	result := make(map[string]string)

	if len(properties) == 0 {
		return result, nil
	}

	cmd := exec.Command("systemctl", "show",
		"--property", strings.Join(properties, ","),
		unit)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not read systemd properties for %s: %w", unit, err)
	}

	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		result[k] = v
	}

	return result, s.Err()
}
