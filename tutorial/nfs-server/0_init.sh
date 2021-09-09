#!/bin/bash
#
# Copyright 2021 Google LLC
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


# get_script() fetches a key from the metadata server and writes it to a file
function get_script() {
    curl -Ss -o "/root/$2" "http://metadata.google.internal/computeMetadata/v1/instance/attributes/$1" -H "Metadata-Flavor: Google"
}

# Fetch scripts, make executable and run
get_script BUILD_NFS_SERVER_SCRIPT 1_build_nfs-server.sh
chmod +x /root/1_build_nfs-server.sh
/root/1_build_nfs-server.sh
# Done
