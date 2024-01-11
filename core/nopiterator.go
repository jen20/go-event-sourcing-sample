package core

// ZeroIterator returns no data
type ZeroIterator struct{}

func (ni ZeroIterator) Next() bool {
	return false
}

func (ni ZeroIterator) Value() (Event, error) {
	return Event{}, nil
}

func (ni ZeroIterator) Close() {
	return
}
