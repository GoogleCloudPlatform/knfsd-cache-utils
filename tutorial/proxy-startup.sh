#!/bin/bash -x
#
# Copyright 2021 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e

# Retrieves VM metadata given the key and default value
# @param (str) key name
# @param (str) default value
function get_metadata {
  local key=$1
  local default=$2
  local url="http://metadata.google.internal/computeMetadata/v1/instance/$key?alt=text"
  local value=$(curl -s -H 'Metadata-Flavor: Google' $url)
  [[ $value =~ .*Error\ 404.* ]] && value=$default
  if [[ -n $value ]]; then
    echo $value
    return 0
  else
    echo ""
    return 1
  fi
}

# Retrieves an attribute from VM metadata
# @param (str) attribute name
function get_attribute() {
  sleep 1
  curl "http://metadata.google.internal/computeMetadata/v1/instance/attributes/$1" -H "Metadata-Flavor: Google"
}

# Retrieves VM IP address from VM metadata
function get_ip_address () {
  curl "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip" -H "Metadata-Flavor: Google"
}

IP_MASK=$(get_ip_address | sed -e 's|\([0-9]*\.[0-9]*\)\..*|\1.0.0/16|')
NFS_SERVER=$(get_attribute NFS_SERVER)
NFS_MOUNT_POINT=$(get_attribute NFS_MOUNT_POINT)

echo "Done reading metadata"

# List attatched NVME local SSDs
DRIVESLIST=`/bin/ls /dev/nvme0n*`
NUMDRIVES=`/bin/ls /dev/nvme0n* | wc -w`

if [ $NUMDRIVES -gt 0 ]
then
  echo "Found attached SSD device(s), initializing FS-Cache..."
  if [ ! -e /dev/md127 ]
  then
    # Make raid array of attatched local ssds
    sudo mdadm --create /dev/md127 --level=0 --force --quiet --raid-devices=$NUMDRIVES $DRIVESLIST --force
  fi

  # Check if disk needs formatting first
  is_formatted=$(fsck -N /dev/md127 | grep ext4 || true)
  if [[ $is_formatted == "" ]]
  then
    mkfs.ext4 -m 0 -F -E lazy_itable_init=0,lazy_journal_init=0,discard /dev/md127
  fi

  mount -o discard,defaults,nobarrier /dev/md127 /var/cache/fscache
  service cachefilesd restart
  FSC=,fsc
  echo "FS-Cache initialized"
else
  echo "No SSD devices(s) found, disabling FS-Cache"
  service cachefilesd stop
  FSC=
fi

echo "Mounting NFS share..."
mkdir -p $NFS_MOUNT_POINT
set +e
while true
do
  mount -t nfs -o vers=3,ac,actimeo=600,noatime,nocto,sync$FSC $NFS_SERVER:$NFS_MOUNT_POINT $NFS_MOUNT_POINT
  if [ $? = 0 ]
  then
    echo "NFS mount succeeded."
    break
  else
    echo "NFS mount failed. Retrying after 15 seconds..."
    sleep 15
  fi
done
set -e

echo "Creating NFS share export."
echo "$NFS_MOUNT_POINT   $IP_MASK(rw,wdelay,no_root_squash,no_subtree_check,fsid=10,sec=sys,rw,secure,no_root_squash,no_all_squash)" > /etc/exports
cat /etc/exports

echo "Starting nfs-kernel-server."
systemctl start portmap
systemctl start nfs-kernel-server
