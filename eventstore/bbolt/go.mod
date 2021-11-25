module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.17
	go.etcd.io/bbolt v1.3.6
	golang.org/x/sys v0.0.0-20211124211545-fe61309f8881 // indirect
)

//replace github.com/hallgren/eventsourcing => ../..
