package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore/memory"
	"testing"
)

type Device struct {
	name string
}

type DeviceAggregate struct {
	eventsourcing.AggregateRoot
	Device
}

type NameSet struct {
	Name string
}

func (d *DeviceAggregate) SetName(name string) {
	d.TrackChange(NameSet{Name: name})
}

// Transition the person state dependent on the events
func (d *DeviceAggregate) Transition(event eventsourcing.Event) {
	switch e := event.Data.(type) {
	case NameSet:
		d.name = e.Name
	}
}

func TestSaveAndGetAggregate(t *testing.T) {
	repo := eventsourcing.NewRepository(memory.Create())

	device := DeviceAggregate{}
	repo.New(&device)
	device.SetName("New name")
	err := repo.Save(&device)
	if err != nil {
		t.Fatal("could not save device")
	}
	copyDevice := DeviceAggregate{}
	err = repo.Get(device.ID(), &copyDevice)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	if device.Version() != copyDevice.Version() {
		t.Fatalf("Wrong version org %q copy %q", device.Version(), copyDevice.Version())
	}
}

func TestCreateAggregate(t *testing.T) {
	repo := eventsourcing.NewRepository(memory.Create())

	device := DeviceAggregate{}
	repo.New(&device)
	if device.Parent() == nil {
		t.Fatalf("Parent not set")
	}
	device.SetName("test")
	repo.Save(&device)

}