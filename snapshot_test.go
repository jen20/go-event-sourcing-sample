package eventsourcing_test

import (
	"encoding/json"
	"fmt"
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
	b, err := json.Marshal(*person)
	fmt.Println(b)
	snapshot.Save(eventsourcing.AggregateRootID(person.ID), person)
	p := Person{}
	fmt.Printf("person memory 1 %p\n", &p)
	err = snapshot.Get(eventsourcing.AggregateRootID(person.ID), &p)
	fmt.Printf("person memory 4 %p\n", &p)
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.Name != person.Name {
		t.Fatalf("wrong Name in snapshot %q expected: %q", p.Name, person.Name)
	}
}
