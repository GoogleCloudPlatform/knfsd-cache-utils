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

# Retrieves an attribute from VM Metadata Server
# @param (str) attribute name
function get_attribute() {
  curl -Ss "http://metadata.google.internal/computeMetadata/v1/instance/attributes/$1" -H "Metadata-Flavor: Google"
}

# Get the Hostname, Set a 60 seocnd Interval and get the Load Balancer IP address
HOSTNAME="$(hostname)"
INTERVAL="${COLLECTD_INTERVAL:-60}"
LOADBALANCER_IP_ADDRESS=$(get_attribute LOADBALANCER_IP)

# Loop every 60 seconds
while sleep "$INTERVAL"
do

# Get the number of NFS Clients from Netstat
num_clients=$(netstat -an | grep $LOADBALANCER_IP_ADDRESS:2049 | grep ESTABLISHED | wc -l)
echo "PUTVAL $HOSTNAME/exec-nfs_connections/gauge-usage interval=$INTERVAL N:$num_clients"

# Get the nfs_inode_cache_active_objects from the cached file
nfs_inode_cache_active_objects=$(cat /statsexport/nfs_inode_cache_active_objects)
echo "PUTVAL $HOSTNAME/exec-nfs_inode_cache_active_objects/gauge-usage interval=$INTERVAL N:$nfs_inode_cache_active_objects"

# Get the nfs_inode_cache_objsize from the cached file
nfs_inode_cache_objsize=$(cat /statsexport/nfs_inode_cache_objsize)
echo "PUTVAL $HOSTNAME/exec-nfs_inode_cache_objsize/gauge-usage interval=$INTERVAL N:$nfs_inode_cache_objsize"

# Get the dentry_cache_active_objects from the cached file
dentry_cache_active_objects=$(cat /statsexport/dentry_cache_active_objects)
echo "PUTVAL $HOSTNAME/exec-dentry_cache_active_objects/gauge-usage interval=$INTERVAL N:$dentry_cache_active_objects"

# Get the dentry_cache_objsize from the cached file
dentry_cache_objsize=$(cat /statsexport/dentry_cache_objsize)
echo "PUTVAL $HOSTNAME/exec-dentry_cache_objsize/gauge-usage interval=$INTERVAL N:$dentry_cache_objsize"

done