setup() {
	load helpers
}

@test "Proxy is running the correct kernel" {
	local EXPECTED="5.17.0-051700-generic"
	local KERNEL="$(exec-remote proxy -- uname -r)"
	if [[ "${KERNEL}" != "${EXPECTED}" ]]; then
		>&2 echo "Expected kernel version '${EXPECTED}' but was '${KERNEL}'"
		return 1
	fi
}

@test "cachefilesd service is running" {
	exec-remote proxy bash -se <<-'EOT'
		if ! systemctl is-active cachefilesd.service; then
			# Run full status so test error can have more context
			systemctl status cachefilesd.service
			>&2 echo "cachefilesd.service is not enabled"
			exit 1
		fi

		# Even if the service is running cachefilesd can be disabled if
		# RUN=yes is not set in /etc/default/cachefilesd
		if ! grep '^RUN=yes$' /etc/default/cachefilesd; then
			>&2 echo "cachefilesd is disabled, set RUN=yes in /etc/default/cachefilesd"
			exit 1
		fi

		# Check the system logs in case there are other ways cachefilesd can be
		# disabled.
		if journalctl --boot --unit=cachefilesd | grep "cachefilesd disabled"; then
			>&2 echo "cachefilesd is disabled, check journalctl to find out why."
			exit 1
		fi
	EOT
}

@test "fscache is mounted on a separate volume" {
	exec-remote proxy -- mountpoint /var/cache/fscache/
}

@test "Client can read via proxy" {
	# Write the file to the source, and then read it back via the proxy
	exec-remote client -- bash -se <<-'EOT'
		FILE="$(mktemp -u "read.XXXXX")"
		head -c 1024 /dev/urandom | base64 > "/tmp/${FILE}"
		cp "/tmp/${FILE}" "/test/source/${FILE}"
		cmp "/tmp/${FILE}" "/test/proxy/${FILE}"
	EOT
}

@test "Client can write via proxy" {
	# Write the file via the proxy, and then read it back from the source
	exec-remote client -- bash -se <<-'EOT'
		FILE="$(mktemp -u "write.XXXXX")"
		head -c 1024 /dev/urandom | base64 > "/tmp/${FILE}"
		cp "/tmp/${FILE}" "/test/proxy/${FILE}"
		cmp "/tmp/${FILE}" "/test/source/${FILE}"
	EOT
}

@test "Metadata caches positive lookups" {
	exec-remote client -- bash -se <<-'EOT'
		FILE="$(mktemp -u "meta-positive.XXXXX")"
		touch "/test/source/${FILE}"

		BEFORE="$(stat -c '%n %i %F %A %y' "/test/proxy/${FILE}")"
		for i in {1..100}; do
			# A single stat doesn't always cache reliably
			stat /test/proxy/${FILE}
		done

		# Delete the file via the source so that the proxy is unaware
		rm "/test/source/${FILE}"

		# Clear file cache on the client
		sudo sh -c 'sync; echo 3 > /proc/sys/vm/drop_caches'

		AFTER="$(stat -c '%n %i %F %A %y' "/test/proxy/${FILE}")"
		[[ "${BEFORE}" == "${AFTER}" ]]
	EOT
}

@test "Metadata caches negative lookups" {
	exec-remote client -- bash -se <<-'EOT'
		FILE="$(mktemp -u "meta-negative.XXXXX")"
		[[ ! -f "/test/source/${FILE}" ]]

		BEFORE="$(stat -c '%n %i %F %A %y' "/test/proxy/${FILE}" 2>&1 || true)"

		# Cannot rely on errexit here because the stat command is expected
		# to fail. So check that the command failed for the right reason.
		if [[ ! "${BEFORE}" == *"No such file or directory"* ]]; then
			>&2 echo "Expected "
			>&2 echo "    ${BEFORE}"
			>&2 echo 'to contain "No such file or directory"'
			exit 1
		fi
		for i in {1..100}; do
			# A single stat doesn't always cache reliably
			stat /test/proxy/${FILE} 2>&1 || true
		done

		# Create the file via the source so that the proxy is unaware
		touch "/test/source/${FILE}"

		# Clear file cache on the client
		sudo sh -c 'sync; echo 3 > /proc/sys/vm/drop_caches'

		AFTER="$(stat -c '%n %i %F %A %y' "/test/proxy/${FILE}" 2>&1 || true)"
		[[ "${BEFORE}" == "${AFTER}" ]]
	EOT
}

@test "Proxy caches file data (slow)" {
	# This test checks two properties of the cache:
	# * The cache actually caches the file data
	# * The file data is cached on disk using cachefilesd
	# It is not worth separating these two tests out as the setup for them is
	# identical, and working with a 1GB file makes the test comparatively slow.

	# Not using a tool such as fincore as we're explicitly testing that the
	# file data was cached to disk using FS-Cache (cachefilesd).

	# If the cache is full, this test would fail as old data would need to be
	# evicted to store the new data, resulting in a net difference of zero.
	# However, this is not considered an issue as these smoke tests are designed
	# to be run on a new instance of the proxy that was created specifically for
	# running the smoke tests. As such its only likely this will happen when
	# developing the smoke tests where a developer is likely to keep re-running
	# the same tests on the same instance.

	# TODO: Clear proxy cache; specifically /var/cache/fscache

	local INITIAL_SIZE="$(exec-remote proxy -- bash -s <<< "df --output=used -B1M /var/cache/fscache | tail -n 1")"

	exec-remote client bash -se <<-'EOT'
		FILE="$(mktemp -u "large.XXXXX")"

		# Clean up after old tests. Cannot re-use an existing file to save
		# time as it might already be cached.
		rm -f /test/source/large.*

		# Seed large (1G) file for the cache test
		# Create the file locally so that it can be compared after deleting the source.
		dd if=/dev/urandom of="/tmp/smoke-large-file" bs=4M count=1G iflag=count_bytes status=none

		# Copy the file to the source
		cp /tmp/smoke-large-file "/test/source/${FILE}"

		# Read the file through the proxy
		cmp /tmp/smoke-large-file "/test/proxy/${FILE}"

		# Read it a few more times to ensure it is fully cached
		for i in {1..10}; do
			cat /test/proxy/${FILE} > /dev/null
		done

		# Clear file cache on the client
		sudo sh -c 'sync; echo 3 > /proc/sys/vm/drop_caches'

		# Remove the file from the source, then test that the proxy still
		# serves the file from the cache.
		rm "/test/source/${FILE}"
		cmp /tmp/smoke-large-file "/test/proxy/${FILE}"
	EOT

	local CACHE_SIZE="$(exec-remote proxy -- bash -s <<< "df --output=used -B1M /var/cache/fscache | tail -n 1")"
	local DIFF=$(( CACHE_SIZE - INITIAL_SIZE ))

	# Check the cache increased by 1024 Mb, allow a 5% tolerence.
	# TODO: see if /proc/fs/fscache/stats can provider better metrics for this test
	if [[ "${DIFF}" -lt 970 ]] || [[ "${DIFF}" -gt 1080 ]]; then
		>&2 echo "Expected cache delta to be between 970 and 1080, but was ${DIFF}"
		>&2 echo "Initial size: ${INITIAL_SIZE}"
		>&2 echo "Cache size  : ${CACHE_SIZE}"
		return 1
	fi
}
