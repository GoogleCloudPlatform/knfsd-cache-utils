#!/bin/bash
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
set -o errexit
set -o pipefail
shopt -s lastpipe

# get_attribute() retrieves an attribute from VM Metadata Server (https://cloud.google.com/compute/docs/metadata/overview)
# @param (str) attribute name
function get_attribute() {
  curl -sS "http://metadata.google.internal/computeMetadata/v1/instance/attributes/$1" -H "Metadata-Flavor: Google"
}

# build_mount_options() builds the mount options from the GCP metadata attributes
# Do not use this directly when mounting NFS exports, instead use the MOUNT_OPTIONS
# variable. This is because other actions by the script can append additional options
# such as 'fsc'.
function build_mount_options() {
  local -a OPTIONS=(
    rw noatime nocto async hard ac
    vers="$(get_attribute NFS_MOUNT_VERSION)"
    proto=tcp
    timeo=600
    retrans=2
    lookupcache=all
    local_lock=none
    nconnect="$(get_attribute NCONNECT)"
    acdirmin="$(get_attribute ACDIRMIN)"
    acdirmax="$(get_attribute ACDIRMAX)"
    acregmin="$(get_attribute ACREGMIN)"
    acregmax="$(get_attribute ACREGMAX)"
    rsize="$(get_attribute RSIZE)"
    wsize="$(get_attribute WSIZE)"
  )

  # NFSv4 does not use mountproto, only add this option if using NFSv3
  if [[ "$(get_attribute NFS_MOUNT_VERSION)" == "3" ]]; then
	OPTIONS+=("mountproto=tcp")
  fi

  local EXTRA_OPTIONS="$(get_attribute MOUNT_OPTIONS)"
  if [[ -n "$EXTRA_OPTIONS" ]]; then
    OPTIONS+=("$EXTRA_OPTIONS")
  fi

  local IFS=,
  echo "${OPTIONS[*]}"
}

# build_export_options() builds the common export options for all exports
# Do not use this directly, the result will be cached in EXPORT_OPTIONS
function build_export_options() {
  local HIDE_OPT
  local NOHIDE="$(get_attribute NOHIDE)"
  if [[ $NOHIDE == 'true' ]]; then
    HIDE_OPT="nohide"
  fi

  local -a OPTIONS=(
    rw
    sync
    wdelay
    no_root_squash
    no_all_squash
    no_subtree_check
    sec=sys
    secure
    $HIDE_OPT
  )

  local EXTRA_OPTIONS="$(get_attribute EXPORT_OPTIONS)"
  if [[ -n "$EXTRA_OPTIONS" ]]; then
    OPTIONS+=("$EXTRA_OPTIONS")
  fi

  local IFS=,
  echo "${OPTIONS[*]}"
}

# mount_nfs_server() mounts an NFS Server from the cache
# @param (str) NFS Sever IP
# @param (str) NFS Server Export Path
# @param (str) Local Mount Path
function mount_nfs_server() {
  if [[ -L "$3" ]]; then
    # terminate so that the proxy does not start with a bad configuration
    echo "ERROR: Cannot mount $1:$2 because $3 matches a symlink" >&2
    exit 1
  fi

  local remote="$1:$2"
  local path="/srv/nfs/$3"

  # Make the local export directory
  mkdir -p "$path"

  # try to mount the NFS Share 3 times 60 seconds apart.
  local -i attempt
  for ((attempt=1; ; attempt++)); do
    echo "(Attempt ${attempt}/3) Mouting NFS Share: $remote..."
    if mount -t nfs -o "$MOUNT_OPTIONS" "$remote" "$path"; then
      echo "NFS mount succeeded for $remote."
      break
    else
      if ((attempt >= 3)); then
        echo "NFS mount failed for $remote. Maximum attempts reached, exiting with status 1..."
        exit 1
      fi
      echo "NFS mount failed for $remote. Retrying after 60 seconds..."
      sleep 60
    fi
  done
}

