package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"time"
)

func main() {

	// Setup a memory based event store
	eventStore := memory.Create()
	repo := eventsourcing.NewRepository(eventStore)
	stream := repo.EventStream()

	// Read the event stream async
	go func(){
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


	err := repo.Save(aggregate)
	if err != nil {
		panic("Could not save the aggregate")
	}

	// Load the saved aggregate
	copy := FrequentFlierAccountAggregate{}
	eventsourcing.CreateAggregate(&copy)
	err = repo.Get(aggregate.ID(), &copy)
	if err != nil {
		panic("Could not get aggregate")
	}

	// Sleep to make sure the events are delivered from the stream
	time.Sleep(time.Millisecond*100)
	fmt.Println("AGGREGATE")
	spew.Dump(copy)


}
