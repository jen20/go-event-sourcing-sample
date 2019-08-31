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
	aggregateRoot eventsourcing.AggregateRoot
	Device
}

type NameSet struct {
	Name string
}

func (d *DeviceAggregate) SetName(name string) {
	d.aggregateRoot.TrackChange(NameSet{Name: name})
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
	device.aggregateRoot.SetParent(&device)
	device.SetName("New name")
	err := repo.Save(&device.aggregateRoot)
	if err != nil {
		t.Fatal("could not save device")
	}
	copyDevice := DeviceAggregate{}
	copyDevice.aggregateRoot.SetParent(&copyDevice)
	err = repo.Get(device.aggregateRoot.ID(), &copyDevice.aggregateRoot)
	if err != nil {
		t.Fatal("could not get aggregate")
	}

	if device.aggregateRoot.Version() != copyDevice.aggregateRoot.Version() {
		t.Fatalf("Wrong version org %q copy %q", device.aggregateRoot.Version(), copyDevice.aggregateRoot.Version())
	}
}
