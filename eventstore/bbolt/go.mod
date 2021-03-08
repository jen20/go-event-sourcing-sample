module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.15-0.20210308210900-d0214026cfcd
	go.etcd.io/bbolt v1.3.4
)

replace github.com/hallgren/eventsourcing => ../../.
