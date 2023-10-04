package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type ExitError struct {
	cause  error
	stderr string
}

func (ee ExitError) String() string {
	return ee.Error()
}

func (ee ExitError) Error() string {
	return fmt.Sprintf("%s\n%s", ee.cause.Error(), ee.stderr)
}

func (ee ExitError) Stderr() string {
	return ee.stderr
}

func Sudo(script string) error {
	cmd := exec.Command("sudo", "/bin/bash", "-s")
	cmd.Stdin = strings.NewReader(script)
	stderr, err := cmd.CombinedOutput()
	if err != nil {
		err = ExitError{cause: err, stderr: string(stderr)}
	}
	return err
}
