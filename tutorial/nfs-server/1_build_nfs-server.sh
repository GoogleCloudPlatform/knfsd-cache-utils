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


# Install NFS kernel server and portmap
echo "Installing NFS kernel server and portmap..."
apt-get update
apt-get install -y nfs-kernel-server portmap
echo "DONE with software install"

# Prep Server by making and exporting a directory and example file
echo "Preparing example file and directory exports"
mkdir /data
dd if=/dev/zero of=/data/test.data count=10240 bs=1048576
chmod a+rw -R /data
exportfs :/data -o rw,sync,no_subtree_check,fsid=10

echo
echo "SUCCESS: Your NFS server is now exporting a 10 gig example file at nfs-server:/data/test.data"
echo -e "${SHELL_DEFAULT}"