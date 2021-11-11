#!/usr/bin/env bash

# Not setting errexit as this creates inconsistencies in how the functions
# handle errors. If a function is called from an "if" clause the entire
# function executes in a context where errexit is disabled.
# This results in the behaviour of code inside a function being inconsistent.
# Depending on the context the same function can execute with errexit enabled
# or disabled simply by calling the function from an if clause.

set -o pipefail
set -o nounset

usage() {
	>&2 cat <<'EOT'
To run this script set the following environment variables:
  PROJECT
  ZONE
    The GCP project and zone

  PROXY_INSTANCE
  PROXY_INSTANCE_GROUP
    The proxy instance or group (MIG) to connect to.
	For the smoke tests the MIG must only have a single instance, otherwise
	its not possible to tell which instance the client will connect to.

  CLIENT_INSTANCE
    GCP name of the client instance.

  NFS_SOURCE
    Share to mount for the NFS source in the format "host_or_ip:/path/to/volume".

  NFS_PROXY
    Share to mount for the NFS proxy in the format "host_or_ip:/path/to/volume".

  TEST_PATH
    Path to use on the NFS shares for test files. This must be the same
	path relative to both shares.

  BATS_IMAGE
    Name of the bats docker image to run.
    Run build.sh to create an image named "bats:knfsd-test".
EOT
	exit 1
}

#######################################
# Checks if an environment variable has a value
# Arguments:
#   Name of the environment variable to check
# Outputs:
#   Writes an error message to stderr if the environment variable does
#   not have a value.
# Returns:
#   0 if the environment variable has a value, non-zero on error
#######################################
require() {
	if [[ -z "${!1:-}" ]]; then
		>&2 echo "ERROR: '${1}' required"
		return 1
	else
		return 0
	fi
}

#######################################
# Validate and initialize the input variables.
# Globals
#   PROJECT
#   ZONE
#   PROXY_INSTANCE
#   PROXY_INSTANCE_GROUP
#   CLIENT_INSTANCE
#   NFS_SOURCE
#   NFS_PROXY
#   TEST_PATH
#   BATS_IMAGE
# Outputs:
#   Writes error messages to stderr
# Returns
#   0 if all the inputs are valid, non-zero on error.
#######################################
initialize() {
	local error=0

	require PROJECT || error=1
	require ZONE || error=1
	require CLIENT_INSTANCE || error=1

	if [[ -z "${PROXY_INSTANCE:-}" ]]; then
		if [[ -z "${PROXY_INSTANCE_GROUP:-}" ]]; then
			>&2 echo "ERROR: 'PROXY_INSTANCE' or 'PROXY_INSTANCE_GROUP' required"
			error=1
		else
			PROXY_INSTANCE="$(resolve-proxy-instance)"
			if [[ -z "${PROXY_INSTANCE}" ]]; then
				>&2 echo "ERORR: Could not resolve PROXY_INSTANCE"
				error=1
			else
				>&2 echo "Resolved PROXY_INSTANCE as ${PROXY_INSTANCE}"
			fi
		fi
	fi

	require NFS_SOURCE || error=1
	require NFS_PROXY || error=1
	require TEST_PATH || error=1

	require BATS_IMAGE || error=1

	return ${error}
}

#######################################
# Get a random instance from proxy MIG.
# Globals:
#   PROJECT
#   ZONE
#   PROXY_INSTANCE_GROUP
# Outputs:
#   The name of an instance from the proxy MIG
#######################################
resolve-proxy-instance() {
	gcloud compute instance-groups list-instances \
		--project="${PROJECT}" \
		--zone="${ZONE}" \
		"${PROXY_INSTANCE_GROUP}" \
		--limit=1 \
		--format='value(instance)'
}

#######################################
# Globals:
#   PROJECT
#   ZONE
#   SSH_PIDS
# Arguments:
#   Logical name of the instance
#   Name of the GCP Compute Instance to connect to
#######################################
start-ssh() {
	local socket="$(control-file "$1")"

	if [[ -S "${socket}" ]]; then
		>&2 echo "Control file '${socket}' already exists."
		return 1
	fi

	gcloud compute ssh \
		--project=$PROJECT \
		--zone=$ZONE \
		"$2" \
		-- -NMS "${socket}" \
		1> >(sed "s/^/(ssh) $1: /") \
		2> >(sed "s/^/(ssh) $1: /" >&2) &

	SSH_PIDS+=($!)
}

