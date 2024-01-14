package eventsourcing_test

import (
	"context"
	"testing"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	snap "github.com/hallgren/eventsourcing/snapshotstore/memory"
)

func TestSaveAndGetSnapshot(t *testing.T) {
	eventrepo := eventsourcing.NewEventRepository(memory.Create())
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

	if person.ID() != twin.ID() {
		t.Fatalf("Wrong id org %q copy %q", person.ID(), twin.ID())
	}

	if person.Name != twin.Name {
		t.Fatalf("Wrong name org: %q copy %q", person.Name, twin.Name)
	}
}

func TestSaveSnapshotWithUnsavedEvents(t *testing.T) {
	eventrepo := eventsourcing.NewEventRepository(memory.Create())
	eventrepo.Register(&Person{})

	snapshotrepo := eventsourcing.NewSnapshotRepository(snap.Create(), eventrepo)

	person, err := CreatePerson("kalle")
	if err != nil {
		t.Fatal(err)
	}
	err = snapshotrepo.SaveSnapshot(person)
	if err == nil {
		t.Fatalf("should not be able to save snapshot with unsaved events")
	}
}

type snapshot struct {
	eventsourcing.AggregateRoot
	unexported string
	Exported   string
}

type Event struct{}
type Event2 struct{}

func New() *snapshot {
	s := snapshot{}
	s.TrackChange(&s, &Event{})
	return &s
}

func (s *snapshot) Command() {
	s.TrackChange(s, &Event2{})
}

func (s *snapshot) Transition(e eventsourcing.Event) {
	switch e.Data().(type) {
	case *Event:
		s.unexported = "unexported"
		s.Exported = "Exported"
	case *Event2:
		s.unexported = "unexported2"
		s.Exported = "Exported2"
	}
}

// Register bind the events to the repository when the aggregate is registered.
func (s *snapshot) Register(f eventsourcing.RegisterFunc) {
	f(&Event{}, &Event2{})
}

type snapshotInternal struct {
	UnExported string
	Exported   string
}

func (s *snapshot) SerializeSnapshot(m eventsourcing.MarshalFunc) ([]byte, error) {
	snap := snapshotInternal{
		UnExported: s.unexported,
		Exported:   s.Exported,
	}
	return m(snap)
}

func (s *snapshot) DeserializeSnapshot(m eventsourcing.UnmarshalFunc, b []byte) error {
	snap := snapshotInternal{}
	err := m(b, &snap)
	if err != nil {
		return err
	}
	s.unexported = snap.UnExported
	s.Exported = snap.Exported
	return nil
}

func TestSnapshotNoneExported(t *testing.T) {
	// use repo to reset events on person to be able to save snapshot
	eventRepo := eventsourcing.NewEventRepository(memory.Create())
	snapshotRepo := eventsourcing.NewSnapshotRepository(snap.Create(), eventRepo)
	snapshotRepo.Register(&snapshot{})

	snap := New()
	err := snapshotRepo.Save(snap)
	if err != nil {
		t.Fatal(err)
	}

	snap.Command()
	snapshotRepo.Save(snap)

	snap2 := snapshot{}
	err = snapshotRepo.GetWithContext(context.Background(), snap.ID(), &snap2)
	if err != nil {
		t.Fatal(err)
	}

	if snap.unexported != snap2.unexported {
		t.Fatalf("none exported value differed %s %s", snap.unexported, snap2.unexported)
	}

	if snap.Exported != snap2.Exported {
		t.Fatalf("exported value differed %s %s", snap.Exported, snap2.Exported)
	}
}
