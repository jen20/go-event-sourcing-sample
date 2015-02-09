package main

import "fmt"

//FrequentFlierAccount represents the state of an instance of the frequent flier
//account aggregate. It tracks changes on itself in the form of domain events.
type FrequentFlierAccount struct {
	id              string
	miles           int
	tierPoints      int
	status          Status
	expectedVersion int
	changes         []interface{}
}

//RecordFlightTaken is used to record the fact that a customer has taken a flight
//which should be attached to this frequent flier account. The number of miles and
//tier points which apply are calculated externally.
//
//If recording this flight takes the account over a status boundary, it will
//automatically upgrade the account to the new status level.
func (self *FrequentFlierAccount) RecordFlightTaken(miles int, tierPoints int) {
	self.trackChange(FlightTaken{MilesAdded: miles, TierPointsAdded: tierPoints})

	if self.tierPoints > 20 && self.status != StatusGold {
		self.trackChange(PromotedToGoldStatus{})
	}
}

//String implements Stringer for FrequentFlierAccount instances.
func (a FrequentFlierAccount) String() string {
	format := `FrequentFlierAccount: %s
	Miles: %d
	TierPoints: %d
	Status: %s
	(Expected Version: %d)
	(Pending Changes: %d)
`
	return fmt.Sprintf(format, a.id, a.miles, a.tierPoints, a.status, a.expectedVersion, len(a.changes))
}

//trackChange is used internally by bevhavious methods to apply a state change to
//the current instance and also track it in order that it can be persisted later.
func (state *FrequentFlierAccount) trackChange(event interface{}) {
	state.changes = append(state.changes, event)
	state.transition(event)
}

//transition imnplements the pattern match against event types used both as part
//of the fold when loading from history and when tracking an individual change.
func (state *FrequentFlierAccount) transition(event interface{}) {
	switch e := event.(type) {

	case FrequentFlierAccountCreated:
		state.id = e.AccountId
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

//NewFrequentFlierAccountFromHistory creates a FrequentFlierAccount given a history
//of the changes which have occurred for that account.
func NewFrequentFlierAccountFromHistory(events []interface{}) *FrequentFlierAccount {
	state := &FrequentFlierAccount{}
	for _, event := range events {
		state.transition(event)
		state.expectedVersion++
	}
	return state
}
