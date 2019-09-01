package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"testing"
)

func TestSaveAndGetAggregate(t *testing.T) {
	repo := eventsourcing.NewRepository(memory.Create())

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save device")
	}
	twin := Person{}
	err = repo.Get(person.ID(), &twin)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	// Check internal aggregate version
	if person.Version() != twin.Version() {
		t.Fatalf("Wrong version org %q copy %q", person.Version(), twin.Version())
	}

	// Check person name
	if person.name != twin.name {
		t.Fatalf("Wrong name org %q copy %q", person.name, twin.name)
	}
}
