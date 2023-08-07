module github.com/hallgren/eventsourcing/eventstore/esdb

go 1.16

require (
	github.com/EventStore/EventStore-Client-Go/v3 v3.1.0
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/hallgren/eventsourcing/core v0.1.0
	golang.org/x/net v0.12.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230726155614-23370e0ffb3e // indirect
	google.golang.org/grpc v1.57.0 // indirect
)

replace github.com/hallgren/eventsourcing/core => ../../core
