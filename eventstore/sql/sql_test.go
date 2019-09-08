package sql_test

import (
	sqldriver "database/sql"
	"fmt"
	"github.com/hallgren/eventsourcing/eventstore/sql"
	"github.com/hallgren/eventsourcing/serializer/json"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func Init() (*sql.SQL, func(), error) {
	dbFile := "test.sql"
	os.Remove(dbFile)
	db, err := sqldriver.Open("sqlite3", dbFile)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not open sqlit3 database %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, nil, fmt.Errorf("could not ping database %v", err)
	}
	s := sql.Open(*db, json.New())
	err = s.Migrate()
	if err != nil {
		return nil, nil, fmt.Errorf("could not migrate database %v", err)
	}
	return s, func(){
		s.Close()
		os.Remove(dbFile)
	}, nil
}

