package main

import "fmt"

// FrequentFlierAccount represents the state of an instance of the frequent flier
// account aggregate. It tracks changes on itself in the form of domain events.
type FrequentFlierAccount struct {
	aggregateRoot AggregateRoot
	miles         int
	tierPoints    int
	status        Status
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
func CreateFrequentFlierAccount(id string) *FrequentFlierAccount {
	self := &FrequentFlierAccount{}
	self.aggregateRoot.trackChange(*self, FrequentFlierAccountCreated{OpeningMiles: 0, OpeningTierPoints: 0}, self.transition)
	return self
}

// NewFrequentFlierAccountFromHistory creates a FrequentFlierAccount given a history
// of the changes which have occurred for that account.
func NewFrequentFlierAccountFromHistory(events []Event) *FrequentFlierAccount {
	state := &FrequentFlierAccount{}
	state.aggregateRoot.buildFromHistory(events, state.transition)
	return state
}

// RecordFlightTaken is used to record the fact that a customer has taken a flight
// which should be attached to this frequent flier account. The number of miles and
// tier points which apply are calculated externally.
//
// If recording this flight takes the account over a status boundary, it will
// automatically upgrade the account to the new status level.
func (self *FrequentFlierAccount) RecordFlightTaken(miles int, tierPoints int) {
	self.aggregateRoot.trackChange(*self, FlightTaken{MilesAdded: miles, TierPointsAdded: tierPoints}, self.transition)

	if self.tierPoints > 10 && self.status != StatusSilver {
		self.aggregateRoot.trackChange(*self, StatusMatched{NewStatus: StatusSilver}, self.transition)
	}

	if self.tierPoints > 20 && self.status != StatusGold {
		self.aggregateRoot.trackChange(*self, PromotedToGoldStatus{}, self.transition)
	}
}

// transition imnplements the pattern match against event types used both as part
// of the fold when loading from history and when tracking an individual change.
func (state *FrequentFlierAccount) transition(event Event) {

	switch e := event.data.(type) {

	case FrequentFlierAccountCreated:
		state.miles = e.OpeningMiles
		state.tierPoints = e.OpeningTierPoints
		state.status = StatusRed

	case StatusMatched:
		state.status = e.NewStatus

	case FlightTaken:
		state.miles = state.miles + e.MilesAdded
		state.tierPoints = state.tierPoints + e.TierPointsAdded

	case PromotedToGoldStatus:
		state.status = StatusGold
	}
}

// String implements Stringer for FrequentFlierAccount instances.
func (a FrequentFlierAccount) String() string {

	var reason string
	var aggregateType string

	if len(a.aggregateRoot.changes) > 0 {
		reason = a.aggregateRoot.changes[0].reason
		aggregateType = a.aggregateRoot.changes[0].aggregateType
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
	(Pending Changes: %d)
	(UnSaved Version: %d)
	(Saved Version: %d)

`
	return fmt.Sprintf(format, a.aggregateRoot.id, a.miles, a.tierPoints, a.status, reason, aggregateType, len(a.aggregateRoot.changes), a.aggregateRoot.currentVersion(), a.aggregateRoot.version)
}
