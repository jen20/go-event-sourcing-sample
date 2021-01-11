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

func TestSQLSnapshotStore(t *testing.T) {
	f := func() (eventsourcing.SnapshotStore, func(), error) {
		seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		db, err := sqldriver.Open("ramsql", fmt.Sprint(seededRand.Intn(1000000)))
		if err != nil {
			return nil, nil, err
		}
		err = db.Ping()
		if err != nil {
			return nil, nil, err
		}
		store := sql.New(db, *eventsourcing.NewSerializer(json.Marshal, json.Unmarshal))
		err = store.MigrateTest()
		if err != nil {
			return nil, nil, err
		}
		return store, func() { store.Close() }, nil
	}
	suite.Test(t, f)
}
