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

# this expects to be loaded from a setup method
# truncate output files so that they always exist
: >/tmp/sysctl
: >/tmp/systemctl

# replace the standard sysctl with noop function for testing
function sysctl() {
	printf '%s\n' "$*" >>/tmp/sysctl
}

# replace the standard systemctl with noof function for testing
function systemctl() {
	printf '%s\n' "$*" >>/tmp/systemctl
}

function find_lines() {
	local rc
	grep "$1" "$2" && true # suppress -e
	rc=$?
	if (( $rc == 1 )); then
		return 0
	else
		return $rc
	fi
}
