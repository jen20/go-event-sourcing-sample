module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing/core v0.1.0
	go.etcd.io/bbolt v1.3.7
	golang.org/x/sys v0.10.0 // indirect
)

replace github.com/hallgren/eventsourcing/core => ../../core
