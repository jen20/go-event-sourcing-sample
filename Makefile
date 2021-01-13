all:
	go fmt ./...
	go build

test:
	# event stores
	cd eventstore/bbolt && go test -count 1 ./...
	cd eventstore/sql && go test -count 1 ./...

	# snapshot stores
	cd snapshotstore/sql && go test -count 1 ./...
	cd snapshotstore/memory && go test -count 1 ./...
	
	# main
	go test -count 1 ./...
