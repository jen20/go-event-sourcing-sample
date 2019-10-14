###Event Sourcing in Go

This is the repository for the code which accompanies my post on Event Sourcing in Go. The original post is [here](http://jen20.com/2015/02/08/event-sourcing-in-go.html).

Remade by @hallgren

# Overview

This package is a try to implement an event sourcing solution with the aggregate concept in mind from Domain Driven Design.

# Event Sourcing

Event Sourcing is a technique to store changes to domain entities as a series of events. The events togheter builds up the final state of the entities in the system.

## Aggregate Root

The aggregate root is the central point where events are bound. The aggregate struct needs to embedded `eventsourcing.AggreateRoot` to get the aggregate behaiviors.

example Person aggregate

```
type Person struct {
	eventsourcing.AggregateRoot
	Name string
	Age  int
}
```

The aggregate also need to implement the `Transition(event eventsourcing.Event)` function. It define how events are transformed to represent the final aggregate state.

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

An event is a clean struct with exported properties that holds the state of the event. 

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


# Repository

## Event Store

## Snapshot Store

## Serializer
