module github.com/hallgren/eventsourcing/eventstore/sql

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.15-0.20210313144026-b38f9554394c // indirect
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
)

replace (
	github.com/hallgren/eventsourcing => ../..
)