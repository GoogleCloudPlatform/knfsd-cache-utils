#!/usr/bin/env bash

function start-postgres() {
	docker compose -f postgres.yaml up --wait
}

function stop-postgres() {
	docker compose -f postgres.yaml down
}

function url() {
	# the container port 5432 will be bound to a random local port
	local port="$(docker compose -f postgres.yaml port postgres 5432)"
	# docker compose port outputs in the format ip:port, separate out the port
	port="${port##*:}"
	printf 'host=127.0.0.1 port=%d user=fsidd password=fsid-test database=fsids' "$port"
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
		shift
		go vet ./... || exit 1

		trap cleanup EXIT
		if ! start-postgres; then
			printf 'ERROR: Failed to start postgres\n'
			exit 1
		fi

		run-tests "$@"
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
