#######################################
# Get the name of the ssh control file from a logical name.
# Arguments:
#   The logical name of the instance, e.g. proxy, client, etc.
# Outputs:
#   Writes the control file name to stdout
#######################################
control-file() {
	echo ".$1.ssh"
}

#######################################
# Execute a command on a remote instance using the SSH shared connection.
# Arguments:
#   Logical name of the instance.
# Outputs:
#   stdout and stderr from the remote instance prefixed with "name: ".
#######################################
exec-remote() {
	local socket="$(control-file "$1")"
	shift

	# Check if the socket exists in case an invalid name was used.
	# This lets us present a better error message.
	if [[ ! -S "${socket}" ]]; then
		>&2 echo "Control file '${socket}' does not exist"
		return 1
	fi

	# Using a random hostname so that if the shared socket is not present
	# the command will fail. This is because we cannot connect directly to the
	# host as it needs an IAP tunnel running on localhost.
	ssh -S "$socket" fc058006-31d5-11ec-aa77-736a09f48eb7.invalid "$@"
}