# add_nfs_export() adds an entry to /etc/exports
# @param (str) Local Directory
NEXT_FSID=1
function add_nfs_export() {
  local FSID

  if [[ $1 == / ]]; then
    # Special handling when re-exporting root exports.
    # For NFS v4 the FSID of the root should be set to 0.
    FSID="fsid=0"
  elif [[ $FSID_MODE == "static" ]]; then
    # Statically assign fsid numbers to exports without using the fsidd service.
    # This doesn't really have any advantage over FSID_MODE=local, but can be
    # useful for testing, debugging or troubleshooting.
    FSID="fsid=${NEXT_FSID}"
    NEXT_FSID=$((NEXT_FSID+1))
  else
    # For FSID_MODE local or external use reexport=auto-fsidnum to automatically
	# assign fsid numbers using the fsidd service.
    # When FSID_MODE is set to external this will ensure that all the knfsd
	# instances in a cluster have the same fsid for each export.
    # This is less useful when FSID_MODE is set to local, but it will still
    # ensure the knfsd instance uses the same fsids after a reboot.
    FSID="reexport=auto-fsidnum"
  fi

  echo "Creating NFS share export for $1..."
  echo "$1   $EXPORT_CIDR(${EXPORT_OPTIONS},${FSID})" >>/etc/exports
  echo "Finished creating NFS share export for $1."
}

# rexport() mounts and reexports an NFS share from another NFS server
# @param (str) NFS Server IP
# @param (str) NFS Server Export Path
# @param (str) Local Mount Path
function reexport() {
	mount_nfs_server "$1" "$2" "$3"
	add_nfs_export "$3"
}

# filter_exports() filters exports base on the include and exclude patterns.
# Reads list of exports from stdin and writes the filtered list to stdout.
# Any parameters are passed to the filter-exports command
function filter_exports() {
	filter-exports "$@" \
		-include "${WORKDIR}/include-filters" \
		-exclude "${WORKDIR}/exclude-filters" \
		-verbose
}

# split() splits a list of comma delimited values
# Leading and trailing whitespace is trimmed, empty values are ignored.
# The results are output one item per line.
function split() {
	tr ',' '\n' | sed 's/^\s*//; s/\s*$//' | sed '/^$/d'
}

# trim_slash() removes any trailing slashes from paths
# For example /local/bin/ will be changed to /local/bin.
# A special case is made for / which will be left unchanged.
function trim_slash() {
  sed '\|^/$| !s|/*$||'
}

# start-services() starts one or more services using systemctl.
# If there is an error starting the services systemctl is used to check the
# status and view the most recent log entries.
function start-services() {
	if ! systemctl start "$@"; then
		systemctl status "$@"
		exit 1
	fi
}

