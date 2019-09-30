package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/snapshotstore/memory"
	"testing"
)

func TestSnapshot(t *testing.T) {
	snapshot := eventsourcing.NewSnapshot(memory.New(json.New()))
	person, err := CreatePerson("morgan")
	if err != nil {
		t.Fatal(err)
	}
	person.GrowOlder()

	snapshot.Save(person.ID, person)
	p := Person{}
	err = snapshot.Get(person.ID, &p)
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.Name != person.Name {
		t.Fatalf("wrong Name in snapshot %q expected: %q", p.Name, person.Name)
	}
}
