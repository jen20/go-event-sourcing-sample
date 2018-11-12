package bbolt_test

import (
	"fmt"
	"go-event-sourcing-sample/pkg/eventsourcing"
	"go-event-sourcing-sample/pkg/eventstore/bbolt"
	"os"
	"testing"
)

// Person aggregate
type Person struct {
	aggregateRoot eventsourcing.AggregateRoot
	name          string
	age           int
}

// PersonCreated event
type PersonCreated struct {
	name       string
	initialAge int
}

// AgedOneYear event
type AgedOneYear struct {
}

// transition the person state dependent on the events
func (person *Person) transition(event eventsourcing.Event) {
	switch e := event.Data.(type) {

	case PersonCreated:
		person.age = e.initialAge
		person.name = e.name

	case AgedOneYear:
		person.age += 1
	}

}

// CreatePersonWithID constructor for the Person that sets the aggregate id from the outside
func CreatePersonWithID(id, name string) (*Person, error) {

	if name == "" {
		return nil, fmt.Errorf("name can't be blank")
	}

	person := &Person{}
	person.aggregateRoot.SetID(id)
	person.aggregateRoot.TrackChange(*person, PersonCreated{name: name, initialAge: 0}, person.transition)
	return person, nil
}

// NewFrequentFlierAccountFromHistory creates a FrequentFlierAccount given a history
// of the changes which have occurred for that account.
func CreatePersonFromHistory(events []eventsourcing.Event) *Person {
	state := &Person{}
	state.aggregateRoot.BuildFromHistory(events, state.transition)
	return state
}

// GrowOlder command
func (person *Person) GrowOlder() {
	person.aggregateRoot.TrackChange(*person, AgedOneYear{}, person.transition)
}

// Benchmark the time it takes to retrieve a list of keys for an entity type
func BenchmarkFetchAndApply101Events(b *testing.B) {
	os.Remove(dbFile)
	eventstore := bbolt.MustOpenBBolt(dbFile)
	defer eventstore.Close()

	person, err := CreatePersonWithID("123", "kalle")
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < 100; i++ {
		person.GrowOlder()
	}

	err = eventstore.Save(person.aggregateRoot.Changes())
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		events, _ := eventstore.Get("123", person.aggregateRoot.Changes()[0].AggregateType)
		p := CreatePersonFromHistory(events)
		if p.age != 100 {
			b.Error("person holds the wrong age")
		}
	}
}