function init() {
	# Set any variables cleanup depends upon as blank before setting the trap.
	# This prevents stray environment variables causing unexpected behaviour.
	WORKDIR=
	trap cleanup EXIT

	# Get Variables from VM Metadata Server
	echo "Reading metadata from metadata server..."

	WORKDIR="$(mktemp -d)"
	# get_attribute INCLUDED_EXPORTS | split >"${WORKDIR}/include-filters"
	# get_attribute EXCLUDED_EXPORTS | split >"${WORKDIR}/exclude-filters"
	get_attribute INCLUDED_EXPORTS >"${WORKDIR}/include-filters"
	get_attribute EXCLUDED_EXPORTS >"${WORKDIR}/exclude-filters"

	EXPORT_MAP=$(get_attribute EXPORT_MAP)
	EXPORT_HOST_AUTO_DETECT=$(get_attribute EXPORT_HOST_AUTO_DETECT)
	EXPORT_CIDR=$(get_attribute EXPORT_CIDR)

	FSID_MODE="$(get_attribute FSID_MODE)"
	get_attribute FSID_DATABASE_CONFIG >/etc/knfsd-fsidd.conf

	MOUNT_OPTIONS="$(build_mount_options)"
	EXPORT_OPTIONS="$(build_export_options)"

	NUM_NFS_THREADS=$(get_attribute NUM_NFS_THREADS)
	VFS_CACHE_PRESSURE=$(get_attribute VFS_CACHE_PRESSURE)
	DISABLED_NFS_VERSIONS=$(get_attribute DISABLED_NFS_VERSIONS)
	READ_AHEAD_KB=$(get_attribute READ_AHEAD_KB)

	ENABLE_STACKDRIVER_METRICS=$(get_attribute ENABLE_STACKDRIVER_METRICS)
	METRICS_AGENT_CONFIG=$(get_attribute METRICS_AGENT_CONFIG)
	ENABLE_KNFSD_AGENT=$(get_attribute ENABLE_KNFSD_AGENT)
	ROUTE_METRICS_PRIVATE_GOOGLEAPIS=$(get_attribute ROUTE_METRICS_PRIVATE_GOOGLEAPIS)

	CUSTOM_PRE_STARTUP_SCRIPT=$(get_attribute CUSTOM_PRE_STARTUP_SCRIPT)
	CUSTOM_POST_STARTUP_SCRIPT=$(get_attribute CUSTOM_POST_STARTUP_SCRIPT)

	# Auto-discovery of exports using NetApp API.
	# Need to be exported so that the netapp-exports tool can read them from
	# the local environment.
	export ENABLE_NETAPP_AUTO_DETECT="$(get_attribute ENABLE_NETAPP_AUTO_DETECT)"
	export NETAPP_HOST="$(get_attribute NETAPP_HOST)"
	export NETAPP_URL="$(get_attribute NETAPP_URL)"
	export NETAPP_USER="$(get_attribute NETAPP_USER)"
	export NETAPP_SECRET="$(get_attribute NETAPP_SECRET)"
	export NETAPP_SECRET_PROJECT="$(get_attribute NETAPP_SECRET_PROJECT)"
	export NETAPP_SECRET_VERSION="$(get_attribute NETAPP_SECRET_VERSION)"
	export NETAPP_CA="$(get_attribute NETAPP_CA)"
	export NETAPP_ALLOW_COMMON_NAME="$(get_attribute NETAPP_ALLOW_COMMON_NAME)"

	# NetApp CA certificate needs to be stored in a file
	if [[ -n "$NETAPP_CA" ]]; then
		echo "$NETAPP_CA" >"${WORKDIR}/netapp-ca.pem"
		export NETAPP_CA="${WORKDIR}/netapp-ca.pem"
	fi

	echo "Done reading metadata."

	# Truncate the exports file to avoid stale/duplicate exports if the server restarts
	: > /etc/exports

	# Run the CUSTOM_PRE_STARTUP_SCRIPT
	echo "Running CUSTOM_PRE_STARTUP_SCRIPT..."
	echo "$CUSTOM_PRE_STARTUP_SCRIPT" > /custom-pre-startup-script.sh
	chmod +x /custom-pre-startup-script.sh
	bash /custom-pre-startup-script.sh
	echo "Finished running CUSTOM_PRE_STARTUP_SCRIPT..."
}

function create-fs-cache() {

	# Check if we are setting up Local SSD's or Persistent Disks
	if [[ -L "/dev/disk/by-id/google-pd-fscache" ]]; then

		echo "Detected a Persistent Disk for FS-Cache..."

		echo "Checking if Persistent Disk needs formatting..."
		is_formatted=$(fsck -N /dev/disk/by-id/google-pd-fscache | grep ext4 || true)
		if [[ $is_formatted == "" ]]; then
			echo "Persistent Disk is not formatted. Formatting..."
			mkfs.ext4 -m 0 -E lazy_itable_init=0,lazy_journal_init=0,discard /dev/disk/by-id/google-pd-fscache
			echo "Finished formatting Persistent Disk."
		else
			echo "Persistent Disk is already formatted."
		fi

		echo "Mounting /dev/disk/by-id/google-pd-fscache to FS-Cache directory (/var/cache/fscache)..."
		mount -o discard,defaults /dev/disk/by-id/google-pd-fscache /var/cache/fscache
		echo "Finished mounting /dev/disk/by-id/google-pd-fscache to FS-Cache directory (/var/cache/fscache)"

		# Start FS-Cache
		start-fs-cache

	else

		echo "No Persistent Disk detected for FS-Cache, assuming Local SSDs are present..."

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
			echo "Finished mounting /dev/md127 to FS-Cache directory (/var/cache/fscache)"

			# Start FS-Cache
			start-fs-cache

		else
			echo "No SSD devices(s) found. FS-Cache will remain disabled."
		fi

	fi

}

function start-fs-cache() {

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

	MOUNT_OPTIONS="${MOUNT_OPTIONS},fsc"
	echo "FS-Cache started."

}

