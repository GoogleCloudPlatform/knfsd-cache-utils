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
	"net/http"

	"github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-agent/client"
	"github.com/acobaugh/osrelease"
	"golang.org/x/sys/unix"
)

func handleOS(*http.Request) (*client.OSResponse, error) {
	kernel, err := kernelVersion()
	if err != nil {
		return nil, err
	}

	os, err := osrelease.Read()
	if err != nil {
		return nil, err
	}

	return &client.OSResponse{
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
