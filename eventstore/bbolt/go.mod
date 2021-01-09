module github.com/hallgren/eventsourcing/eventstore/bbolt

go 1.13

require (
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/hallgren/eventsourcing v0.0.12
	go.etcd.io/bbolt v1.3.4
)

replace (
	"github.com/hallgren/eventsourcing" => ../..
)