package main

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

func (e FrequentFlierAccountCreated) Apply(state *FrequentFlierAccount) {
    state.id = e.AccountId
    state.miles = e.OpeningMiles
    state.tierPoints = e.OpeningTierPoints
    state.status = StatusRed
}

type StatusMatched struct {
	NewStatus Status
}

func (e StatusMatched) Apply(state *FrequentFlierAccount) {
    state.status = e.NewStatus
}

type FlightTaken struct {
	MilesAdded      int
	TierPointsAdded int
}

func (e FlightTaken) Apply(state *FrequentFlierAccount) {
    state.miles = state.miles + e.MilesAdded
    state.tierPoints = state.tierPoints + e.TierPointsAdded
}

type PromotedToGoldStatus struct{}

func (e PromotedToGoldStatus) Apply(state *FrequentFlierAccount) {
    state.status = StatusGold
}
