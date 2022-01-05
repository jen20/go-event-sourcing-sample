module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.19-0.20220112075710-12ad94b190ba
	go.etcd.io/bbolt v1.3.6
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)

replace github.com/hallgren/eventsourcing => ../..
