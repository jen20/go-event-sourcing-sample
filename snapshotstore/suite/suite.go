package suite

import (
	"testing"

	"github.com/hallgren/eventsourcing"
)

type storeProvider interface {
	// Setup runs before all test and provides the store to test.
	Setup() (eventsourcing.SnapshotStore, error)
	// Cleanup is called after each test.
	Cleanup()
	// Teardown is called after all test has passed.
	Teardown()
}

func Test(t *testing.T, provider storeProvider) {
	tests := []struct {
		title string
		run   func(t *testing.T, es eventsourcing.SnapshotStore)
	}{
		{"Basics", TestSnapshot},
	}
	store, err := provider.Setup()
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Teardown()
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			test.run(t, store)
			provider.Cleanup()
		})
	}
}

func TestSnapshot(t *testing.T, snapshot eventsourcing.SnapshotStore) {
	snap := eventsourcing.Snapshot{
		Version:       10,
		GlobalVersion: 5,
		ID:            "123",
		Type:          "Person",
		State:         []byte{},
	}

	err := snapshot.Save(snap)
	if err != nil {
		t.Fatal(err)
	}

	snap2, err := snapshot.Get("123", "Person")
	if err != nil {
		t.Fatalf("could not get snapshot %v", err)
	}
	if snap.ID != snap2.ID {
		t.Fatalf("wrong ID in snapshot %q expected: %q", snap.ID, snap2.ID)
	}
	if snap.Type != snap2.Type {
		t.Fatalf("wrong Type in snapshot %q expected: %q", snap.Type, snap2.Type)
	}
	if snap.GlobalVersion != snap2.GlobalVersion {
		t.Fatalf("wrong GlobalVersion in snapshot %q expected: %q", snap.GlobalVersion, snap2.GlobalVersion)
	}
	if snap.Version != snap2.Version {
		t.Fatalf("wrong Version in snapshot %q expected: %q", snap.Version, snap2.Version)
	}
	if string(snap.State) != string(snap2.State) {
		t.Fatalf("wrong State in snapshot %q expected: %q", snap.State, snap2.State)
	}
}
