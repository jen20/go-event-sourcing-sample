package bbolt

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hallgren/eventsourcing/core"
	"go.etcd.io/bbolt"
)

type iterator struct {
	tx              *bbolt.Tx
	bucketName      string
	firstEventIndex uint64
	cursor          *bbolt.Cursor
}

// Close closes the iterator
func (i *iterator) Close() {
	i.tx.Rollback()
}

// Next return the next event
func (i *iterator) Next() (core.Event, error) {
	var k, obj []byte
	if i.cursor == nil {
		bucket := i.tx.Bucket([]byte(i.bucketName))
		if bucket == nil {
			return core.Event{}, core.ErrNoMoreEvents
		}
		i.cursor = bucket.Cursor()
		k, obj = i.cursor.Seek(itob(i.firstEventIndex))
		if k == nil {
			return core.Event{}, core.ErrNoMoreEvents
		}
	} else {
		k, obj = i.cursor.Next()
	}
	if k == nil {
		return core.Event{}, core.ErrNoMoreEvents
	}
	bEvent := boltEvent{}
	err := json.Unmarshal(obj, &bEvent)
	if err != nil {
		return core.Event{}, errors.New(fmt.Sprintf("could not deserialize event, %v", err))
	}

	event := core.Event{
		AggregateID:   bEvent.AggregateID,
		AggregateType: bEvent.AggregateType,
		Version:       core.Version(bEvent.Version),
		GlobalVersion: core.Version(bEvent.GlobalVersion),
		Timestamp:     bEvent.Timestamp,
		Metadata:      bEvent.Metadata,
		Data:          bEvent.Data,
		Reason:        bEvent.Reason,
	}
	return event, nil
}
