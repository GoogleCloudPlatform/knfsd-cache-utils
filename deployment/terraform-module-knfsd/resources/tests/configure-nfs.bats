function setup_file() {
	mkdir -p /etc/default
}

function setup() {
	source /opt/bats/lib/bats-support/load.bash
	source /opt/bats/lib/bats-assert/load.bash
	load ./common.bash
	load ../proxy-startup.sh

	NFS_KERNEL_SERVER_CONF="$(<"$BATS_TEST_DIRNAME"/../nfs-kernel-server.conf)"
}

@test "sets cache pressure" {
	VFS_CACHE_PRESSURE=10
	run configure-nfs
	assert_success
	actual="$(</tmp/sysctl)"
	assert_equal "$actual" 'vm.vfs_cache_pressure=10'
}

@test "disable nfs versions" {
	NUM_NFS_THREADS=16
	DISABLED_NFS_VERSIONS="3,4.0,4.2"

	run configure-nfs
	assert_success

	actual="$(find_lines '^RPCNFSDCOUNT=' /etc/default/nfs-kernel-server)"
	assert_equal "$actual" 'RPCNFSDCOUNT="16 --no-nfs-version 3 --no-nfs-version 4.0 --no-nfs-version 4.2"'
}

@test "set RPC thread count" {
	NUM_NFS_THREADS=8
	DISABLED_NFS_VERSIONS=""

	run configure-nfs
	assert_success

	actual="$(find_lines '^RPCNFSDCOUNT=' /etc/default/nfs-kernel-server)"
	assert_equal "$actual" 'RPCNFSDCOUNT="8"'
}
