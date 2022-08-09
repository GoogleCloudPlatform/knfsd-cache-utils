#!/usr/bin/env bash

# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Usage: ./run-tests.sh [test]
#   test - Optional, name of a specific *.bats test file in the tests directory.
#          For example "tests/configure-nfs.bats", defaults to "tests".

if ! HASH="$(sha1sum tests/Dockerfile | cut -d ' ' -f 1)"; then
	echo "ERROR: could not create sha1sum for tests/Dockerfile" >&2
	exit 1
fi

BATS_IMAGE=bats:proxy-startup-tests-"$HASH"

if ! docker image inspect "${BATS_IMAGE}" >/dev/null 2>/dev/null; then
	if ! docker build -t "${BATS_IMAGE}" tests; then
		echo "ERROR: could not build docker image" >&2
		exit 1
	fi
fi

docker run --interactive --tty --rm \
	--mount "type=bind,source=$PWD,target=/code,readonly" \
	"${BATS_IMAGE}" "${1-tests}"
