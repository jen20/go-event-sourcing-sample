package eventsourcing_test

import (
	"context"
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	snap "github.com/hallgren/eventsourcing/snapshotstore/memory"
)

func TestSaveAndGetSnapshot(t *testing.T) {
	eventrepo := eventsourcing.NewRepository(memory.Create())
	eventrepo.Register(&Person{})

	snapshotrepo := eventsourcing.NewSnapshotRepository(snap.Create(), eventrepo)

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = snapshotrepo.Save(person)
	if err != nil {
		t.Fatalf("could not save aggregate, err: %v", err)
	}

	twin := Person{}
	err = snapshotrepo.GetWithContext(context.Background(), person.ID(), &twin)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	// Check internal aggregate version
	if person.Version() != twin.Version() {
		t.Fatalf("Wrong version org %q copy %q", person.Version(), twin.Version())
	}
}

func TestSaveSnapshotWithUnsavedEvents(t *testing.T) {
	eventrepo := eventsourcing.NewRepository(memory.Create())
	eventrepo.Register(&Person{})

	snapshotrepo := eventsourcing.NewSnapshotRepository(snap.Create(), eventrepo)

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = snapshotrepo.SaveSnapshot(person)
	if err == nil {
		t.Fatalf("could save snapshot with unsaved events")
	}
}
