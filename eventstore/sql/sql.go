package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore"
	"github.com/imkira/go-observer"
)

type SQL struct {
	db sql.DB
	serializer eventstore.EventSerializer
	eventsProperty  observer.Property                // A property to which all event changes for all event types are published
}

func Open(db sql.DB, serializer eventstore.EventSerializer) *SQL {
	return &SQL{
		db: db,
		serializer: serializer,
		eventsProperty: observer.NewProperty(nil),
	}
}

func (sql *SQL) Close() {
	sql.db.Close()
}

func (sql *SQL) Save(events []eventsourcing.Event) error {
	// If no event return no error
	if len(events) == 0 {
		return nil
	}
	aggregateID := events[0].AggregateRootID
	aggregateType := events[0].AggregateType

	// the current version of that is the last event saved
	selectStm := `Select data from events where aggregate_id=? and aggregate_type=? order by version desc limit 1`
	rows, err := sql.db.Query(selectStm, aggregateID, aggregateType)
	if err != nil {
		return err
	}
	defer rows.Close()

	currentVersion := eventsourcing.Version(0)
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return err
		}
		event,err := sql.serializer.DeserializeEvent([]byte(data))
		if err != nil {
			return fmt.Errorf("Could not deserialize event %v", err)
		}
		currentVersion = event.Version
	}

	//Validate events
	err = eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
		return err
	}

	tx, err := sql.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("Could not start a write transaction, %v", err)
	}
	defer tx.Rollback()
	insert := `Insert into events (aggregate_id, version, reason, aggregate_type, data, meta_data) values ($1, $2, $3, $4, $5, $6)`
	for _, event := range events {
		d, err := sql.serializer.SerializeEvent(event)
		if err != nil {
			return err
		}
		_, err = tx.Exec(insert, event.AggregateRootID, event.Version, event.Reason, event.AggregateType, string(d), "")
		if err != nil {
			return err
		}
	}
	tx.Commit()

	// publish to stream after transaction is complete
	for _, event := range events {
		sql.eventsProperty.Update(event)
	}
	return nil
}

func (sql *SQL) Get(id string, aggregateType string) (events []eventsourcing.Event, err error) {
	selectStm := `Select data from events where aggregate_id=? and aggregate_type=? order by version asc`
	rows, err := sql.db.Query(selectStm, id, aggregateType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		event,err := sql.serializer.DeserializeEvent([]byte(data))
		if err != nil {
			return nil, fmt.Errorf("Could not deserialize event %v", err)
		}
		events = append(events, event)
	}
	return
}

func (sql *SQL) GlobalGet(start int, count int) []eventsourcing.Event {
	selectStm := `Select data from events where id>=? order by id asc limit ?`
	rows, err := sql.db.Query(selectStm, start, count)
	if err != nil {
		return nil
	}
	defer rows.Close()
	events, err := sql.transform(rows)
	if err != nil {
		return nil
	}
	return events
}

func (sql *SQL) EventStream() observer.Stream {
	return sql.eventsProperty.Observe()
}

func (sql *SQL) transform(rows *sql.Rows) (events []eventsourcing.Event, err error) {
	events = make([]eventsourcing.Event, 0)
	for rows.Next() {
		var data string
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}
		event,err := sql.serializer.DeserializeEvent([]byte(data))
		if err != nil {
			return nil, fmt.Errorf("Could not deserialize event %v", err)
		}
		events = append(events, event)
	}
	return
}