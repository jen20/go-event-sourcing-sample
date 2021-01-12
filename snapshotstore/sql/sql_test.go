package sql_test

import (
	sqldriver "database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore/sql"
	"github.com/hallgren/eventsourcing/snapshotstore/suite"
	_ "github.com/proullon/ramsql/driver"
)

type provider struct {
	db *sqldriver.DB
}

func (p *provider) Setup() (eventsourcing.SnapshotStore, error) {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	db, err := sqldriver.Open("ramsql", fmt.Sprint(seededRand.Intn(1000000)))
	if err != nil {
		return nil, err
	}
	p.db = db

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	store := sql.New(db, *eventsourcing.NewSerializer(json.Marshal, json.Unmarshal))
	err = store.MigrateTest()
	return store, err
}

func (p *provider) Cleanup() { p.db.Exec(`delete from snapshots;`) }

func (p *provider) Teardown() { p.db.Close() }

func TestSQLSnapshotStore(t *testing.T) {
	suite.Test(t, new(provider))
}
