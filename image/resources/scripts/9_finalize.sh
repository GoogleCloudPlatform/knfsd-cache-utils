#!/bin/bash
#
# Copyright 2020 Google LLC
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

set -o errexit
set -o pipefail

# Remove temporary files owned by build user
find /tmp -user build -delete

echo "Verifying file ownership"
# Make sure the build user / group does not own any files on disk we're about
# to image; excluding the home directory since we're about to delete that.
FOUND=$(find / -mount \( -user build -prune -or -group build -prune \) -and -not \( -path ~build -prune \) )
if [[ -n "${FOUND}" ]]; then
	>&2 echo "ERROR: Found files owned by build"
	>&2 echo "${FOUND}"
	exit 1
fi

echo "Removing build user"
userdel -rf build

echo "Shutting down"
shutdown -h now
