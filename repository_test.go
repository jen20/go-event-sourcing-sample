package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/snapshotstore"
	"testing"
	"time"
)

func TestSaveAndGetAggregate(t *testing.T) {
	serializer := json.New()
	serializer.Register(&Person{}, &Born{}, &AgedOneYear{})
	repo := eventsourcing.NewRepository(memory.Create(serializer), nil)

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save aggregate")
	}
	twin := Person{}
	err = repo.Get(string(person.AggregateID), &twin)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	// Check internal aggregate version
	if person.AggregateVersion != twin.AggregateVersion {
		t.Fatalf("Wrong version org %q copy %q", person.AggregateVersion, twin.AggregateVersion)
	}

	// Check person Name
	if person.Name != twin.Name {
		t.Fatalf("Wrong Name org %q copy %q", person.Name, twin.Name)
	}
}

func TestSaveAndGetAggregateSnapshotAndEvents(t *testing.T) {
	serializer := json.New()
	serializer.Register(&Person{}, &Born{}, &AgedOneYear{})
	snapshot := snapshotstore.New(serializer)
	repo := eventsourcing.NewRepository(memory.Create(serializer), snapshot)

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save aggregate")
	}

	// save person to snapshot store
	err = repo.SaveSnapshot(person)
	if err != nil {
		t.Fatal(err)
	}
	person.GrowOlder()
	repo.Save(person)
	twin := Person{}
	err = repo.Get(string(person.AggregateID), &twin)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	// Check internal aggregate version
	if person.AggregateVersion != twin.AggregateVersion {
		t.Fatalf("Wrong version org %q copy %q", person.AggregateVersion, twin.AggregateVersion)
	}

	// Check person Name
	if person.Name != twin.Name {
		t.Fatalf("Wrong Name org %q copy %q", person.Name, twin.Name)
	}
}

func TestSaveSnapshotWithUnsavedEvents(t *testing.T) {
	serializer := json.New()
	serializer.Register(&Person{}, &Born{}, &AgedOneYear{})
	snapshot := snapshotstore.New(serializer)
	repo := eventsourcing.NewRepository(memory.Create(serializer), snapshot)

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	// save person to snapshot store
	err = repo.SaveSnapshot(person)
	if err == nil {
		t.Fatalf("should not save snapshot with unsaved events %v", err)
	}
}

func TestEventStream(t *testing.T) {
	serializer := json.New()
	serializer.Register(&Person{}, &Born{}, &AgedOneYear{})
	snapshotstore := snapshotstore.New(serializer)
	eventstore := memory.Create(serializer)
	repo := eventsourcing.NewRepository(eventstore, snapshotstore)
	stream := repo.EventStream()

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	person.GrowOlder()
	person.GrowOlder()
	person.GrowOlder()

	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save aggregate")
	}

	counter := 0
outer:
	for {
		select {
		// wait for changes
		case <-stream.Changes():
			// advance to next value
			stream.Next()
			counter++
		case <-time.After(10 * time.Millisecond):
			// The stream has 10 milliseconds to deliver the events
			break outer
		}
	}

	if counter != 4 {
		t.Errorf("Not all events was received from the stream, got %q", counter)
	}
}
