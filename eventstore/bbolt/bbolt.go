package bbolt

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore"
	"go.etcd.io/bbolt"
)

const (
	globalEventOrderBucketName = "global_event_order"
)

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// BBolt is a handler for event streaming
type BBolt struct {
	db         *bbolt.DB                // The bbolt db where we store everything
	serializer eventsourcing.Serializer // The serializer
}

type boltEvent struct {
	AggregateID   string
	Version       uint64
	GlobalVersion uint64
	Reason        string
	AggregateType string
	Timestamp     time.Time
	Data          []byte
	MetaData      map[string]interface{}
}

// MustOpenBBolt opens the event stream found in the given file. If the file is not found it will be created and
// initialized. Will panic if it has problems persisting the changes to the filesystem.
func MustOpenBBolt(dbFile string, s eventsourcing.Serializer) *BBolt {
	db, err := bbolt.Open(dbFile, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	// Ensure that we have a bucket to store the global event ordering
	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(globalEventOrderBucketName)); err != nil {
			return errors.New("could not create global event order bucket")
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return &BBolt{
		db:         db,
		serializer: s,
	}
}

// Save an aggregate (its events)
func (e *BBolt) Save(events []eventsourcing.Event) (eventsourcing.Version, error) {
	// Return if there is no events to save
	if len(events) == 0 {
		return 0, nil
	}

	// get bucket name from first event
	aggregateType := events[0].AggregateType
	aggregateID := events[0].AggregateID
	bucketName := aggregateKey(aggregateType, aggregateID)

	tx, err := e.db.Begin(true)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	evBucket := tx.Bucket([]byte(bucketName))
	if evBucket == nil {
		// Ensure that we have a bucket named events_aggregateType_aggregateID for the given aggregate
		err = e.createBucket([]byte(bucketName), tx)
		if err != nil {
			return 0, errors.New("could not create aggregate events bucket")
		}
		evBucket = tx.Bucket([]byte(bucketName))
	}

	currentVersion := eventsourcing.Version(0)
	cursor := evBucket.Cursor()
	k, obj := cursor.Last()
	if k != nil {
		lastEvent := eventsourcing.Event{}
		err := e.serializer.Unmarshal(obj, &lastEvent)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("could not serialize event, %v", err))
		}
		currentVersion = lastEvent.Version
	}

	//Validate events
	err = eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
		return 0, err
	}

	globalBucket := tx.Bucket([]byte(globalEventOrderBucketName))
	if globalBucket == nil {
		return 0, errors.New("global bucket not found")
	}

	var globalSequence uint64
	for _, event := range events {
		sequence, err := evBucket.NextSequence()
		if err != nil {
			return 0, errors.New(fmt.Sprintf("could not get sequence for %#v", bucketName))
		}

		// We need to establish a global event order that spans over all buckets. This is so that we can be
		// able to play the event (or send) them in the order that they was entered into this database.
		// The global sequence bucket contains an ordered line of pointer to all events on the form bucket_name:seq_num
		globalSequence, err = globalBucket.NextSequence()
		if err != nil {
			return 0, errors.New("could not get next sequence for global bucket")
		}

		// marshal the event.Data separately to be able to handle the type info
		eventData, err := e.serializer.Marshal(event.Data)

		// build the internal bolt event
		bEvent := boltEvent{
			AggregateID:   event.AggregateID,
			AggregateType: event.AggregateType,
			Version:       uint64(event.Version),
			GlobalVersion: globalSequence,
			Reason:        event.Reason,
			Timestamp:     event.Timestamp,
			MetaData:      event.MetaData,
			Data:          eventData,
		}

		value, err := e.serializer.Marshal(bEvent)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("could not serialize event, %v", err))
		}

		err = evBucket.Put(itob(sequence), value)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("could not save event %#v in bucket", event))
		}
		err = globalBucket.Put(itob(globalSequence), value)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("could not save global sequence pointer for %#v", bucketName))
		}
	}
	return eventsourcing.Version(globalSequence), tx.Commit()
}

// Get aggregate events
func (e *BBolt) Get(id string, aggregateType string, afterVersion eventsourcing.Version) ([]eventsourcing.Event, error) {
	bucketName := aggregateKey(aggregateType, id)

	tx, err := e.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	evBucket := tx.Bucket([]byte(bucketName))
	if evBucket == nil {
		return nil, eventsourcing.ErrNoEvents
	}

	cursor := evBucket.Cursor()
	events := make([]eventsourcing.Event, 0)
	firstEvent := afterVersion + 1

	for k, obj := cursor.Seek(itob(uint64(firstEvent))); k != nil; k, obj = cursor.Next() {
		bEvent := boltEvent{}
		err := e.serializer.Unmarshal(obj, &bEvent)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("could not deserialize event, %v", err))
		}
		f, ok := e.serializer.Type(bEvent.AggregateType, bEvent.Reason)
		if !ok {
			// if the typ/reason is not register jump over the event
			continue
		}
		eventData := f()
		err = e.serializer.Unmarshal(bEvent.Data, &eventData)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("could not deserialize event data, %v", err))
		}
		event := eventsourcing.Event{
			AggregateID:   bEvent.AggregateID,
			AggregateType: bEvent.AggregateType,
			Version:       eventsourcing.Version(bEvent.Version),
			GlobalVersion: eventsourcing.Version(bEvent.GlobalVersion),
			Reason:        bEvent.Reason,
			Timestamp:     bEvent.Timestamp,
			MetaData:      bEvent.MetaData,
			Data:          eventData,
		}
		events = append(events, event)
	}
	if len(events) == 0 {
		return nil, eventsourcing.ErrNoEvents
	}
	return events, nil
}

// Close closes the event stream and the underlying database
func (e *BBolt) Close() error {
	return e.db.Close()
}

// CreateBucket creates a bucket
func (e *BBolt) createBucket(bucketName []byte, tx *bbolt.Tx) error {
	// Ensure that we have a bucket named event_type for the given type
	if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
		return errors.New(fmt.Sprintf("could not create bucket for %s: %s", bucketName, err))
	}
	return nil

}

// aggregateKey generate a aggregate key to store events against from aggregateType and aggregateID
func aggregateKey(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}
