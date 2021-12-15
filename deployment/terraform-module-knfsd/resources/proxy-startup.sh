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
set -e

# get_attribute() retrieves an attribute from VM Metadata Server (https://cloud.google.com/compute/docs/metadata/overview)
# @param (str) attribute name
function get_attribute() {
  sleep 1
  curl -sS "http://metadata.google.internal/computeMetadata/v1/instance/attributes/$1" -H "Metadata-Flavor: Google"
}

# build_mount_options() builds the mount options from the GCP metadata attributes
# Do not use this directly when mounting NFS exports, instead use the MOUNT_OPTIONS
# variable. This is because other actions by the script can append additional options
# such as 'fsc'.
function build_mount_options() {
  local -a OPTIONS=(
    rw noatime nocto sync hard ac
    vers=3
    proto=tcp
    mountvers=3
    mountproto=tcp
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
  local -a OPTIONS=(
    rw
    sync
    wdelay
    no_root_squash
    no_all_squash
    no_subtree_check
    sec=sys
    secure
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
  if is_protected_path "$3"; then
    # immediatley terminate so that the proxy does not start with a bad configuration
    echo "ERROR: Cannot mount $1:$2 because $3 a system path"
    exit 1
  fi

  # Make the local export directory
  mkdir -p $3

  # Disable exit on non-zero code and try to mount the NFS Share 3 times 60 seconds apart.
  set +e
  ITER=1
  until [ "$ITER" -ge 4 ]
  do
    echo "(Attempt $ITER/3) Mouting NFS Share: $1:$2..."
    mount -t nfs -o "$MOUNT_OPTIONS" $1:$2 $3
    if [ $? = 0 ]; then
      echo "NFS mount succeeded for $1:$2."
      break
    else
      if [ $ITER = 3 ]; then
        echo "NFS mount failed for $1:$2. Maximum attempts reached, exiting with status 1..."
        exit 1
      fi
      echo "NFS mount failed for $1:$2. Retrying after 60 seconds..."
      sleep 60
    fi
  done
  set -e
}

# add_nfs_export() adds an entry to /etc/exports
# @param (str) Local Directory
# @param (str) Special Options
FSID=10 # Set Initial FSID
function add_nfs_export() {

  echo "Creating NFS share export for $1..."
  echo "$1   $EXPORT_CIDR(${EXPORT_OPTIONS},fsid=${FSID}${2})" >>/etc/exports
  echo "Finished creating NFS share export for $1."

  FSID=$((FSID + 10))

}

PROTECTED_PATHS=(
  # Standard system paths, does not include /media or /mnt as these are valid
  # export targets.
  # Although /home is a common export it's not supported by the proxy as it can
  # cause issues with the GCP infrastructure such as the "gcloud compute ssh"
  # command. GCP would create user directories and create ssh keys in the home
  # directory.
  # The root / is also included but this is handled by is_protected_path
  /             /bin          /boot         /dev          /etc
  /home         /lib          /lib32        /lib64        /libx32
  /lost+found   /opt          /proc         /root         /run
  /sbin         /snap         /srv          /sys          /tmp
  /usr          /var

  /usr/bin      /usr/config   /usr/games    /usr/include  /usr/lib
  /usr/lib32    /usr/lib64    /usr/libexec  /usr/libx32   /usr/local
  /usr/sbin     /usr/share    /usr/src

  /usr/local/include          /usr/local/bin      /usr/local/etc
  /usr/local/games            /usr/local/include  /usr/local/lib
  /usr/local/man              /usr/local/sbin     /usr/local/share
  /usr/local/src

  # Our custom directory, need to look at moving this into either /var or /run
  /statsexport
)

# is_protected_path() checks if a path is a system path
#
# Protected paths are not allowed to be used for mounts and includes paths such
# as /bin. This is used when auto-discovering mounts to avoid accidentally
# mounting over the root filesystem or bin folders.
#
# Sub-directories of protected paths are allowed though as it is valid to have
# an export such as /opt/software, or /var/data.
#
# This is not designed to be fool-proof, and is only aimed to provide basic
# protection for critical folders such as /bin so that the system remains
# usable.
#
# @param (str) The path to check
# @return 0 if the path is protected, non-zero if it is not protected.
function is_protected_path() {
  for p in "${PROTECTED_PATHS[@]}"; do
    if [[ "$1" == "$p" ]] || [[ "$1" == "$p/" ]]; then
      return 0
    fi
  done
  return 1
}

# is_protected_path() checks if a path is in the excluded export list
# @param (str) The path to check
# @return 0 if the path is protected, non-zero if it is not protected.
function is_excluded_export() {
  for p in "${EXCLUDED_EXPORTS[@]}"; do
    if [[ "$1" == "$p" ]] || [[ "$1" == "$p/" ]]; then
      return 0
    fi
  done
  return 1
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
  sed '\|^/$| !s|/+$||'
}

# Get Variables from VM Metadata Server
echo "Reading metadata from metadata server..."

EXPORT_MAP=$(get_attribute EXPORT_MAP)
EXPORT_HOST_AUTO_DETECT=$(get_attribute EXPORT_HOST_AUTO_DETECT)
DISCO_MOUNT_EXPORT_MAP=$(get_attribute DISCO_MOUNT_EXPORT_MAP)
readarray -t EXCLUDED_EXPORTS < <(get_attribute EXCLUDED_EXPORTS | split | trim_slash)
EXPORT_CIDR=$(get_attribute EXPORT_CIDR)
MOUNT_OPTIONS="$(build_mount_options)"
EXPORT_OPTIONS="$(build_export_options)"

NFS_KERNEL_SERVER_CONF=$(get_attribute NFS_KERNEL_SERVER_CONF)
NUM_NFS_THREADS=$(get_attribute NUM_NFS_THREADS)
VFS_CACHE_PRESSURE=$(get_attribute VFS_CACHE_PRESSURE)
DISABLED_NFS_VERSIONS=$(get_attribute DISABLED_NFS_VERSIONS)
READ_AHEAD_KB=$(get_attribute READ_AHEAD_KB)

ENABLE_STACKDRIVER_METRICS=$(get_attribute ENABLE_STACKDRIVER_METRICS)
COLLECTD_METRICS_CONFIG=$(get_attribute COLLECTD_METRICS_CONFIG)
COLLECTD_METRICS_SCRIPT=$(get_attribute COLLECTD_METRICS_SCRIPT)
COLLECTD_ROOT_EXPORT_SCRIPT=$(get_attribute COLLECTD_ROOT_EXPORT_SCRIPT)

CUSTOM_PRE_STARTUP_SCRIPT=$(get_attribute CUSTOM_PRE_STARTUP_SCRIPT)
CUSTOM_POST_STARTUP_SCRIPT=$(get_attribute CUSTOM_POST_STARTUP_SCRIPT)

echo "Done reading metadata."

# Run the CUSTOM_PRE_STARTUP_SCRIPT
echo "Running CUSTOM_PRE_STARTUP_SCRIPT..."
echo "$CUSTOM_PRE_STARTUP_SCRIPT" > /custom-pre-startup-script.sh
chmod +x /custom-pre-startup-script.sh
bash /custom-pre-startup-script.sh
echo "Finished running CUSTOM_PRE_STARTUP_SCRIPT..."

# Set NFS Client Timeout
# Load NFS if it isn't already loaded
modprobe nfs nfs_mountpoint_expiry_timeout=-1
# In case NFS was already loaded set the option directly
sysctl -w fs.nfs.nfs_mountpoint_timeout=-1

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
  MOUNT_OPTIONS="${MOUNT_OPTIONS},fsc"
  echo "FS-Cache started."
else
  echo "No SSD devices(s) found. FS-Cache will remain disabled."
fi

# Truncate the exports file to avoid stale/duplicate exports if the server restarts
: > /etc/exports

echo "Mount options : ${MOUNT_OPTIONS}"
echo "Export options: ${EXPORT_OPTIONS}"

# Loop through $EXPORT_MAP and mount each share defined in the EXPORT_MAP
echo "Beginning processing of standard NFS re-exports (EXPORT_MAP)..."
for i in $(echo $EXPORT_MAP | sed "s/,/ /g"); do

  # Split the components of the entry in EXPORT_MAP
  REMOTE_IP="$(echo $i | cut -d';' -f1)"
  REMOTE_EXPORT="$(echo $i | cut -d';' -f2)"
  LOCAL_EXPORT="$(echo $i | cut -d';' -f3)"

  # Mount the NFS Server export
  mount_nfs_server "$REMOTE_IP" "$REMOTE_EXPORT" "$LOCAL_EXPORT"

  # Create /etc/exports entry for filesystem
  add_nfs_export "$LOCAL_EXPORT" ""

done
echo "Finished processing of standard NFS re-exports (EXPORT_MAP)."

# Loop through $EXPORT_HOST_AUTO_DETECT and detect re-export mount NFS Exports
echo "Beginning processing of dynamically detected host exports (EXPORT_HOST_AUTO_DETECT)..."
for REMOTE_IP in $(echo $EXPORT_HOST_AUTO_DETECT | sed "s/,/ /g"); do

  # Detect the mounts on the NFS Server
  for REMOTE_EXPORT in $(showmount -e --no-headers $REMOTE_IP | awk '{print $1}'); do

    # Mount the NFS Server export
    if is_excluded_export "$REMOTE_EXPORT"; then
      echo "Skipped "$REMOTE_EXPORT", exported was excluded"
    else
      mount_nfs_server "$REMOTE_IP" "$REMOTE_EXPORT" "$REMOTE_EXPORT"
      # Create /etc/exports entry for filesystem
      add_nfs_export "$REMOTE_EXPORT" ""
    fi

  done


done
echo "Finished processing of dynamically detected host exports (EXPORT_HOST_AUTO_DETECT)."

# Loop through $DISCO_MOUNT_EXPORT_MAP and mount each share defined in the DISCO_MOUNT_EXPORT_MAP
echo "Beginning processing of crossmount NFS re-exports (DISCO_MOUNT_EXPORT_MAP)..."
for i in $(echo $DISCO_MOUNT_EXPORT_MAP | sed "s/,/ /g"); do

  # Split the components of the entry in EXPORT_MAP
  REMOTE_IP="$(echo $i | cut -d';' -f1)"
  REMOTE_EXPORT="$(echo $i | cut -d';' -f2)"
  LOCAL_EXPORT="$(echo $i | cut -d';' -f3)"

  # Mount the NFS Server export
  if is_excluded_export "$REMOTE_EXPORT"; then
    echo "Skipped "$REMOTE_EXPORT", exported was excluded"
  else
    mount_nfs_server "$REMOTE_IP" "$REMOTE_EXPORT" "$REMOTE_EXPORT"

    # Discover NFS crossmounts via tree command
    echo "Discovering NFS crossmounts for $REMOTE_IP:$REMOTE_EXPORT..."
    tree -d $LOCAL_EXPORT >/dev/null
    echo "Finished discovering NFS crossmounts for $REMOTE_IP:$REMOTE_EXPORT..."

    # Create an individual export for each crossmount
    for mountpoint in $(df -h | grep $REMOTE_IP:$REMOTE_EXPORT | awk '{print $6}'); do
      add_nfs_export "$mountpoint" ",crossmnt"
    done

  fi

done
echo "Finished processing of crossmount NFS re-exports (DISCO_MOUNT_EXPORT_MAP)."


# Set Read ahead Value to 8 MiB
# Originally read ahead default to rsize * 15, but with rsizes now allowing 1 MiB
# a 15 MiB read ahead was too large. Newer versions of Ubuntu changed the
# default to a fixed value of 128 KiB which is now too small.
# Currently we're assuming the max read size of 1 MiB and using rsize * 8.
echo "Setting read ahead for NFS mounts"
findmnt -rnu -t nfs -o MAJ:MIN,TARGET |
while read MOUNT; do
  DEVICE="$(cut -d ' ' -f 1 <<< "$MOUNT")"
  MOUNT_PATH="$(cut -d ' ' -f 2- <<< "$MOUNT")"
  echo "Setting read ahead for $MOUNT_PATH..."
  echo "$READ_AHEAD_KB" > /sys/class/bdi/"$DEVICE"/read_ahead_kb
done
echo "Finished setting read ahead for NFS mounts"

# Set VFS Cache Pressure
echo "Setting VFS Cache Pressure to $VFS_CACHE_PRESSURE..."
sysctl vm.vfs_cache_pressure=$VFS_CACHE_PRESSURE
echo "Finished setting VFS Cache Pressure."

# Set NFS Kernel Server Config
echo "$NFS_KERNEL_SERVER_CONF" >/etc/default/nfs-kernel-server

# Build Flags to Disable NFS Versions
# Note.... Due to https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=738063 and https://unix.stackexchange.com/questions/205403/disable-nfsv4-server-on-debian-allow-nfsv3
# we have to set these flags under the RPCNFSDCOUNT config value in /etc/default/nfs-kernel-server as RPCMOUNTDOPTS does not work. This should be resolved when we upgrade to
# nfs-utils/1:2.5.4-1~exp1 +
for v in $(echo $DISABLED_NFS_VERSIONS | sed "s/,/ /g"); do
  DISABLED_NFS_VERSIONS_FLAGS="$DISABLED_NFS_VERSIONS_FLAGS --no-nfs-version $v"
done

# Set number of NFS Threads and NFS Server Disable Flags
echo "Setting number of NFS Threads to $NUM_NFS_THREADS..."
sed -i "s/^\(RPCNFSDCOUNT=\).*/\1\"${NUM_NFS_THREADS}${DISABLED_NFS_VERSIONS_FLAGS}\"/" /etc/default/nfs-kernel-server
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

# Start NFS Server
echo "Starting nfs-kernel-server..."
systemctl start portmap
systemctl start nfs-kernel-server
echo "Finished starting nfs-kernel-server..."

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
