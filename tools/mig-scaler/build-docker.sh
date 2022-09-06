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

IMAGE=golang:1.17

cd "$(dirname "$0")"

rm -f ./mig-scaler

if ! CONTAINER=$(docker create --interactive "${IMAGE}")
then
	>&2 echo 'Failed to create golang container'
	exit 1
fi

docker cp ./ "${CONTAINER}":/go/src/

docker start --interactive --attach "${CONTAINER}" <<-EOF
	cd /go/src
	export CGO_ENABLED=0
	export GO11MODULE=on
	go build -o /go/bin/
EOF

docker cp \
	"${CONTAINER}:/go/bin/mig-scaler" \
	"./mig-scaler"

docker rm "${CONTAINER}" > /dev/null
