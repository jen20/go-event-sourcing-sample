module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/hallgren/eventsourcing v0.0.15-0.20210310221131-4bb8960a066e
	go.etcd.io/bbolt v1.3.4
)

replace github.com/hallgren/eventsourcing => ../..
