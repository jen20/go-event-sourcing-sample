module github.com/hallgren/eventsourcing/eventstore/sql

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.15-0.20210312153409-f001132bb326 // indirect
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
)

replace (
	github.com/hallgren/eventsourcing => ../..
)