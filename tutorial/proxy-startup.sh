#!/bin/bash
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

# This script is a simplified version of the main proxy startup script from the
# Terraform module for the purpose of the tutorial. This script has limited
# features and does not support configuring the the source server.

# Exit immediately if a command exits with a non-zero status
set -o errexit
set -o pipefail
shopt -s lastpipe

IP_MASK="10.0.0.0/8"
NFS_SERVER="nfs-server"
NFS_MOUNT_POINT="/data"

function create-fs-cache() {
	# List attatched NVME local SSDs
	echo "Detecting local SSDs drives..."
	DRIVESLIST=$(/bin/find /dev/disk/by-id/ -regex '/dev/disk/by-id/google-local-ssd-[0-9]+$\|/dev/disk/by-id/google-local-nvme-ssd-[0-9]+$')
	NUMDRIVES=$(/bin/find /dev/disk/by-id/ -regex '/dev/disk/by-id/google-local-ssd-[0-9]+$\|/dev/disk/by-id/google-local-nvme-ssd-[0-9]+$' | wc -w)
	echo "Detected $NUMDRIVES drives. Names: $DRIVESLIST."

	# If there are local NVMe drives attached, start the process of formatting and mounting
	if [ $NUMDRIVES -gt 0 ]; then
		echo "Found attached SSD device(s), initializing FS-Cache..."
		if [ ! -e /dev/md127 ]; then
			# Make RAID array of attatched Local SSDs
			echo "Creating RAID array from Local SSDs..."
			mdadm --create /dev/md127 --level=0 --force --quiet --raid-devices=$NUMDRIVES $DRIVESLIST --force
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
		if ! systemctl start cachefilesd; then
			# Sometimes cachefilesd reports an error when starting but does
			# start correctly. This is likely an error in the init script or
			# with how systemd integrates with init scripts.
			# Trying a second time normally works. If you check
			# /proc/fs/fscache/caches the cache is actually active.
			if ! systemctl start cachefilesd; then
				# Second attempt failed, this is now a genuine error so report
				# what went wrong and terminate.
				systemctl status cachefilesd
				exit 1
			fi
		fi

		echo "FS-Cache started."
	else
		echo "No SSD devices(s) found, cannot initialize FS-Cache."
    exit 1
	fi
}

function mount-nfs-server() {
  local remote="$NFS_SERVER:$NFS_MOUNT_POINT"
  local path="/srv/nfs/$NFS_MOUNT_POINT"

  # Make the local export directory
  mkdir -p "$path"

  # In the main terraform script this only attempts 3 times, 60 seconds appart.
  # For the demo, keep trying 15 seconds apart to get faster feedback.
  # The demo is expected to be interactive so this will allow the user time
  # to diagnose and fix the issue.
  local -i attempt
  while true; do
    echo "(Attempt ${attempt}) Mouting NFS Share: $remote..."
    if mount -t nfs -o "$MOUNT_OPTIONS" "$remote" "$path"; then
      echo "NFS mount succeeded for $remote."
      break
    else
      echo "NFS mount failed for $remote. Retrying after 15 seconds..."
      sleep 15
    fi
  done
}

function export-nfs-share() {
  echo "Creating NFS share export."
  echo "$NFS_MOUNT_POINT   $IP_MASK(rw,wdelay,no_root_squash,no_subtree_check,fsid=10,sec=sys,rw,secure,no_root_squash,no_all_squash)" > /etc/exports
  cat /etc/exports
}

function start-nfs() {
	# Start NFS Server
	echo "Starting nfs-kernel-server..."
	if ! systemctl start portmap nfs-kernel-server; then
		systemctl status portmap nfs-kernel-server
		exit 1
	fi
	echo "Finished starting nfs-kernel-server..."
}

create-fs-cache
mount-nfs-server
export-nfs-share
start-nfs
