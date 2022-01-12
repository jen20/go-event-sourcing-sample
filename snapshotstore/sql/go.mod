module github.com/hallgren/eventsourcing/snapshotstore/sql

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.19-0.20220112075710-12ad94b190ba
	github.com/proullon/ramsql v0.0.0-20211120092837-c8d0a408b939
)

replace github.com/hallgren/eventsourcing => ../..
