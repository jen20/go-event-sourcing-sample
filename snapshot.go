package eventsourcing

type snapshotStore interface {
	Save(id AggregateRootID, a interface{}) error
	Get(id AggregateRootID, a interface{}) error
}

type Snapshot struct {
	store snapshotStore
}

func NewSnapshot(store snapshotStore) *Snapshot {
	return &Snapshot{store:store}
}

func (s *Snapshot) Save(id AggregateRootID, a interface{}) error {
	return s.store.Save(id, a)
}

func (s *Snapshot) Get(id AggregateRootID, a interface{}) error {
	return s.store.Get(id, a)
}