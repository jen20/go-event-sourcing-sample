package sql_test

import (
	sqldriver "database/sql"
	"errors"
	"fmt"
	s "github.com/hallgren/eventsourcing/eventstore/sql"
	"github.com/hallgren/eventsourcing/eventstore/suite"
	"github.com/hallgren/eventsourcing/serializer/json"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

func TestSuite(t *testing.T) {
	f := func() (suite.Eventstore, func(), error) {
		dbFile := "test.sql"
		os.Remove(dbFile)
		db, err := sqldriver.Open("sqlite3", dbFile)
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("could not open sqlit3 database %v", err))
		}
		err = db.Ping()
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("could not ping database %v", err))
		}
		es := s.Open(*db, json.New())
		err = es.Migrate()
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("could not migrate database %v", err))
		}
		return es, func() {
			es.Close()
			os.Remove(dbFile)
		}, nil
	}
	suite.Test(t, f)
}
