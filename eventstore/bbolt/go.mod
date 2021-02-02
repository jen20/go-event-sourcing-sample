module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.13
	go.etcd.io/bbolt v1.3.4
)

replace github.com/hallgren/eventsourcing => ../..
