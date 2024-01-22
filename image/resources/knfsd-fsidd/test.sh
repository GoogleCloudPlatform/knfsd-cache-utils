#!/usr/bin/env bash

# Bind the container port 5432 to a random local port to
# avoid conflicting with any existing PostgreSQL instances.
export POSTGRES_PORT=0
compose_opts=(-f postgres.compose.yaml)

if [[ $CI == "cloudbuild" ]]; then
	# When running on Cloud Build bind to the standard 5432 port
	export POSTGRES_PORT=5432
	compose_opts+=(-f cloudbuild.compose.yaml)
fi

function compose() {
	docker compose "${compose_opts[@]}" "$@"
}

function start-postgres() {
	compose up --wait
}

function stop-postgres() {
	compose down
}

function url() {
	if [[ $CI == "cloudbuild" ]]; then
		# The docker compose command is not available from the container running
		# the go tests on Cloud Build.
		# When running on Cloud Build assume the PostgreSQL is accessable by
		# host name, and is bound to port 5432 as Cloud Build uses a different
		# networking setup.
		printf 'host=postgres port=5432 user=fsidd password=fsid-test database=fsids'
	else
		local port="$(compose port postgres 5432)"
		# docker compose port outputs in the format ip:port, separate out the port
		port="${port##*:}"
		printf 'host=127.0.0.1 port=%d user=fsidd password=fsid-test database=fsids' "$port"
	fi
}

function run-tests() {
	TEST_DATABASE_URL="$(url)" \
	go test -tags=test.sql "$@" ./...
}

function cleanup() {
	if ! stop-postgres; then
		printf 'ERROR: Failed to stop postgres, container might still be running\n'
		exit 1
	fi
}

case "$1" in
	# By default automatically start postgres, run the tests then stop postgres.
	# Shortcut for:
	#   ./test.sh up && ./test.sh run; ./test.sh down
	"")
		go vet ./... || exit 1

		trap cleanup EXIT
		if ! start-postgres; then
			printf 'ERROR: Failed to start postgres\n'
			exit 1
		fi

		run-tests
		;;

	up) start-postgres;;

	down) stop-postgres;;

	run)
		shift
		go vet ./... || exit 1
		run-tests "$@"
		;;

	# print the database URL for use with debugging in vscode
	# in .vscode/settings.json add:
	#   "go.testTags": "test.sql",
	#   "go.testEnvVars": {
	#     "TEST_DATABASE_URL": "<paste url here>"
	#   }
	url) url && printf '\n';;

	*) printf 'Unknown command "%s"\n' "$1";;
esac
