[![Build Status](https://travis-ci.org/hallgren/eventsourcing.svg?branch=master)](https://travis-ci.org/hallgren/eventsourcing)
[![Go Report Card](https://goreportcard.com/badge/github.com/hallgren/eventsourcing)](https://goreportcard.com/report/github.com/hallgren/eventsourcing)
[![codecov](https://codecov.io/gh/hallgren/eventsourcing/branch/master/graph/badge.svg)](https://codecov.io/gh/hallgren/eventsourcing)

# Overview

This package is an experiment to try to generialize [@jen20's](https://github.com/jen20) way of implementing event sourcing. You can find the original blog post [here](https://jen20.dev/post/event-sourcing-in-go/) and github repo [here](https://github.com/jen20/go-event-sourcing-sample).

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

The aggregate needs to implement the `Transition(event eventsourcing.Event)` and `Regsiter(r eventsourcing.RegsiterFunc)` methods to fulfill the aggregate interface. This methods define how events are transformed to build the aggregate state and which events to register into the repository. 

Example of the Transition method from the `Person` aggregate.

```go
// Transition the person state dependent on the events
func (person *Person) Transition(event eventsourcing.Event) {
    switch e := event.Data().(type) {
    case *Born:
            person.Age = 0
            person.Name = e.Name
    case *AgedOneYear:
            person.Age += 1
    }
}
```

The `Born` event sets the `Person` property `Age` and `Name`, and the `AgedOneYear` adds one year to the `Age` property. This makes the state of the aggregate flexible and could easily change in the future if required.

Example or the Register method:

```go
// Register callback method that register Person events to the repository
func (person *Person) Register(r eventsourcing.RegsiterFunc) {
    r(&Born{}, &AgedOneYear{})
}
```

The `Born` and `AgedOneYear` events are now registered to the repository when the aggregate is registered.

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
	person.TrackChange(&person, &Born{Name: name})
	return &person, nil
}
```

When a person is created, more events could be created via methods on the `Person` aggregate. Below is the `GrowOlder` method which in turn triggers the event `AgedOneYear`. This event is tracked on the person aggregate.

```go
// GrowOlder command
func (person *Person) GrowOlder() {
	person.TrackChange(person, &AgedOneYear{})
}
```

Internally the `TrackChange` methods calls the `Transition` method on the aggregate to transform the aggregate based on the newly created event.

To bind metadata to events use the `TrackChangeWithMetadata` method.
  
The internal `Event` looks like this.

```go
type Event struct {
    // aggregate identifier 
    aggregateID string
    // the aggregate version when this event was created
    version         Version
    // the global version is based on all events (this value is only set after the event is saved to the event store) 
    globalVersion   Version
    // aggregate type (Person in the example above)
    aggregateType   string
    // UTC time when the event was created  
    timestamp       time.Time
    // the specific event data specified in the application (Born{}, AgedOneYear{})
    data            interface{}
    // data that don´t belongs to the application state (could be correlation id or other request references)
    metadata        map[string]interface{}
}
```

To access properties on the event you can use the corresponding methods exposing them, e.g `AggregateID()`. This prevent external parties to modify the event from the outside.
### Aggregate ID

The identifier on the aggregate is default set by a random generated string via the crypt/rand pkg. It is possible to change the default behaivior in two ways.

* Set a specific id on the aggregate via the SetID func.

```go
var id = "123"
person := Person{}
err := person.SetID(id)
```

* Change the id generator via the global eventsourcing.SetIDFunc function.

```go
var counter = 0
f := func() string {
	counter++
	return fmt.Sprint(counter)
}

eventsourcing.SetIDFunc(f)
```

## Repository

The repository is used to save and retrieve aggregates. The main functions are:

```go
// saves the events on the aggregate
Save(a aggregate) error

// retrieves and build an aggregate from events based on its identifier
// possible to cancel from the outside
GetWithContext(ctx context.Context, id string, a aggregate) error

// retrieves and build an aggregate from events based on its identifier
Get(id string, a aggregate) error
```

The repository constructor input values is an event store, this handles the reading and writing of events and builds the aggregate based on the events.

```go
repo := NewRepository(eventStore EventStore) *Repository
```

Here is an example of a person being saved and fetched from the repository.

```go
// the person aggregate has to be registered in the repository
repo.Register(&Person{})

person := person.CreatePerson("Alice")
person.GrowOlder()
repo.Save(person)
twin := Person{}
repo.Get(person.Id, &twin)
```

### Event Store

The only thing an event store handles are events, and it must implement the following interface.

```go
// saves events to the under laying data store.
Save(events []eventsourcing.Event) error

// fetches events based on identifier and type but also after a specific version. The version is used to load event that happened after a snapshot was taken.
Get(id string, aggregateType string, afterVersion eventsourcing.Version) (eventsourcing.Iterator, error)
```

Currently, there are three implementations.

* SQL
* Bolt
* Event Store DB
* RAM Memory

Post release v0.0.7 event stores `bbolt`, `sql` and `esdb` are their own submodules.
This reduces the dependency graph of the `github.com/hallgren/eventsourcing` module, as each submodule contains their own dependencies not pollute the main module.
Submodules needs to be fetched separately via go get.

`go get github.com/hallgren/eventsourcing/eventstore/sql`  
`go get github.com/hallgren/eventsourcing/eventstore/bbolt`
`go get github.com/hallgren/eventsourcing/eventstore/esdb`

The memory based event store is part of the main module and does not need to be fetched separately.

## Serializer

To store events they have to be serialised into `[]byte`. Default the `encoding/json` is used but it could be replaced on the repoository by setting the `repo.Serializer` and `repo.Deserializer` function properties.

### Event Subscription

The repository expose four possibilities to subscribe to events in realtime as they are saved to the repository.

`All(func (e Event)) *subscription` subscribes to all events.

`AggregateID(func (e Event), events ...aggregate) *subscription` events bound to specific aggregate based on type and identity.
This makes it possible to get events pinpointed to one specific aggregate instance.

`Aggregate(func (e Event), aggregates ...aggregate) *subscription` subscribes to events bound to specific aggregate type. 
 
`Event(func (e Event), events ...interface{}) *subscription` subscribes to specific events. There are no restrictions that the events need
to come from the same aggregate, you can mix and match as you please.

`Name(f func(e Event), aggregate string, events ...string) *subscription` subscribes to events based on aggregate type and event name.

The subscription is realtime and events that are saved before the call to one of the subscribers will not be exposed via the `func(e Event)` function. If the application 
depends on this functionality make sure to call Subscribe() function on the subscriber before storing events in the repository. 

The event subscription enables the application to make use of the reactive patterns and to make it more decoupled. Check out the [Reactive Manifesto](https://www.reactivemanifesto.org/) 
for more detailed information. 

Example on how to set up the event subscription and consume the event `FrequentFlierAccountCreated`

```go
// Setup a memory based repository
repo := eventsourcing.NewRepository(memory.Create())
repo.Register(&FrequentFlierAccountAggregate{})

// subscriber that will trigger on every saved events
s := repo.Subscribers().All(func(e eventsourcing.Event) {
    switch e := event.Data().(type) {
        case *FrequentFlierAccountCreated:
            // e now have type info
            fmt.Println(e)
        }
    }
)

// stop subscription
s.Close()
```

## Custom made components

Parts of this package may not fulfill your application need, either it can be that the event store uses the wrong database for storage.

#### Event Store

A custom-made event store has to implement the following functions to fulfill the interface in the repository.  

```go
type EventStore interface {
    Save(events []Event) error
    Get(id string, aggregateType string, afterVersion Version) (Iterator, error)
}
```
