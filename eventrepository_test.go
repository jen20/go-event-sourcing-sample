package eventsourcing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
)

func TestSaveAndGetAggregate(t *testing.T) {
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if err != nil {
		t.Fatalf("could not save aggregate, err: %v", err)
	}

	// make sure the global version is set to 1
	if person.GlobalVersion() != 1 {
		t.Fatalf("global version is: %d expected: 1", person.GlobalVersion())
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

	// Check person Name
	if person.Name != twin.Name {
		t.Fatalf("Wrong Name org %q copy %q", person.Name, twin.Name)
	}
}

func TestGetWithContext(t *testing.T) {
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})
	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save aggregate")
	}

	twin := Person{}
	err = repo.GetWithContext(context.Background(), person.ID(), &twin)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	// Check internal aggregate version
	if person.Version() != twin.Version() {
		t.Fatalf("Wrong version org %q copy %q", person.Version(), twin.Version())
	}

	// Check person Name
	if person.Name != twin.Name {
		t.Fatalf("Wrong Name org %q copy %q", person.Name, twin.Name)
	}
}

func TestGetWithContextCancel(t *testing.T) {
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save aggregate")
	}

	twin := Person{}
	ctx, cancel := context.WithCancel(context.Background())

	// cancel the context
	cancel()
	err = repo.GetWithContext(ctx, person.ID(), &twin)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected error context.Canceled but was %v", err)
	}
}

func TestGetNoneExistingAggregate(t *testing.T) {
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	p := Person{}
	err := repo.Get("none_existing", &p)
	if err != eventsourcing.ErrAggregateNotFound {
		t.Fatal("could not get aggregate")
	}
}

func TestSubscriptionAllEvent(t *testing.T) {
	counter := 0
	f := func(e eventsourcing.Event) {
		counter++
	}
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	s := repo.Subscribers().All(f)
	defer s.Close()

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
	if counter != 4 {
		t.Errorf("No global events was received from the stream, got %q", counter)
	}
}

func TestSubscriptionSpecificEvent(t *testing.T) {
	counter := 0
	f := func(e eventsourcing.Event) {
		counter++
	}
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	s := repo.Subscribers().Event(f, &Born{}, &AgedOneYear{})
	defer s.Close()

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
	if counter != 4 {
		t.Errorf("No global events was received from the stream, got %q", counter)
	}
}

func TestSubscriptionAggregate(t *testing.T) {
	counter := 0
	f := func(e eventsourcing.Event) {
		counter++
	}
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	s := repo.Subscribers().Aggregate(f, &Person{})
	defer s.Close()

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
	if counter != 4 {
		t.Errorf("No global events was received from the stream, got %q", counter)
	}
}

func TestSubscriptionSpecificAggregate(t *testing.T) {
	counter := 0
	f := func(e eventsourcing.Event) {
		counter++
	}
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	s := repo.Subscribers().AggregateID(f, person)
	defer s.Close()

	person.GrowOlder()
	person.GrowOlder()
	person.GrowOlder()

	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save aggregate")
	}
	if counter != 4 {
		t.Errorf("No global events was received from the stream, got %q", counter)
	}
}

func TestEventChainDoesNotHang(t *testing.T) {
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&Person{})

	// eventChan can hold 5 events before it get full and blocks.
	eventChan := make(chan eventsourcing.Event, 5)
	doneChan := make(chan struct{})
	f := func(e eventsourcing.Event) {
		eventChan <- e
	}

	// for every AgedOnYear create a new person and make it grow one year older
	go func() {
		for e := range eventChan {
			switch e.Data().(type) {
			case *AgedOneYear:
				person, err := CreatePerson("kalle")
				if err != nil {
					t.Fatal(err)
				}
				person.GrowOlder()
				repo.Save(person)
			}
		}
		close(doneChan)
	}()

	// create the initial person and setup event subscription on the specific person events
	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	s := repo.Subscribers().AggregateID(f, person)
	defer s.Close()

	// subscribe to all events and filter out AgedOneYear
	ageCounter := 0
	s2 := repo.Subscribers().All(func(e eventsourcing.Event) {
		switch e.Data().(type) {
		case *AgedOneYear:
			// will match three times on the initial person and one each on the resulting AgedOneYear event
			ageCounter++
		}
	})
	defer s2.Close()

	person.GrowOlder()
	person.GrowOlder()
	person.GrowOlder()

	err = repo.Save(person)
	if err != nil {
		t.Fatal("could not save aggregate")
	}
	close(eventChan)
	<-doneChan
	if ageCounter != 6 {
		t.Errorf("wrong number in ageCounter expected 6, got %v", ageCounter)
	}
}

func TestSaveWhenAggregateNotRegistered(t *testing.T) {
	repo := eventsourcing.NewEventRepository(memory.Create())

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if !errors.Is(err, eventsourcing.ErrAggregateNotRegistered) {
		t.Fatalf("could save aggregate that was not registered, err: %v", err)
	}
}

func TestSaveWhenEventNotRegistered(t *testing.T) {
	repo := eventsourcing.NewEventRepository(memory.Create())
	repo.Register(&PersonNoRegisterEvents{})

	person, err := CreatePersonNoRegisteredEvents("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.Save(person)
	if !errors.Is(err, eventsourcing.ErrEventNotRegistered) {
		t.Fatalf("could save aggregate events that was not registered, err: %v", err)
	}
}

// Person aggregate
type PersonNoRegisterEvents struct {
	eventsourcing.AggregateRoot
	Name string
	Age  int
	Dead int
}

// Born event
type BornNoRegisteredEvents struct {
	Name string
}

// CreatePersonNoRegisteredEvents constructor for the PersonNoRegisteredEvents
func CreatePersonNoRegisteredEvents(name string) (*PersonNoRegisterEvents, error) {
	if name == "" {
		return nil, errors.New("name can't be blank")
	}
	person := PersonNoRegisterEvents{}
	person.TrackChange(&person, &BornNoRegisteredEvents{Name: name})
	return &person, nil
}

// Transition the person state dependent on the events
func (person *PersonNoRegisterEvents) Transition(event eventsourcing.Event) {
	switch e := event.Data().(type) {
	case *BornNoRegisteredEvents:
		person.Age = 0
		person.Name = e.Name
	}
}

// Register bind the events to the repository when the aggregate is registered.
func (person *PersonNoRegisterEvents) Register(f eventsourcing.RegisterFunc) {
}
