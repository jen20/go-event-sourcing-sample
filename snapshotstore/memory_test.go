package snapshotstore_test

import (
	"errors"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/serializer/json"
	"github.com/hallgren/eventsourcing/snapshotstore"
	"testing"
)

// Person aggregate
type Person struct {
	eventsourcing.AggregateRoot
	Name string
	Age  int
	Dead int
}

// Born event
type Born struct {
	Name string
}

// AgedOneYear event
type AgedOneYear struct {
}

// CreatePerson constructor for the Person
func CreatePerson(name string) (*Person, error) {
	if name == "" {
		return nil, errors.New("name can't be blank")
	}
	person := Person{}
	err := person.TrackChange(&person, &Born{Name: name})
	if err != nil {
		return nil, err
	}
	return &person, nil
}

// GrowOlder command
func (person *Person) GrowOlder() error {
	return person.TrackChange(person, &AgedOneYear{})
}

// Transition the person state dependent on the events
func (person *Person) Transition(event eventsourcing.Event) {
	switch e := event.Data.(type) {
	case *Born:
		person.Age = 0
		person.Name = e.Name
	case *AgedOneYear:
		person.Age += 1
	}
}

func TestSnapshot(t *testing.T) {
	snapshot := snapshotstore.New(json.New())
	person, err := CreatePerson("morgan")
	if err != nil {
		t.Fatal(err)
	}
	person.GrowOlder()

	snapshot.Save(person.AggregateID, person)
	// save the version we expect in the snapshot
	personVersion := person.AggregateVersion

	// generate events that are not stored in the snapshot
	person.GrowOlder()
	person.GrowOlder()
	p := Person{}
	err = snapshot.Get(person.AggregateID, &p)
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if p.Name != person.Name {
		t.Fatalf("wrong Name in snapshot %q expected: %q", p.Name, person.Name)
	}

	if p.AggregateVersion != personVersion {
		t.Fatalf("wrong AggregateVersion in snapshot %q expected: %q", p.AggregateVersion, personVersion)
	}
}

func TestGetNoneExistingSnapshot(t *testing.T) {
	snapshot := snapshotstore.New(json.New())

	p := Person{}
	err := snapshot.Get("noneExistingID", &p)
	if err == nil {
		t.Fatalf("could get none existing snapshot %v", err)
	}
}

func TestSaveEmptySnapshotID(t *testing.T) {
	snapshot := snapshotstore.New(json.New())

	p := Person{}
	err := snapshot.Save("", &p)
	if err == nil {
		t.Fatalf("could save blank snapshot id %v", err)
	}
}
