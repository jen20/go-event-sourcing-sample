package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"github.com/hallgren/eventsourcing/serializer/unsafe"
	"time"
)

func main() {

	// Setup a memory based event store
	eventStore := memory.Create(unsafe.New())
	repo := eventsourcing.NewRepository(eventStore, nil)
	stream := repo.EventStream()

	// Read the event stream async
	go func() {
		for {
			<-stream.Changes()
			// advance to next value
			stream.Next()
			event := stream.Value().(eventsourcing.Event)
			fmt.Println("STREAM EVENT")
			spew.Dump(event)
		}
	}()

	// Creates the aggregate and adds a second event
	aggregate := CreateFrequentFlierAccount("morgan")
	aggregate.RecordFlightTaken(10, 5)

	// saves the events to the memory backed eventstore
	err := repo.Save(aggregate)
	if err != nil {
		panic("Could not save the aggregate")
	}

	// Load the saved aggregate
	copy := FrequentFlierAccountAggregate{}
	err = repo.Get(string(aggregate.AggregateID), &copy)
	if err != nil {
		panic("Could not get aggregate")
	}

	// Sleep to make sure the events are delivered from the stream
	time.Sleep(time.Millisecond * 100)
	fmt.Println("AGGREGATE")
	spew.Dump(copy)

}
