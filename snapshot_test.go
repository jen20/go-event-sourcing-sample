package eventsourcing_test

import (
	"encoding/xml"
	memory2 "github.com/hallgren/eventsourcing/eventstore/memory"
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore/memory"
	memsnap "github.com/hallgren/eventsourcing/snapshotstore/memory"
)

func TestSnapshot(t *testing.T) {
	ser := eventsourcing.NewSerializer(xml.Marshal, xml.Unmarshal)
	s := eventsourcing.SnapshotNew(memory.New(), *ser)

	// use repo to reset events on person to be able to save snapshot
	repo := eventsourcing.NewRepository(memory2.Create(), s)

	person, err := CreatePersonWithID("123", "kalle")
	repo.Save(person)

	err = s.Save(person)
	if err != nil {
		t.Fatal(err)
	}

	p := Person{}
	err = s.Get("123", &p)
	if err != nil {
		t.Fatalf("could not get person from snapshot %v", err)
	}
	if person.ID() != p.ID() {
		t.Fatalf("wrong ID in snapshot %q expected: %q", person.ID(), p.ID())
	}
	if p.Age != person.Age {
		t.Fatalf("wrong Age in snapshot %d expected: %d", p.Age, person.Age)
	}
	if p.ID() != person.ID() {
		t.Fatalf("wrong id %s %s", p.ID(), person.ID())
	}
	if p.Version() != person.Version() {
		t.Fatalf("wrong version %d %d", p.Version(), person.Version())
	}
	if p.GlobalVersion() != person.GlobalVersion() {
		t.Fatalf("wrong global version %d %d", p.GlobalVersion(), person.GlobalVersion())
	}

	// store the snapshot once more
	person.Age = 99
	s.Save(person)

	err = s.Get(person.ID(), &p)
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.Age != person.Age {
		t.Fatalf("wrong age %d %d", p.Age, person.Age)
	}
}
func TestGetNoneExistingSnapshot(t *testing.T) {
	ser := eventsourcing.NewSerializer(xml.Marshal, xml.Unmarshal)
	s := eventsourcing.SnapshotNew(memsnap.New(), *ser)
	p := Person{}
	err := s.Get("noneExistingID", &p)
	if err == nil {
		t.Fatalf("could get none existing snapshot %v", err)
	}
}

func TestSaveEmptySnapshotID(t *testing.T) {
	ser := eventsourcing.NewSerializer(xml.Marshal, xml.Unmarshal)
	s := eventsourcing.SnapshotNew(memsnap.New(), *ser)
	p := Person{}
	err := s.Save(&p)
	if err == nil {
		t.Fatalf("could save blank snapshot id %v", err)
	}
}
