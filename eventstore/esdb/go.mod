module github.com/hallgren/eventsourcing/eventstore/esdb

go 1.16

require (
	github.com/EventStore/EventStore-Client-Go/v3 v3.0.0
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/hallgren/eventsourcing/base v0.0.1
	golang.org/x/net v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
)

replace github.com/hallgren/eventsourcing/base => ../../base