#######################################
# Wait until the SSH control file is created
# Arguments:
#   Logical name of the instance.
# Returns:
#   0 if the master control file was created, non-zero on error.
#######################################
wait-ssh() {
	local socket="$(control-file "$1")"
	local start="$(date +%s)"

	while [[ ! -S "${socket}" ]]; do
		sleep 1
		local now="$(date +%s)"
		if (( now - start > 30 )); then
			break
		fi
	done

	# Set the exit code. Final test to see if the socket exists
	[[ -S "${socket}" ]]
}

#######################################
# Stop all the SSH master instances.
# Global:
#   SSH_PIDS
#######################################
stop-ssh() {
	if (( ${#SSH_PIDS[@]} > 0 )); then
		kill -s SIGTERM "${SSH_PIDS[@]}"
	fi
}

#######################################
# Mount the NFS shares for testing.
# Global:
#   NFS_SOURCE
#	NFS_PROXY
#	TEST_PATH
# Outputs:
#   stdout and stderr from the client.
#######################################
mount-shares() {
	>&2 echo "Mounting NFS shares on client"
	exec-remote client sudo bash -es -- "${NFS_SOURCE}" "${NFS_PROXY}" "${TEST_PATH}" \
		1> >(sed "s/^/client: /") \
		2> >(sed "s/^/client: /" >&2) \
		<<-'EOT'
			apt-get install -qq --no-install-recommends nfs-common

			# Unmount the shares if they're already mounted
			mountpoint -q /mnt/source && umount -f /mnt/source
			mountpoint -q /mnt/proxy && umount -f /mnt/proxy

			# Create the directories for the mounts.
			# Mark them as immutable so that if the mount fails no data
			# can be accidentally wrote to the local disk.
			mkdir -p /mnt/source /mnt/proxy
			chattr +i /mnt/source /mnt/proxy

			# Disable as much caching in the client as reasonable so that the
			# client has to requery the proxy during the smoke tests.
			mount "$1" /mnt/source -o vers=3,proto=tcp,noac,noatime,nocto,rsize=1048576,wsize=1048576,lookupcache=none,nolock
			mount "$2" /mnt/proxy -o vers=3,proto=tcp,noac,noatime,nocto,rsize=1048576,wsize=1048576,lookupcache=none,nolock

			# Create the test directory if it doesn't already exist and
			# assign ownership to our test user.
			mkdir -p "/mnt/source/$3"
			rm -f "/mnt/source/$3/*"
			chown "$SUDO_UID:$SUDO_GID" "/mnt/source/$3"
			chmod 775 "/mnt/source/$3"

			# Add symlinks to the test directories to simplify the tests.
			# This avoids needing to parameterise every test with a path.
			mkdir -p /test
			ln -snf /mnt/source/$3 /test/source
			ln -snf /mnt/proxy/$3 /test/proxy
		EOT
}

#######################################################################

source helpers.bash

if ! initialize; then
	usage
	exit 1
fi

#######################################################################
# Start SSH connections
# Using the SSH master control feature so that we can re-use the same SSH
# connection across multiple commands.

SSH_PIDS=()
trap stop-ssh EXIT

if ! start-ssh proxy "${PROXY_INSTANCE}"; then
	>&2 echo "ERROR: Could not connect to proxy"
	exit 1
fi

if ! start-ssh client "${CLIENT_INSTANCE}"; then
	>&2 echo "ERROR: Could not connect to client"
	exit 1
fi

if ! wait-ssh proxy; then
	>&2 echo "ERROR: Could not connect to proxy"
	exit 1
fi

if ! wait-ssh client; then
	>&2 echo "ERROR: Could not connect to client"
	exit 1
fi

#######################################################################

if ! mount-shares; then
	>&2 echo "ERROR: Failed to mount shares on the client"
	exit 1
fi

>&2 echo "Setup Complete"
>&2 echo "───────────────────────────────────────────────────────────────────────"
>&2 echo "Running Tests"

docker run --interactive --tty --rm \
	--mount "type=bind,source=$PWD,target=/code,readonly" \
	"${BATS_IMAGE}" .
