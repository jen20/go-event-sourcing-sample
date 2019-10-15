###Event Sourcing in Go

This is the repository for the code which accompanies my post on Event Sourcing in Go. The original post is [here](http://jen20.com/2015/02/08/event-sourcing-in-go.html).

Remade by @hallgren

# Overview

This package is a try to implement an event sourcing solution with the aggregate concept in mind from Domain Driven Design.

# Event Sourcing

Event Sourcing is a technique to store changes to domain entities as a series of events. The events togheter builds up the final state of the system.

## Aggregate Root

The aggregate root is the central point where events are bound. The aggregate struct needs to embedded `eventsourcing.AggreateRoot` to get the aggregate behaiviors.

Below a Person aggregate where the Aggregate Root is embedded next to the Name and Age properties.

```
type Person struct {
	eventsourcing.AggregateRoot
	Name string
	Age  int
}
```

The aggregate needs to implement the `Transition(event eventsourcing.Event)` function to fulfill the aggregate interface. This function define how events are transformed to build the aggregate state.

example of the Transition function from the Person aggregate

```
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
```

In the example we can see that the `Born` event sets the Person property Age and Name and that the `AgedOneYear` adds one year to the Age property. This makes the state of the aggregate very flexible and it could change in the future if required.

## Aggregate Event

An event is a clean struct with exported properties that containce the state of the event.

example of two events from the Person aggregate

```
// Initial event
type Born struct {
        Name string
}

// event that happens once a year
type AgedOneYear struct {
}

```

When an aggregate is first created an event is needed to initialize the state of the aggregate. No event no aggregate. Below is an example where a constructor is defined that returns the Person aggregate and inside binds an event via the TrackChange function. Its also possible to define rules that the aggregate must uphold before an event is created, in this case a name != "".

```
// CreatePerson constructor for Person
func CreatePerson(name string) (*Person, error) {
	if name == "" {
		return nil, fmt.Errorf("Name can't be blank")
	}
	person := Person{}
	err := person.TrackChange(&person, &Born{Name: name})
	if err != nil {
		return nil, err
	}
	return &person, nil
}
```

when the person is created more events could be created via functions on the Person aggregate. Below is the GrowOlder command that results in the event AgedOneYear event.

```
// GrowOlder command
func (person *Person) GrowOlder() error {
	return person.TrackChange(person, &AgedOneYear{})
}
```

Internally the TrackChange functions calls the Transition function in the aggregate to transform the aggregate based on the newly created event.

The internal Event looks like this

```
type Event struct {
	AggregateRootID AggregateRootID // aggregate identifier 
	Version         Version // aggregate version
	Reason          string // name of the event (Born / AgedOneYear in the example above) 
	AggregateType   string // aggregate name (Person)
	Data            interface{} // The data from the event defined on the aggregate level
	MetaData        map[string]interface{} // extra data that not belongs to the application state (correlation id or other request reference)
}
```

# Repository

The repository expose functions to save and retrieve aggregates. Its main functions are:

`Save(aggregate aggregate) error` stores the aggregates events
`Get(id string, aggregate aggregate) error` retrieves and build an aggregate from events based on the aggregates identifier

It is possible to save a snapshot of the aggregate reducing the amount of event needed to be fetched and applied to the aggregate.

`SaveSnapshot(aggregate aggregate) error` saves the hole aggregate (no unsaved events are allowed in the aggregate in the operation) 

The constructor takes in an event store and a snapshot store that handles the reading and writing of the events and snapshots.

`NewRepository(eventStore eventStore, snapshotStore snapshotStore) *Repository`

## Event Store

## Snapshot Store

## Serializer
