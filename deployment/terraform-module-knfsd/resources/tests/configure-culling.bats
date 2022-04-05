function setup() {
	source /opt/bats/lib/bats-support/load.bash
	source /opt/bats/lib/bats-assert/load.bash
	load ./common.bash
	load ../proxy-startup.sh

	cp "$BATS_TEST_DIRNAME"/cachefilesd.conf.example /etc/cachefilesd.conf
}

function assert_cull() {
	actual="$(find_lines 'nocull' /etc/cachefilesd.conf)"
	assert_equal "$actual" ""
}

function assert_nocull() {
	actual="$(find_lines 'nocull' /etc/cachefilesd.conf)"
	assert_equal "$actual" "nocull"
}

@test "culling default removes nocull" {
	echo "nocull" >>/etc/cachefilesd.conf

	CULLING=default
	run configure-culling
	assert_success
	assert_cull

	# run a second time to check it handles when nocull is not present
	run configure-culling
	assert_success
	assert_cull

	# check no services were started
	assert_equal "$(cat /tmp/systemctl)" ""
}

@test "culling none sets nocull" {
	CULLING=none

	run configure-culling
	assert_success
	assert_nocull

	# run a second time to check it doesn't add a duplicate nocull
	run configure-culling
	assert_success
	assert_nocull

	# check no services were started
	assert_equal "$(cat /tmp/systemctl)" ""
}

@test "culling custom sets nocull" {
	CULLING=custom

	run configure-culling
	assert_success
	assert_nocull

	# check custom culling agent started
	assert_equal "$(cat /tmp/systemctl)" "start knfsd-cull"

	# run a second time to check it doesn't add a duplicate nocull
	run configure-culling
	assert_success
	assert_nocull
}

@test "culling custom creates config" {
	CULLING=custom
	CULLING_LAST_ACCESS=4h
	CULLING_THRESHOLD=30
	CULLING_INTERVAL=30s
	CULLING_QUIET_PERIOD=1h

	run configure-culling
	assert_success
	diff -u "$BATS_TEST_DIRNAME"/expected/configure-culling/culling-custom-creates-config /etc/knfsd-cull.conf
}

@test "culling default quiet period" {
	CULLING=custom
	CULLING_LAST_ACCESS=4h
	CULLING_THRESHOLD=30
	CULLING_INTERVAL=30s
	CULLING_QUIET_PERIOD=

	run configure-culling
	assert_success
	diff -u "$BATS_TEST_DIRNAME"/expected/configure-culling/culling-default-quiet-period /etc/knfsd-cull.conf
}
