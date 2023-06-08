function setup_file() {
	: >/etc/nfs.conf
	rm -rf /etc/nfs.conf.d
	mkdir -p /etc/nfs.conf.d
}

function setup() {
	source /opt/bats/lib/bats-support/load.bash
	source /opt/bats/lib/bats-assert/load.bash
	load ./common.bash
	load ../proxy-startup.sh
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

	assert [ -f /etc/nfs.conf.d/knfsd.conf ]
	assert_equal "$(nfsconf --get nfsd vers2)" no
	assert_equal "$(nfsconf --get nfsd vers3)" no
	assert_equal "$(nfsconf --get nfsd vers4)" "" # not set
	assert_equal "$(nfsconf --get nfsd vers4.0)" no
	assert_equal "$(nfsconf --get nfsd vers4.1)" "" # not set
	assert_equal "$(nfsconf --get nfsd vers4.2)" no
}

@test "set RPC thread count" {
	NUM_NFS_THREADS=42
	DISABLED_NFS_VERSIONS=""

	run configure-nfs
	assert_success

	assert [ -f /etc/nfs.conf.d/knfsd.conf ]
	assert_equal "$(nfsconf --get nfsd threads)" 42
}
