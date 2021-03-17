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

	statement := `SELECT state, version, global_version from snapshots where id=$1 AND type=$2 LIMIT 1`
	var state []byte
	var version uint64
	var globalVersion uint64
	err = tx.QueryRow(statement, id, typ).Scan(&state, &version, &globalVersion)
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		return eventsourcing.ErrSnapshotNotFound
	}
	err = s.serializer.Unmarshal(state, a)
	if err != nil {
		return err
	}
	r := a.Root()
	snapshot := eventsourcing.Snapshot{
		Version:       eventsourcing.Version(version),
		GlobalVersion: eventsourcing.Version(globalVersion),
	}
	r.BuildFromSnapshot(a, snapshot)
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
	if err == sql.ErrNoRows {
		// insert
		statement = `INSERT INTO snapshots (state, id, type, version, global_version) VALUES ($1, $2, $3, $4, $5)`
		_, err = tx.Exec(statement, string(data), root.ID(), typ, root.Version(), root.GlobalVersion())
		if err != nil {
			return err
		}
	} else {
		// update
		statement = `UPDATE snapshots set state=$1, version=$2, global_version=$3 where id=$4 AND type=$5`
		_, err = tx.Exec(statement, string(data), root.Version(), root.GlobalVersion(), root.ID(), typ)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
