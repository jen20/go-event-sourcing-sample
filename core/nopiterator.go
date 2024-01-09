package core

// NopIterator returns no data
type NopIterator struct{}

func (ni NopIterator) Next() bool {
	return false
}

func (ni NopIterator) Value() (Event, error) {
	return Event{}, nil
}

func (ni NopIterator) Close() {
	return
}
