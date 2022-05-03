module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.20
	go.etcd.io/bbolt v1.3.6
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
)

//replace github.com/hallgren/eventsourcing => ../..
