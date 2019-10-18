# Overview

This package is an experiment to try to generialize @jen20's way of implementing event sourcing. You can find the original blog post [here](http://jen20.com/2015/02/08/event-sourcing-in-go.html) and github repo [here](https://github.com/jen20/go-event-sourcing-sample)

# Event Sourcing

Event Sourcing is a technique to make it possible to capture all changes to an application state as a sequence of events. Read more about it [here](https://martinfowler.com/eaaDev/EventSourcing.html)

## Aggregate Root

The aggregate root is the central point where events are bound. The aggregate struct needs to be embedded `eventsourcing.AggreateRoot` to get the aggregate behaviors.

Below a Person aggregate where the Aggregate Root is embedded next to the Name and Age properties.

```
type Person struct {
	eventsourcing.AggregateRoot
	Name string
	Age  int
}
```

The aggregate needs to implement the `Transition(event eventsourcing.Event)` function to fulfill the aggregate interface. This function define how events are transformed to build the aggregate state.

Example of the Transition function from the Person aggregate.

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

In this example we can see that the `Born` event sets the Person property Age and Name and that the `AgedOneYear` adds one year to the Age property. This makes the state of the aggregate very flexible and it could change in the future if required.

## Aggregate Event

An event is a clean struct with exported properties that contains the state of the event.

Example of two events from the Person aggregate.

```
// Initial event
type Born struct {
        Name string
}

// event that happens once a year
type AgedOneYear struct {}
```

When an aggregate is first created an event is needed to initialize the state of the aggregate. No event, no aggregate. Below is an example where a constructor is defined that returns the Person aggregate and inside it binds an event via the TrackChange function. It's possible to define rules that the aggregate must uphold before an event is created, in this case the person's name can´t be blank.

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

When a person is created more events could be created via functions on the Person aggregate. Below is the GrowOlder function that results in the event AgedOneYear. This event is tracked on the person aggregate.

```
// GrowOlder command
func (person *Person) GrowOlder() error {
	return person.TrackChange(person, &AgedOneYear{})
}
```

Internally the TrackChange functions calls the Transition function in the aggregate to transform the aggregate based on the newly created event.

The internal Event looks like this.

```
type Event struct {
	AggregateRootID AggregateRootID // aggregate identifier 
	Version         Version // aggregate version
	Reason          string // name of the event (Born / AgedOneYear in the example above) 
	AggregateType   string // aggregate name (Person in the example above)
	Data            interface{} // The data from the event (Born{}, AgedOneYear{})
	MetaData        map[string]interface{} // extra data that don´t belongs to the application state (correlation id or other request reference)
}
```

# Repository

The repository is used to save and retrieve aggregates. Its main functions are:

```
Save(aggregate aggregate) error // stores the aggregates events
Get(id string, aggregate aggregate) error // retrieves and build an aggregate from events based on an identifier
```
It is possible to save a snapshot of an aggregate reducing the amount of event needed to be fetched and applied.

```
SaveSnapshot(aggregate aggregate) error // saves the aggregate (no unsaved events are allowed on the aggregate when doing this operation)
```

The constructor of the repository takes in an event store and a snapshot store that handles the reading and writing of the events and snapshots. We will dig deeper on the internals below

```
NewRepository(eventStore eventStore, snapshotStore snapshotStore) *Repository
```

Here is an example of a person being saved and fetched from the repository.

```
person := person.CreatePerson("Alice")
person.GrowOlder()
repo.Save(person)
twin := Person{}
repo.Get(person.Id, &twin)
```

## Event Store

The only thing an event store must handle are events. An event has to support the following interface.

```
Save(events []eventsourcing.Event) error // saves events to an underlaying data store.
Get(id string, aggregateType string, afterVersion eventsourcing.Version) ([]eventsourcing.Event, error) // fetches events based on identifier and type but also after a specific version. This is used to only load event that has happened after a snapshot is taken.
```

The event store also has a function that fetch events that are not based on identifier or type. It could be used to build separate representations often called projections.

```
GlobalGet(start int, count int) []eventsourcing.Event
```

Currently there are three event store implementations.

* SQL
* Bolt
* RAM Memory

## Snapshot Store

A snapshot store handles snapshots. The properties of an aggregate have to be exported for them to be saved in the snapshot.

```
Get(id string, a interface{}) error` // get snapshot by identifier
Save(id string, a interface{}) error` // saves snapshot
```

## Serializer

The event and snapshot stores all depend on serializer to transform events and aggregates []byte.

```
SerializeSnapshot(interface{}) ([]byte, error)
DeserializeSnapshot(data []byte, a interface{}) error
SerializeEvent(event eventsourcing.Event) ([]byte, error)
DeserializeEvent(v []byte) (event eventsourcing.Event, err error)
```

### JSON

The json serializer has the following extra function.

```
Register(aggregate aggregate, events ...interface{}) error
```

It needs to register the aggregate and its events for it to maintain correct type info after Marshall.

```
j := json.New()
err := j.Register(&Person{}, &Born{}, &AgedOneYear{})
```

### Unsafe

The unsafe serializer stores the underlying memory representation of a struct directly. This makes it as its name implies unsafe to use if you are unsure what you are doing. [Here](https://youtu.be/4xB46Xl9O9Q?t=610) is the video that explains the reason for this serializer.
