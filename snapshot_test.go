package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore/memory"
	"testing"
)

func TestSnapshot(t *testing.T) {
	snapshot := eventsourcing.NewSnapshot(memory.New())
	person, err := CreatePerson("morgan")
	if err != nil {
		t.Fatal(err)
	}
	person.GrowOlder()
	snapshot.Save(eventsourcing.AggregateRootID(person.ID()), person)
	p := Person{}
	err = snapshot.Get(eventsourcing.AggregateRootID(person.ID()), interface{}(&p))
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.name != person.name {
		t.Fatalf("wrong name in snapshot %q expected: %q", p.name, person.name)
	}
}
