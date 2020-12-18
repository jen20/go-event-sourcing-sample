package sql

import (
	"database/sql"
	"errors"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore"
)

type SQL struct {
	db         *sql.DB
	serializer eventsourcing.Serializer
}

func New(db *sql.DB, serializer eventsourcing.Serializer) *SQL {
	return &SQL{
		db:         db,
		serializer: serializer,
	}
}

// Close the connection
func (sql *SQL) Close() {
	sql.db.Close()
}

func (sql *SQL) Get(id string,a eventsourcing.Aggregate) error {
	return errors.New("not implemented")
}

// Save persists the snapshot
func (sql *SQL) Save(a eventsourcing.Aggregate) error {
	root := a.Root()
	err := snapshotstore.Validate(*root)
	if err != nil {
		return err
	}

	_, err = sql.serializer.Marshal(a)
	if err != nil {
		return err
	}
	return errors.New("not implemented")
}
