[![Build Status](https://travis-ci.org/hallgren/eventsourcing.svg?branch=master)](https://travis-ci.org/hallgren/eventsourcing)
[![Go Report Card](https://goreportcard.com/badge/github.com/hallgren/eventsourcing)](https://goreportcard.com/report/github.com/hallgren/eventsourcing)
[![codecov](https://codecov.io/gh/hallgren/eventsourcing/branch/master/graph/badge.svg)](https://codecov.io/gh/hallgren/eventsourcing)



# Overview

This package is an experiment to try to generialize [@jen20's](https://github.com/jen20) way of implementing event sourcing. You can find the original blog post [here](http://jen20.com/2015/02/08/event-sourcing-in-go.html) and github repo [here](https://github.com/jen20/go-event-sourcing-sample).

## Event Sourcing

[Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html) is a technique to make it possible to capture all changes to an application state as a sequence of events.

### Aggregate Root

The *aggregate root* is the central point where events are bound. The aggregate struct needs to embed `eventsourcing.AggreateRoot` to get the aggregate behaviors.

Below, a *Person* aggregate where the Aggregate Root is embedded next to the `Name` and `Age` properties.

```go
type Person struct {
	eventsourcing.AggregateRoot
	Name string
	Age  int
}
```

The aggregate needs to implement the `Transition(event eventsourcing.Event)` function to fulfill the aggregate interface. This function define how events are transformed to build the aggregate state.

Example of the Transition function from the `Person` aggregate.

```go
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

In this example we can see that the `Born` event sets the `Person` property `Age` and `Name`, and that the `AgedOneYear` adds one year to the `Age` property. This makes the state of the aggregate flexible and could easily change in the future if required.

### Aggregate Event

An event is a clean struct with exported properties that contains the state of the event.

Example of two events from the `Person` aggregate.

```go
// Initial event
type Born struct {
        Name string
}

// Event that happens once a year
type AgedOneYear struct {}
```

When an aggregate is first created, an event is needed to initialize the state of the aggregate. No event, no aggregate. Below is an example of a constructor that returns the `Person` aggregate and inside it binds an event via the `TrackChange` function. It's possible to define rules that the aggregate must uphold before an event is created, in this case the person's name must not be blank.

```go
// CreatePerson constructor for Person
func CreatePerson(name string) (*Person, error) {
	if name == "" {
		return nil, errors.New("name can't be blank")
	}
	person := Person{}
	err := person.TrackChange(&person, &Born{Name: name})
	if err != nil {
		return nil, err
	}
	return &person, nil
}
```

When a person is created, more events could be created via functions on the `Person` aggregate. Below is the `GrowOlder` function which in turn triggers the event `AgedOneYear`. This event is tracked on the person aggregate.

```go
// GrowOlder command
func (person *Person) GrowOlder() error {
	return person.TrackChange(person, &AgedOneYear{})
}
```

Internally the `TrackChange` functions calls the `Transition` function on the aggregate to transform the aggregate based on the newly created event.

The internal `Event` looks like this.

```go
type Event struct {
	AggregateRootID AggregateRootID // aggregate identifier 
	Version         Version // aggregate version
	Reason          string // name of the event (Born / AgedOneYear in the example above) 
	AggregateType   string // aggregate name (Person in the example above)
	Data            interface{} // The data from the event (Born{}, AgedOneYear{})
	MetaData        map[string]interface{} // extra data that don´t belongs to the application state (correlation id or other request reference)
}
```

## Repository

The repository is used to save and retrieve aggregates. The main functions are:

```go
Save(aggregate aggregate) error // stores the aggregates events
Get(id string, aggregate aggregate) error // retrieves and build an aggregate from events based on an identifier
```

It is possible to save a snapshot of an aggregate reducing the amount of event needed to be fetched and applied.

```go
SaveSnapshot(aggregate aggregate) error // saves the aggregate (en error will be returned if there are unsaved events on the aggregate when doing this operation)
```

The repository constructor input values is an event store and a snapshot store, this handles the reading and writing of events and snapshots. We will dig deeper on the internals below.

```go
NewRepository(eventStore eventStore, snapshotStore snapshotStore) *Repository
```

Here is an example of a person being saved and fetched from the repository.

```go
person := person.CreatePerson("Alice")
person.GrowOlder()
repo.Save(person)
twin := Person{}
repo.Get(person.Id, &twin)
```

### Event Store

The only thing an event store handles are events and it must implement the following interface.

```go
Save(events []eventsourcing.Event) error // saves events to an underlaying data store.
Get(id string, aggregateType string, afterVersion eventsourcing.Version) ([]eventsourcing.Event, error) // fetches events based on identifier and type but also after a specific version. The version is used to only load event that has happened after a snapshot is taken.
```

Appart the manadatory `Get` and `Save`functions an event store could also implement the `GlobalGet` function for fetching events based on the order they were saved. This function makes it possible to build separate representations often called projections 

The event store also has a function that fetch events that are not based on identifier or type. It could be used to build separate representations often called projections or [CQRS](https://martinfowler.com/bliki/CQRS.html).

```go
GlobalGet(start uint64, count int) []eventsourcing.Event
```

Currently there are three implementations.

* SQL
* Bolt
* RAM Memory

### Snapshot Store

A snapshot store handles snapshots. The properties of an aggregate have to be exported for them to be saved in the snapshot.

```
Get(id string, a interface{}) error` // get snapshot by identifier
Save(id string, a interface{}) error` // saves snapshot
```

## Serializer

The event and snapshot stores depends on a serializer to transform events and aggregates to the `[]byte` data type.

```go
SerializeSnapshot(interface{}) ([]byte, error)
DeserializeSnapshot(data []byte, a interface{}) error
SerializeEvent(event eventsourcing.Event) ([]byte, error)
DeserializeEvent(v []byte) (event eventsourcing.Event, err error)
```

### JSON

The JSON serializer has the following extra function.

```go
Register(aggregate aggregate, events ...interface{}) error
```

It needs to register the aggregate together with the events that belongs to it. It has to do this to maintain correct type info after Marshall.

```go
j := json.New()
err := j.Register(&Person{}, &Born{}, &AgedOneYear{})
```

### Unsafe

The unsafe serializer stores the underlying memory representation of a struct directly. This makes it as its name implies, unsafe to use if you are unsure what you are doing. [Here](https://youtu.be/4xB46Xl9O9Q?t=610) is the video that explains the reason for this serializer.
