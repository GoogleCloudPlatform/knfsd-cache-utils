#!/bin/bash -x
#
# Copyright 2020 Google Inc.
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

# Exit immediately if a command exits with a non-zero status
set -e

# Retrieves an attribute from VM Metadata Server
# @param (str) attribute name
function get_attribute() {
  sleep 1
  curl -sS "http://metadata.google.internal/computeMetadata/v1/instance/attributes/$1" -H "Metadata-Flavor: Google"
}

# Get Variables from VM Metadata Server
echo "Reading metadata from metadata server..."
EXPORT_MAP=$(get_attribute EXPORT_MAP)
EXPORT_CIDR=$(get_attribute EXPORT_CIDR)
NCONNECT_VALUE=$(get_attribute NCONNECT_VALUE)
VFS_CACHE_PRESSURE=$(get_attribute VFS_CACHE_PRESSURE)
NUM_NFS_THREADS=$(get_attribute NUM_NFS_THREADS)
ENABLE_STACKDRIVER_METRICS=$(get_attribute ENABLE_STACKDRIVER_METRICS)
COLLECTD_METRICS_CONFIG=$(get_attribute COLLECTD_METRICS_CONFIG)
COLLECTD_METRICS_SCRIPT=$(get_attribute COLLECTD_METRICS_SCRIPT)
COLLECTD_ROOT_EXPORT_SCRIPT=$(get_attribute COLLECTD_ROOT_EXPORT_SCRIPT)
NFS_KERNEL_SERVER_CONF=$(get_attribute NFS_KERNEL_SERVER_CONF)
echo "Done reading metadata."

# List attatched NVME local SSDs
echo "Detecting local NVMe drives..."
DRIVESLIST=$(/bin/ls /dev/nvme0n*)
NUMDRIVES=$(/bin/ls /dev/nvme0n* | wc -w)
echo "Detected $NUMDRIVES drives. Names: $DRIVESLIST."

# If there are local NVMe drives attached, start the process of formatting and mounting
if [ $NUMDRIVES -gt 0 ]; then
  echo "Found attached SSD device(s), initializing FS-Cache..."
  if [ ! -e /dev/md127 ]; then
    # Make RAID array of attatched Local SSDs
    echo "Creating RAID array from Local SSDs..."
    sudo mdadm --create /dev/md127 --level=0 --force --quiet --raid-devices=$NUMDRIVES $DRIVESLIST --force
    echo "Finished creating RAID array from Local SSDs."
  fi

  # Check if the RAID array has already been formatted
  echo "Checking if RAID array needs formatting..."
  is_formatted=$(fsck -N /dev/md127 | grep ext4 || true)
  if [[ $is_formatted == "" ]]; then
    echo "RAID array is not formatted. Formatting..."
    mkfs.ext4 -m 0 -F -E lazy_itable_init=0,lazy_journal_init=0,discard /dev/md127
    echo "Finished formatting RAID array."
  else
    echo "RAID array is already formatted."
  fi

  # Mount /dev/md127 to /var/cache/fscache
  echo "Mounting /dev/md127 to FS-Cache directory (/var/cache/fscache)..."
  mount -o discard,defaults,nobarrier /dev/md127 /var/cache/fscache
  echo "Finished /dev/md127 to FS-Cache directory (/var/cache/fscache)"

  # Start FS-Cache
  echo "Starting FS-Cache..."
  systemctl start cachefilesd
  FSC=,fsc
  echo "FS-Cache started."
else
  echo "No SSD devices(s) found. FS-Cache will remain disabled."
  FSC=
fi

# Set the FSID
FSID=10

# Loop through $EXPORT_MAP and mount each share defined in the EXPORT_MAP
for i in $(echo $EXPORT_MAP | sed "s/,/ /g"); do

  # Split the components of the entry in EXPORT_MAP
  REMOTE_IP="$(echo $i | cut -d';' -f1)"
  REMOTE_EXPORT="$(echo $i | cut -d';' -f2)"
  LOCAL_EXPORT="$(echo $i | cut -d';' -f3)"

  # Make the local export directory
  mkdir -p $LOCAL_EXPORT

  # Disable exit on non-zero code and continuously try to mount the NFS Share. If this takes too long we will be replaced by the mig.
  set +e
  while true; do
    echo "Attempting to mount NFS Share: $REMOTE_IP:$REMOTE_EXPORT..."
    mount -t nfs -o vers=3,ac,actimeo=600,noatime,nocto,nconnect=$NCONNECT_VALUE,sync,$FSC $REMOTE_IP:$REMOTE_EXPORT $LOCAL_EXPORT
    if [ $? = 0 ]; then
      echo "NFS mount succeeded for $REMOTE_IP:$REMOTE_EXPORT."
      break
    else
      echo "NFS mount failed for $REMOTE_IP:$REMOTE_EXPORT. Retrying after 15 seconds..."
      sleep 15
    fi
  done
  set -e

  # Create /etc/exports entry for filesystem
  echo "Creating NFS share export for $REMOTE_IP:$REMOTE_EXPORT..."
  echo "$LOCAL_EXPORT   $EXPORT_CIDR(rw,wdelay,no_root_squash,no_subtree_check,fsid=$FSID,sec=sys,rw,secure,no_root_squash,no_all_squash)" >>/etc/exports
  echo "Finished creating NFS share export for $REMOTE_IP:$REMOTE_EXPORT."

  # Increment FSID
  FSID=$((FSID + 10))

done

# Set VFS Cache Pressure
echo "Setting VFS Cache Pressure to $VFS_CACHE_PRESSURE..."
sysctl vm.vfs_cache_pressure=$VFS_CACHE_PRESSURE
echo "Finished setting VFS Cache Pressure."

# Set NFS Kernel Server Config
echo "$NFS_KERNEL_SERVER_CONF" > /etc/default/nfs-kernel-server

# Set number of NFS Threads
echo "Setting number of NFS Threads to $NUM_NFS_THREADS..."
sed -i "s/^\(RPCNFSDCOUNT=\).*/\1${NUM_NFS_THREADS}/" /etc/default/nfs-kernel-server
echo "Finished setting number of NFS Threads."

# Enable Metrics if Configured
if [ "$ENABLE_STACKDRIVER_METRICS" = "true" ]; then

  echo "Configuring Stackdriver metrics..."
  echo "$COLLECTD_METRICS_SCRIPT" >/etc/stackdriver/knfsd-export.sh
  chmod +x /etc/stackdriver/knfsd-export.sh
  echo "$COLLECTD_METRICS_CONFIG" >/etc/stackdriver/collectd.d/knfsd.conf
  echo "Finished configuring Stackdriver metrics..."

  echo "Setting up root export script..."
  echo "$COLLECTD_ROOT_EXPORT_SCRIPT" >/collectd-root-export-script.sh
  chmod +x /collectd-root-export-script.sh
  nohup /collectd-root-export-script.sh &
  echo "Finished setting up root export script..."

  echo "Starting Stackdriver Agent..."
  systemctl start stackdriver-agent
  echo "Finished starting Stackdriver Agent."

else
  echo "Metrics are disabled. Skipping..."
fi


echo "Starting nfs-kernel-server..."
systemctl start portmap
systemctl start nfs-kernel-server
echo "Finished starting nfs-kernel-server..."


echo "Reached Proxy Startup Exit. Happy caching!"
