package main

import (
	"fmt"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"github.com/hallgren/eventsourcing/serializer/unsafe"
	"time"
)

func main() {
	var c = make(chan eventsourcing.Event)
	// Setup a memory based event store
	eventStore := memory.Create(unsafe.New())
	repo := eventsourcing.NewRepository(eventStore, nil)
	f := func(e eventsourcing.Event) {
		fmt.Printf("Event from stream %q\n", e)
		// Its a good practice making this function as fast as possible not blocking the event sourcing call for to long
		// Here we use the go-observer pkg to store the events in a stream to be consumed async
		c <- e
	}
	repo.Subscribe(f)

	// Read the event stream async
	go func() {
		for {
			// advance to next value
			event := <-c
			fmt.Println("STREAM EVENT")
			fmt.Println(event)
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
	fmt.Println(copy)
}
