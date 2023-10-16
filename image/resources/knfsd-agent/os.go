package main

import (
	"net/http"

	"github.com/acobaugh/osrelease"
	"golang.org/x/sys/unix"
)

type OSResponse struct {
	Kernel string            `json:"kernel"`
	OS     map[string]string `json:"os"`
}

func handleOS(*http.Request) (*OSResponse, error) {
	kernel, err := kernelVersion()
	if err != nil {
		return nil, err
	}

	os, err := osrelease.Read()
	if err != nil {
		return nil, err
	}

	return &OSResponse{
		Kernel: kernel,
		OS:     os,
	}, nil
}

func kernelVersion() (string, error) {
	var uts unix.Utsname
	err := unix.Uname(&uts)
	if err != nil {
		return "", err
	}
	return unix.ByteSliceToString(uts.Release[:]), nil
}
