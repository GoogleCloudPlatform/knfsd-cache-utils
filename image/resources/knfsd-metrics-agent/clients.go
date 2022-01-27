package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent/internal/collectd"
)

// TODO: Consider counting unique clients (by IP) as well as total connections
var connectedClients = collectd.NewGauge("nfs_connections", "usage", nil)

func countConnectedClients() (count int, err error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(
		"ss", "--no-header", "--oneline", "--numeric",
		"--tcp", "--udp",
		"state", "established",
		"sport", "2049",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			err = fmt.Errorf("command terminated with exit code %d\n%s", exit.ExitCode(), stderr.String())
		}
		return
	}

	count = 0
	s := bufio.NewScanner(&stdout)
	for s.Scan() {
		count++
	}

	err = s.Err()
	if err != nil {
		return 0, err
	}

	return count, nil
}