function start-fsidd() {
	case "$FSID_MODE" in
	static)
		echo "Skipping fsidd service"
		;;
	local)
		echo "Starting fsidd..."
		start-services fsidd.service
		echo "Finished Starting fsidd."
		;;
	external)
		echo "Starting knfsd-fsidd..."
		start-services knfsd-fsidd.socket knfsd-fsidd.service
		echo "Finished Starting knfsd-fsidd."
		;;
	*)
		echo "Unknown FSID_MODE \"$FSID_MODE\"."
		exit 1
		;;
	esac
}

function export-map() {
	# Loop through $EXPORT_MAP and mount each share defined in the EXPORT_MAP
	echo "Beginning processing of standard NFS re-exports (EXPORT_MAP)..."

	for i in $(echo $EXPORT_MAP | sed "s/,/ /g"); do
		# Split the components of the entry in EXPORT_MAP
		REMOTE_IP="$(echo $i | cut -d';' -f1)"
		REMOTE_EXPORT="$(echo $i | cut -d';' -f2)"
		LOCAL_EXPORT="$(echo $i | cut -d';' -f3)"
		reexport "$REMOTE_IP" "$REMOTE_EXPORT" "$LOCAL_EXPORT"
	done

	echo "Finished processing of standard NFS re-exports (EXPORT_MAP)."
}

function export-auto-detect() {
	# Loop through $EXPORT_HOST_AUTO_DETECT and detect re-export mount NFS Exports
	echo "Beginning processing of dynamically detected host exports (EXPORT_HOST_AUTO_DETECT)..."

	for REMOTE_IP in $(echo $EXPORT_HOST_AUTO_DETECT | sed "s/,/ /g"); do
		# Detect the mounts on the NFS Server
		showmount -e --no-headers $REMOTE_IP | filter_exports -field 1 | awk '{print $1}' | sort |
		while read -r REMOTE_EXPORT; do
			# Mount the NFS Server export
			reexport "$REMOTE_IP" "$REMOTE_EXPORT" "$REMOTE_EXPORT"
		done
	done

	echo "Finished processing of dynamically detected host exports (EXPORT_HOST_AUTO_DETECT)."
}

function export-netapp() {
	if [[ "$ENABLE_NETAPP_AUTO_DETECT" == "true" ]]; then
		echo "Beginning processing of dynamically detected NetApp exports (ENABLE_NETAPP_AUTO_DETECT)..."

		netapp-exports | filter_exports -field 2 |
		while read -r REMOTE; do
			REMOTE_IP="$(echo "$REMOTE" | cut -d ' ' -f1)"
			REMOTE_EXPORT="$(echo "$REMOTE" | cut -d ' ' -f2-)"

			# Mount the NFS Server export
			reexport "$REMOTE_IP" "$REMOTE_EXPORT" "$REMOTE_EXPORT"
		done

		echo "Finished processing of dynamically detected NetApp exports (ENABLE_NETAPP_AUTO_DETECT)."
	fi
}

function configure-read-ahead() {
	# Set Read ahead Value to 8 MiB
	# Originally read ahead default to rsize * 15, but with rsizes now allowing 1 MiB
	# a 15 MiB read ahead was too large. Newer versions of Ubuntu changed the
	# default to a fixed value of 128 KiB which is now too small.
	# Currently we're assuming the max read size of 1 MiB and using rsize * 8.
	echo "Setting read ahead for NFS mounts..."

	findmnt -rnu -t nfs,nfs4 -o MAJ:MIN,TARGET |
	while read -r MOUNT; do
		DEVICE="$(cut -d ' ' -f 1 <<< "$MOUNT")"
		MOUNT_PATH="$(cut -d ' ' -f 2- <<< "$MOUNT")"
		echo "Setting read ahead for $MOUNT_PATH..."
		echo "$READ_AHEAD_KB" > /sys/class/bdi/"$DEVICE"/read_ahead_kb
	done
	echo "Finished setting read ahead for NFS mounts"
}

