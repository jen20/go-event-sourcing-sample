###Event Sourcing in Go

This is the repository for the code which accompanies my post on Event Sourcing in Go. The original post is [here](http://jen20.com/2015/02/08/event-sourcing-in-go.html).

Remade by @hallgren

# Overview

This package is a try to implement event sourcing with the aggregate concept in mind from Domain Driven Design in mind.

# Event Sourcing

Event Sourcing is a technique to store changes to domain entities as a sequence of event that sums up to the final state
of the entity.

## Aggregate Root

The aggregate root is there the events are bound. It takes in commands that transform to one or more events if the rules 
in the aggregate are ok.

## Aggregate Event

# Repository

## Event Store

## Snapshot Store

## Serializer