package main

import (
	"fmt"
)

func main() {
	history := []interface{}{
		FrequentFlierAccountCreated{AccountId: "1234567", OpeningMiles: 10000, OpeningTierPoints: 0},
		StatusMatched{NewStatus: StatusSilver},
		FlightTaken{MilesAdded: 2525, TierPointsAdded: 5},
		FlightTaken{MilesAdded: 2512, TierPointsAdded: 5},
		FlightTaken{MilesAdded: 5600, TierPointsAdded: 5},
		FlightTaken{MilesAdded: 3000, TierPointsAdded: 3},
	}

	aggregate := NewFrequentFlierAccountFromHistory(history)
	fmt.Println("Before RecordFlightTaken")
	fmt.Println(aggregate)

	aggregate.RecordFlightTaken(1000, 3)
	fmt.Println("After RecordFlightTaken")
	fmt.Println(aggregate)
}
