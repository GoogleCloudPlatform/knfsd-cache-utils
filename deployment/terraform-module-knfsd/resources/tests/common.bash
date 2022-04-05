# this expects to be loaded from a setup method
# truncate output files so that they always exist
: >/tmp/sysctl
: >/tmp/systemctl

# replace the standard sysctl with noop function for testing
function sysctl() {
	printf '%s\n' "$*" >>/tmp/sysctl
}

# replace the standard systemctl with noof function for testing
function systemctl() {
	printf '%s\n' "$*" >>/tmp/systemctl
}

function find_lines() {
	local rc
	grep "$1" "$2" && true # suppress -e
	rc=$?
	if (( $rc == 1 )); then
		return 0
	else
		return $rc
	fi
}
