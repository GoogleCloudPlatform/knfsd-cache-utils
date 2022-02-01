function setup() {
	load ../proxy-startup.sh
}

@test "is_excluded_export" {
	# The EXCLUDED_EXPORTS array will never contain trailing slashes as the
	# input is run through the trim_slash function
	EXCLUDED_EXPORTS=(
		/bin
		/home
		/usr
	)

	is_excluded_export /bin
	is_excluded_export /bin/
	is_excluded_export /home
	is_excluded_export /usr
	! is_excluded_export /usr/local
	! is_excluded_export /
}

@test "is_excluded_export can exclude root" {
	EXCLUDED_EXPORTS=(
		/
	)

	is_excluded_export /
	! is_excluded_export /other
}
