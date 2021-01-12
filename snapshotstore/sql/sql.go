package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore"
)

// SQL is the struct holding the underlying database and serializer
type SQL struct {
	db         *sql.DB
	serializer eventsourcing.Serializer
}

// New returns a SQL struct
func New(db *sql.DB, serializer eventsourcing.Serializer) *SQL {
	return &SQL{
		db:         db,
		serializer: serializer,
	}
}

// Close the connection
func (s *SQL) Close() {
	s.db.Close()
}

// Get retrieves the persisted snapshot
func (s *SQL) Get(id string, a eventsourcing.Aggregate) error {
	typ := reflect.TypeOf(a).Elem().Name()

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statement := `SELECT data from snapshots where id=$1 AND type=$2 LIMIT 1`
	var data []byte
	err = tx.QueryRow(statement, id, typ).Scan(&data)
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		return eventsourcing.ErrSnapshotNotFound
	}
	err = s.serializer.Unmarshal(data, a)
	if err != nil {
		return err
	}
	return nil
}

// Save persists the snapshot
func (s *SQL) Save(a eventsourcing.Aggregate) error {
	root := a.Root()
	err := snapshotstore.Validate(*root)
	if err != nil {
		return err
	}

	typ := reflect.TypeOf(a).Elem().Name()
	data, err := s.serializer.Marshal(a)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("could not start a write transaction, %v", err))
	}
	defer tx.Rollback()

	statement := `SELECT id from snapshots where id=$1 AND type=$2 LIMIT 1`
	var id string
	err = tx.QueryRow(statement, root.ID(), typ).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows  {
		// insert
		statement = `INSERT INTO snapshots (data, id, type) VALUES ($1, $2, $3)`
		_, err = tx.Exec(statement, string(data), root.ID(), typ)
		if err != nil {
			return err
		}
	} else {
		// update
		statement = `UPDATE snapshots set data=$1 where id=$2 AND type=$3`
		_, err = tx.Exec(statement, string(data), root.ID(), typ)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
