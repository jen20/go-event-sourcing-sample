module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.19
	go.etcd.io/bbolt v1.3.6
	golang.org/x/sys v0.0.0-20220310020820-b874c991c1a5 // indirect
)

//replace github.com/hallgren/eventsourcing => ../..
