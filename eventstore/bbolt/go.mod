module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.17-0.20210416123647-9984a8d1d11c
	go.etcd.io/bbolt v1.3.5
)

//replace github.com/hallgren/eventsourcing => ../..
