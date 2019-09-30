package main

import (
	"fmt"
	"github.com/hallgren/eventsourcing"
)

// FrequentFlierAccount represents the state of an instance of the frequent flier
// account aggregate. It tracks changes on itself in the form of domain events.
type FrequentFlierAccount struct {
	miles      int
	tierPoints int
	status     Status
}

type FrequentFlierAccountAggregate struct {
	eventsourcing.AggregateRoot
	FrequentFlierAccount
}

// Status represents the Red, Silver or Gold tier level of a FrequentFlierAccount
type Status int

const (
	StatusRed    Status = iota
	StatusSilver Status = iota
	StatusGold   Status = iota
)

// go:generate stringer -type=Status

type FrequentFlierAccountCreated struct {
	AccountId         string
	OpeningMiles      int
	OpeningTierPoints int
}

type StatusMatched struct {
	NewStatus Status
}

type FlightTaken struct {
	MilesAdded      int
	TierPointsAdded int
}

type PromotedToGoldStatus struct {
}

// CreateFrequentFlierAccount constructor
func CreateFrequentFlierAccount(id string) *FrequentFlierAccountAggregate {
	self := FrequentFlierAccountAggregate{}
	//self.aggregateRoot.SetID(id)
	self.TrackChange(&self, &FrequentFlierAccountCreated{OpeningMiles: 0, OpeningTierPoints: 0})
	return &self
}

// NewFrequentFlierAccountFromHistory creates a FrequentFlierAccount given a history
// of the changes which have occurred for that account.
func NewFrequentFlierAccountFromHistory(events []eventsourcing.Event) *FrequentFlierAccountAggregate {
	state := FrequentFlierAccountAggregate{}
	state.BuildFromHistory(&state, events)
	return &state
}

// RecordFlightTaken is used to record the fact that a customer has taken a flight
// which should be attached to this frequent flier account. The number of miles and
// tier points which apply are calculated externally.
//
// If recording this flight takes the account over a status boundary, it will
// automatically upgrade the account to the new status level.
func (self *FrequentFlierAccountAggregate) RecordFlightTaken(miles int, tierPoints int) {
	err := self.TrackChange(self, &FlightTaken{MilesAdded: miles, TierPointsAdded: tierPoints})
	if err != nil {
		panic(err)
	}

	if self.tierPoints > 10 && self.status != StatusSilver {
		err = self.TrackChange(self, &StatusMatched{NewStatus: StatusSilver})
		if err != nil {
			panic(err)
		}
	}

	if self.tierPoints > 20 && self.status != StatusGold {
		err := self.TrackChange(self, &PromotedToGoldStatus{})
		panic(err)
	}
}

// Transition implements the pattern match against event types used both as part
// of the fold when loading from history and when tracking an individual change.
func (state *FrequentFlierAccountAggregate) Transition(event eventsourcing.Event) {

	switch e := event.Data.(type) {

	case *FrequentFlierAccountCreated:
		state.miles = e.OpeningMiles
		state.tierPoints = e.OpeningTierPoints
		state.status = StatusRed

	case *StatusMatched:
		state.status = e.NewStatus

	case *FlightTaken:
		state.miles = state.miles + e.MilesAdded
		state.tierPoints = state.tierPoints + e.TierPointsAdded

	case *PromotedToGoldStatus:
		state.status = StatusGold
	}
}

// String implements Stringer for FrequentFlierAccount instances.
func (a FrequentFlierAccountAggregate) String() string {

	var reason string
	var aggregateType string

	if len(a.Events()) > 0 {
		reason = a.Events()[0].Reason
		aggregateType = a.Events()[0].AggregateType
	} else {
		reason = "No reason"
		aggregateType = "No aggregateType"
	}

	format := `FrequentFlierAccount: %s
	Miles: %d
	TierPoints: %d
	Status: %s
	(First Reason: %s)
	(First AggregateType %s)
	(Pending Events: %d)
	(Version: %d)
`
	return fmt.Sprintf(format, a.ID(), a.miles, a.tierPoints, a.status, reason, aggregateType, len(a.Events()), a.Version())
}