function configure-nfs() {
	# Set VFS Cache Pressure
	echo "Setting VFS Cache Pressure to $VFS_CACHE_PRESSURE..."
	sysctl vm.vfs_cache_pressure=$VFS_CACHE_PRESSURE
	echo "Finished setting VFS Cache Pressure."

	# Build Flags to Disable NFS Versions
	DISABLED_NFS_VERSIONS_FLAGS=("vers2=no")
	for v in $(echo $DISABLED_NFS_VERSIONS | sed "s/,/ /g"); do
		DISABLED_NFS_VERSIONS_FLAGS+=("vers$v=no")
	done
	DISABLED_NFS_VERSIONS_FLAGS=$(printf '%s\n' "${DISABLED_NFS_VERSIONS_FLAGS[@]}")

	# Set NFS Kernel Server Config
	echo "Setting number of NFS Threads to $NUM_NFS_THREADS..."
	cat <<-EOF >/etc/nfs.conf.d/knfsd.conf
		[nfsd]
		threads=${NUM_NFS_THREADS}
		${DISABLED_NFS_VERSIONS_FLAGS}
	EOF
	echo "Finished setting number of NFS Threads."

}

function configure-metrics() {

	# If needed, override the Monitoring API to use an IP address from private.googleapis.com
	if [ "$ROUTE_METRICS_PRIVATE_GOOGLEAPIS" = "true" ]; then
		echo "Enabling metrics.googleapis.com routing via private.googleapis.com..."
		grep -qxF '199.36.153.11 monitoring.googleapis.com' /etc/hosts || echo '199.36.153.11 monitoring.googleapis.com' >> /etc/hosts
		grep -qxF '199.36.153.11 cloudtrace.googleapis.com' /etc/hosts || echo '199.36.153.11 cloudtrace.googleapis.com' >> /etc/hosts
		grep -qxF '199.36.153.11 logging.googleapis.com' /etc/hosts || echo '199.36.153.11 logging.googleapis.com' >> /etc/hosts
		echo "Finished enabling metrics.googleapis.com routing via private.googleapis.com."
	else
		sed -i '/199.36.153.11 monitoring.googleapis.com/d' /etc/hosts
		sed -i '/199.36.153.11 cloudtrace.googleapis.com/d' /etc/hosts
		sed -i '/199.36.153.11 logging.googleapis.com/d' /etc/hosts
	fi

	# Enable Metrics if Configured
	if [ "$ENABLE_STACKDRIVER_METRICS" = "true" ]; then
		echo "Starting Metrics Agents..."
		printf '%s' "$METRICS_AGENT_CONFIG" >/etc/knfsd-metrics-agent/custom.yaml
		systemctl start google-cloud-ops-agent knfsd-metrics-agent
		echo "Finished starting Metrics Agents."
	else
		echo "Metrics are disabled. Skipping..."
	fi
}

function start-nfs() {
	# Enable Knfsd Agent if Configured
	if [[ "$ENABLE_KNFSD_AGENT" = "true" ]]; then
		echo "Starting Knfsd Agent..."
		start-services knfsd-agent
		echo "Finished Starting Knfsd Agent."
	else
		echo "Knfsd Agent disabled. Skipping..."
	fi

	# Start NFS Server
	echo "Starting nfs-kernel-server..."
	start-services portmap nfs-kernel-server
	echo "Finished starting nfs-kernel-server..."
}

function post-startup() {
	# Run the CUSTOM_POST_STARTUP_SCRIPT
	echo "Running CUSTOM_POST_STARTUP_SCRIPT..."
	echo "$CUSTOM_POST_STARTUP_SCRIPT" > /custom-post-startup-script.sh
	chmod +x /custom-post-startup-script.sh
	bash /custom-post-startup-script.sh
	echo "Finished running CUSTOM_POST_STARTUP_SCRIPT..."

	echo "NFS Mounts"
	findmnt -ut nfs

	echo "NFS Exports"
	exportfs -s

	echo "Reached Proxy Startup Exit. Happy caching!"
}

function cleanup() {
	# Do not call this explicitly, the init function sets an exit trap
	if [[ -n "${WORKDIR}" ]] && [[ -d "${WORKDIR}" ]]; then
		rm -rf "${WORKDIR}"
	fi
}

function main() {
	init
	create-fs-cache

	echo "Mount options : ${MOUNT_OPTIONS}"
	echo "Export options: ${EXPORT_OPTIONS}"

	export-map
	export-auto-detect
	export-netapp

	configure-read-ahead
	configure-nfs
	configure-metrics

	start-fsidd
	start-nfs
	post-startup
}

# Do not execute the main function if this script has been loaded by bats for
# unit testing.
if [[ -z "${BATS_VERSION}" ]]; then
	main
fi
