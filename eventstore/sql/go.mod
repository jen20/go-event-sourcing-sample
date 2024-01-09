module github.com/hallgren/eventsourcing/eventstore/sql

go 1.13

require (
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/hallgren/eventsourcing/core v0.2.0
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.27.10 // indirect
	github.com/proullon/ramsql v0.0.1
	github.com/ziutek/mymysql v1.5.4 // indirect
)

replace github.com/hallgren/eventsourcing/core => ../../core
