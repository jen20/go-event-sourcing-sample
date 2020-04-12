package bbolt_test

import (
	"github.com/hallgren/eventsourcing/eventstore/bbolt"
	"github.com/hallgren/eventsourcing/eventstore/suite"
	"github.com/hallgren/eventsourcing/serializer/json"
	"os"
	"testing"
)

func TestSuite(t *testing.T) {
	f := func() (suite.Eventstore, func(), error) {
		dbFile := "bolt.db"
		es := bbolt.MustOpenBBolt(dbFile, json.New())
		return es, func(){
			es.Close()
			os.Remove(dbFile)
		}, nil
	}
	suite.Test(t, f)
}
