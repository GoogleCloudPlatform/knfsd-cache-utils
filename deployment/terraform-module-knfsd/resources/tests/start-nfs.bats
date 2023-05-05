function setup() {
	source /opt/bats/lib/bats-support/load.bash
	source /opt/bats/lib/bats-assert/load.bash
	load ./common.bash
	load ../proxy-startup.sh

	cp "$BATS_TEST_DIRNAME"/cachefilesd.conf.example /etc/cachefilesd.conf
}

@test start_basic {
	run start-nfs
	assert_success
	assert_equal "$(cat /tmp/systemctl)" "start portmap nfs-kernel-server"
}

@test start_with_agent {
	ENABLE_KNFSD_AGENT=true
	run start-nfs
	assert_success
	assert_equal "$(cat /tmp/systemctl)" "$(cat <<-EOT
		start knfsd-agent
		start portmap nfs-kernel-server
	EOT
	)"
}

@test start_with_custom_culling {
	CULLING=custom
	run start-nfs
	assert_success
	assert_equal "$(cat /tmp/systemctl)" "$(cat <<-EOT
		start knfsd-cull
		start portmap nfs-kernel-server
	EOT
	)"
}
