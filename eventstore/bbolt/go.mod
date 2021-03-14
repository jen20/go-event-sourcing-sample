module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.15-0.20210313144026-b38f9554394c
	go.etcd.io/bbolt v1.3.4
)

replace (
	github.com/hallgren/eventsourcing => ../..
)
