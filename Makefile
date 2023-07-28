all:
	go fmt ./...
	go build

	#core
	cd core && go build
	# event stores
	cd eventstore/bbolt && go build
	cd eventstore/sql && go build
	cd eventstore/esdb && go build
test:
	#core
	cd core && go test -count 1 ./...
	# event stores
	cd eventstore/bbolt && go test -count 1 ./...
	cd eventstore/sql && go test -count 1 ./...
	cd eventstore/esdb && go test esdb_test.go -count 1 ./...

	# main
	go test -count 1 ./...
