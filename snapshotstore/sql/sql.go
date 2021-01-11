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
func (sql *SQL) Close() {
	sql.db.Close()
}

// Get retrieves the persisted snapshot
func (sql *SQL) Get(id string, a eventsourcing.Aggregate) error {
	typ := reflect.TypeOf(a).Elem().Name()

	tx, err := sql.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statement := `SELECT data from snapshots where id=$1 AND type=$2 LIMIT 1`
	rows, err := tx.Query(statement, id, typ)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			return err
		}
		err = sql.serializer.Unmarshal(data, a)
		if err != nil {
			return err
		}
	} else {
		return eventsourcing.ErrSnapshotNotFound
	}
	return nil
}

// Save persists the snapshot
func (sql *SQL) Save(a eventsourcing.Aggregate) error {
	root := a.Root()
	err := snapshotstore.Validate(*root)
	if err != nil {
		return err
	}

	typ := reflect.TypeOf(a).Elem().Name()
	data, err := sql.serializer.Marshal(a)
	if err != nil {
		return err
	}

	tx, err := sql.db.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("could not start a write transaction, %v", err))
	}
	defer tx.Rollback()

	statement := `SELECT data from snapshots where id=$1 AND type=$2 LIMIT 1`
	rows, err := tx.Query(statement, root.ID(), typ)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		// update
		statement = `UPDATE snapshots set data=$1 where id=$2 AND type=$3`
		_, err = tx.Exec(statement, string(data), root.ID(), typ)
		if err != nil {
			return err
		}
	} else {
		// insert
		statement = `INSERT INTO snapshots (data, id, type) VALUES ($1, $2, $3)`
		_, err = tx.Exec(statement, string(data), root.ID(), typ)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
