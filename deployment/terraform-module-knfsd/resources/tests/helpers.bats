function setup() {
	source /opt/bats/lib/bats-support/load.bash
	source /opt/bats/lib/bats-assert/load.bash
	load ../proxy-startup.sh
}

@test "split" {
	run split <<< "foo,bar,baz"
	assert_success
	assert_equal ${#lines[@]} 3
	assert_line -n 0 foo
	assert_line -n 1 bar
	assert_line -n 2 baz
}

@test "split single item" {
	run split <<< "foo"
	assert_success
	assert_output "foo"
}

@test "split trims whitespace" {
	run split <<< "foo   ,   bar   ,   baz"
	assert_success
	assert_equal ${#lines[@]} 3
	assert_line -n 0 foo
	assert_line -n 1 bar
	assert_line -n 2 baz
}

@test "split ignores empty items" {
	run split <<< "foo,,bar"
	assert_success
	assert_equal ${#lines[@]} 2
	assert_line -n 0 foo
	assert_line -n 1 bar
}

@test "split empty" {
	run split <<< ""
	assert_success
	assert_equal ${#lines[@]} 0
}

@test "trim_slash" {
	run trim_slash <<-EOT
		/a/
		/b
		/c/d/
		/e/f
	EOT

	assert_success
	assert_equal ${#lines[@]} 4
	assert_line -n 0 /a
	assert_line -n 1 /b
	assert_line -n 2 /c/d
	assert_line -n 3 /e/f
}

@test "trim slash preserves root" {
	run trim_slash <<< "/"
	assert_success
	assert_output "/"
}

@test "trim slash removes multiple trailing slashes" {
	run trim_slash <<< "/foo/bar///"
	assert_success
	assert_output "/foo/bar"
}
