package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/snapshotstore"
	"testing"
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
		t.Fatal("could not save device")
	}
	twin := Person{}
	err = repo.Get(string(person.ID), &twin)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	// Check internal aggregate version
	if person.Version != twin.Version {
		t.Fatalf("Wrong version org %q copy %q", person.Version, twin.Version)
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
		t.Fatal("could not save device")
	}

	// save person to snapshot store
	err = snapshot.Save(person.ID, person)
	if err != nil {
		t.Fatal(err)
	}
	person.GrowOlder()
	repo.Save(person)
	twin := Person{}
	err = repo.Get(string(person.ID), &twin)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	// Check internal aggregate version
	if person.Version != twin.Version {
		t.Fatalf("Wrong version org %q copy %q", person.Version, twin.Version)
	}

	// Check person Name
	if person.Name != twin.Name {
		t.Fatalf("Wrong Name org %q copy %q", person.Name, twin.Name)
	}
}